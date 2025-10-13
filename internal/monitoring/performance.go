package monitoring

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// PerformanceConfig holds configuration for performance monitoring
type PerformanceConfig struct {
	// Monitoring settings
	Enabled            bool          `yaml:"enabled"`              // Enable performance monitoring
	SampleRate         float64       `yaml:"sample_rate"`          // Sample rate for monitoring (0.0-1.0)
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold"` // Threshold for slow queries

	// Metrics settings
	MetricsRetention time.Duration `yaml:"metrics_retention"` // How long to keep metrics
	MaxMetrics       int           `yaml:"max_metrics"`       // Maximum number of metrics to keep

	// Alerting settings
	AlertSlowQueries bool          `yaml:"alert_slow_queries"` // Alert on slow queries
	AlertThreshold   time.Duration `yaml:"alert_threshold"`    // Alert threshold
}

// DefaultPerformanceConfig returns default performance monitoring configuration
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		Enabled:            true,
		SampleRate:         1.0,                    // Monitor all queries
		SlowQueryThreshold: 100 * time.Millisecond, // 100ms
		MetricsRetention:   24 * time.Hour,         // 24 hours
		MaxMetrics:         10000,                  // 10000 metrics
		AlertSlowQueries:   true,
		AlertThreshold:     1 * time.Second, // 1 second
	}
}

// ProductionPerformanceConfig returns production-optimized performance monitoring configuration
func ProductionPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		Enabled:            true,
		SampleRate:         0.1,                   // Monitor 10% of queries
		SlowQueryThreshold: 50 * time.Millisecond, // 50ms
		MetricsRetention:   7 * 24 * time.Hour,    // 7 days
		MaxMetrics:         100000,                // 100000 metrics
		AlertSlowQueries:   true,
		AlertThreshold:     500 * time.Millisecond, // 500ms
	}
}

// QueryMetric represents a database query metric
type QueryMetric struct {
	Query        string        `json:"query"`
	Duration     time.Duration `json:"duration"`
	Timestamp    time.Time     `json:"timestamp"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	RowsAffected int           `json:"rows_affected,omitempty"`
}

// PerformanceStats represents performance statistics
type PerformanceStats struct {
	TotalQueries    int64         `json:"total_queries"`
	SlowQueries     int64         `json:"slow_queries"`
	AverageDuration time.Duration `json:"average_duration"`
	MaxDuration     time.Duration `json:"max_duration"`
	MinDuration     time.Duration `json:"min_duration"`
	ErrorCount      int64         `json:"error_count"`
	SuccessRate     float64       `json:"success_rate"`
}

// PerformanceMonitor provides performance monitoring functionality
type PerformanceMonitor struct {
	config    *PerformanceConfig
	logger    *slog.Logger
	metrics   []QueryMetric
	stats     PerformanceStats
	mutex     sync.RWMutex
	startTime time.Time
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *PerformanceConfig, logger *slog.Logger) *PerformanceMonitor {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	return &PerformanceMonitor{
		config:    config,
		logger:    logger,
		metrics:   make([]QueryMetric, 0),
		startTime: time.Now(),
	}
}

// RecordQuery records a database query metric
func (pm *PerformanceMonitor) RecordQuery(query string, duration time.Duration, success bool, err error, rowsAffected int) {
	if !pm.config.Enabled {
		return
	}

	// Apply sampling
	if pm.shouldSample() {
		metric := QueryMetric{
			Query:        query,
			Duration:     duration,
			Timestamp:    time.Now(),
			Success:      success,
			RowsAffected: rowsAffected,
		}

		if err != nil {
			metric.Error = err.Error()
		}

		pm.addMetric(metric)

		// Check for slow queries
		if duration > pm.config.SlowQueryThreshold {
			pm.logger.Warn("Slow query detected",
				"query", query,
				"duration", duration,
				"threshold", pm.config.SlowQueryThreshold,
			)
		}

		// Check for alerts
		if pm.config.AlertSlowQueries && duration > pm.config.AlertThreshold {
			pm.logger.Error("Query exceeded alert threshold",
				"query", query,
				"duration", duration,
				"threshold", pm.config.AlertThreshold,
			)
		}
	}
}

// shouldSample determines if a query should be sampled
func (pm *PerformanceMonitor) shouldSample() bool {
	// Simple sampling based on timestamp
	// In a real implementation, you might use a more sophisticated sampling method
	if pm.config.SampleRate >= 1.0 {
		return true
	}
	// Use a more random approach for better sampling distribution
	// Use microseconds for better randomness in tight loops
	return time.Now().UnixMicro()%1000 < int64(pm.config.SampleRate*1000)
}

// addMetric adds a metric to the collection
func (pm *PerformanceMonitor) addMetric(metric QueryMetric) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// Add metric
	pm.metrics = append(pm.metrics, metric)

	// Update stats
	pm.updateStats(metric)

	// Cleanup old metrics if needed
	if len(pm.metrics) > pm.config.MaxMetrics {
		pm.cleanupOldMetrics()
	}
}

// updateStats updates the performance statistics
func (pm *PerformanceMonitor) updateStats(metric QueryMetric) {
	pm.stats.TotalQueries++

	if metric.Duration > pm.config.SlowQueryThreshold {
		pm.stats.SlowQueries++
	}

	if !metric.Success {
		pm.stats.ErrorCount++
	}

	// Update duration stats
	if pm.stats.TotalQueries == 1 {
		pm.stats.AverageDuration = metric.Duration
		pm.stats.MaxDuration = metric.Duration
		pm.stats.MinDuration = metric.Duration
	} else {
		// Update average duration
		totalDuration := pm.stats.AverageDuration * time.Duration(pm.stats.TotalQueries-1)
		pm.stats.AverageDuration = (totalDuration + metric.Duration) / time.Duration(pm.stats.TotalQueries)

		// Update max duration
		if metric.Duration > pm.stats.MaxDuration {
			pm.stats.MaxDuration = metric.Duration
		}

		// Update min duration
		if metric.Duration < pm.stats.MinDuration {
			pm.stats.MinDuration = metric.Duration
		}
	}

	// Update success rate
	pm.stats.SuccessRate = float64(pm.stats.TotalQueries-pm.stats.ErrorCount) / float64(pm.stats.TotalQueries)
}

// cleanupOldMetrics removes old metrics based on retention policy
func (pm *PerformanceMonitor) cleanupOldMetrics() {
	cutoff := time.Now().Add(-pm.config.MetricsRetention)

	// Find the first metric that's not too old
	cutoffIndex := 0
	for i, metric := range pm.metrics {
		if metric.Timestamp.After(cutoff) {
			cutoffIndex = i
			break
		}
	}

	// Remove old metrics
	if cutoffIndex > 0 {
		pm.metrics = pm.metrics[cutoffIndex:]
	}

	// Also limit by max metrics count
	if len(pm.metrics) > pm.config.MaxMetrics {
		// Keep only the most recent metrics
		startIndex := len(pm.metrics) - pm.config.MaxMetrics
		pm.metrics = pm.metrics[startIndex:]
	}
}

// GetStats returns the current performance statistics
func (pm *PerformanceMonitor) GetStats() PerformanceStats {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	return pm.stats
}

// GetMetrics returns the current metrics (limited to recent ones)
func (pm *PerformanceMonitor) GetMetrics(limit int) []QueryMetric {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	if limit <= 0 || limit > len(pm.metrics) {
		limit = len(pm.metrics)
	}

	// Return the most recent metrics
	start := len(pm.metrics) - limit
	if start < 0 {
		start = 0
	}

	metrics := make([]QueryMetric, limit)
	copy(metrics, pm.metrics[start:])

	return metrics
}

// GetSlowQueries returns slow queries
func (pm *PerformanceMonitor) GetSlowQueries(limit int) []QueryMetric {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	slowQueries := make([]QueryMetric, 0)

	for _, metric := range pm.metrics {
		if metric.Duration > pm.config.SlowQueryThreshold {
			slowQueries = append(slowQueries, metric)
		}
	}

	if limit > 0 && limit < len(slowQueries) {
		slowQueries = slowQueries[len(slowQueries)-limit:]
	}

	return slowQueries
}

// GetErrorQueries returns queries that resulted in errors
func (pm *PerformanceMonitor) GetErrorQueries(limit int) []QueryMetric {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	errorQueries := make([]QueryMetric, 0)

	for _, metric := range pm.metrics {
		if !metric.Success {
			errorQueries = append(errorQueries, metric)
		}
	}

	if limit > 0 && limit < len(errorQueries) {
		errorQueries = errorQueries[len(errorQueries)-limit:]
	}

	return errorQueries
}

// Reset resets all metrics and statistics
func (pm *PerformanceMonitor) Reset() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	pm.metrics = make([]QueryMetric, 0)
	pm.stats = PerformanceStats{}
	pm.startTime = time.Now()

	pm.logger.Info("Performance monitor reset")
}

// LogStats logs the current performance statistics
func (pm *PerformanceMonitor) LogStats() {
	stats := pm.GetStats()

	pm.logger.Info("Performance statistics",
		"total_queries", stats.TotalQueries,
		"slow_queries", stats.SlowQueries,
		"average_duration", stats.AverageDuration,
		"max_duration", stats.MaxDuration,
		"min_duration", stats.MinDuration,
		"error_count", stats.ErrorCount,
		"success_rate", stats.SuccessRate,
	)
}

// DatabaseQueryWrapper wraps database queries with performance monitoring
type DatabaseQueryWrapper struct {
	monitor *PerformanceMonitor
}

// NewDatabaseQueryWrapper creates a new database query wrapper
func NewDatabaseQueryWrapper(monitor *PerformanceMonitor) *DatabaseQueryWrapper {
	return &DatabaseQueryWrapper{
		monitor: monitor,
	}
}

// WrapQuery wraps a database query with performance monitoring
func (w *DatabaseQueryWrapper) WrapQuery(ctx context.Context, query string, fn func() error) error {
	start := time.Now()

	err := fn()

	duration := time.Since(start)
	success := err == nil

	w.monitor.RecordQuery(query, duration, success, err, 0)

	return err
}

// WrapQueryWithResult wraps a database query with performance monitoring and result counting
func (w *DatabaseQueryWrapper) WrapQueryWithResult(ctx context.Context, query string, fn func() (int, error)) (int, error) {
	start := time.Now()

	rowsAffected, err := fn()

	duration := time.Since(start)
	success := err == nil

	w.monitor.RecordQuery(query, duration, success, err, rowsAffected)

	return rowsAffected, err
}
