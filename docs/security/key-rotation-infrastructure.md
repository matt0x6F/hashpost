# HashPost Key Rotation Infrastructure

## Executive Summary

This document describes the automated key rotation infrastructure for HashPost's domain-separated Identity-Based Encryption (IBE) system. The solution addresses the security assessment's recommendation for "automated key rotation scheduling" while considering distributed system constraints and operational overhead.

## Architecture Overview

### Current State
HashPost implements domain-separated IBE with five cryptographic domains:
- `DOMAIN_USER_PSEUDONYMS` - User pseudonym generation
- `DOMAIN_USER_CORRELATION` - User self-correlation
- `DOMAIN_MOD_CORRELATION` - Moderator correlation
- `DOMAIN_ADMIN_CORRELATION` - Admin correlation
- `DOMAIN_LEGAL_CORRELATION` - Legal compliance

### Rotation Requirements
Each domain requires different rotation strategies based on data sensitivity and operational impact.

## Distributed System Constraints

### Single Point of Rotation
In a distributed HashPost deployment, key rotation must be coordinated to prevent:
- **Race conditions** between multiple servers attempting rotation
- **Inconsistent key states** across the cluster
- **Data corruption** from concurrent migration operations

### Solution: Leader-Based Rotation
```go
type RotationCoordinator struct {
    nodeID        string
    isLeader      bool
    leaderElection *LeaderElection
    rotationLock  *DistributedLock
    ibeSystem     *ibe.IBESystem
    logger        zerolog.Logger
}

type LeaderElection struct {
    etcdClient    *clientv3.Client
    electionKey   string
    leaseID       clientv3.LeaseID
    isLeader      bool
}
```

### Leader Election Implementation
```go
func (rc *RotationCoordinator) StartLeaderElection() error {
    // Use etcd for leader election
    lease, err := rc.etcdClient.Grant(context.Background(), 30) // 30 second lease
    if err != nil {
        return fmt.Errorf("failed to create lease: %w", err)
    }
    
    // Try to become leader
    txn := rc.etcdClient.Txn(context.Background())
    txn.If(clientv3.Compare(clientv3.CreateRevision(rc.electionKey), "=", 0)).
        Then(clientv3.OpPut(rc.electionKey, rc.nodeID, clientv3.WithLease(lease.ID))).
        Else(clientv3.OpGet(rc.electionKey))
    
    resp, err := txn.Commit()
    if err != nil {
        return fmt.Errorf("leader election failed: %w", err)
    }
    
    if resp.Succeeded {
        rc.isLeader = true
        rc.leaseID = lease.ID
        rc.logger.Info().Msg("Became rotation coordinator leader")
        
        // Start rotation scheduler
        go rc.startRotationScheduler()
        
        // Keep lease alive
        go rc.keepLeaseAlive()
    } else {
        rc.logger.Info().Msg("Another node is rotation coordinator leader")
    }
    
    return nil
}

func (rc *RotationCoordinator) keepLeaseAlive() {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        if rc.isLeader {
            _, err := rc.etcdClient.KeepAliveOnce(context.Background(), rc.leaseID)
            if err != nil {
                rc.logger.Error().Err(err).Msg("Failed to keep lease alive")
                rc.isLeader = false
                break
            }
        }
    }
}
```

## Domain-Specific Rotation Strategies

### 1. User Pseudonyms Domain (Low Overhead)

**Strategy**: Regeneration (no data migration)
**Rotation Schedule**: Annual
**Overhead**: Minimal

```go
type PseudonymRotation struct {
    ibeSystem *ibe.IBESystem
    userDAO   *dao.UserDAO
    logger    zerolog.Logger
}

func (pr *PseudonymRotation) Rotate() error {
    // 1. Generate new domain master key
    newDomainKey := pr.generateNewDomainKey(DOMAIN_USER_PSEUDONYMS)
    
    // 2. Update IBE system with new key
    pr.ibeSystem.UpdateDomainMaster(DOMAIN_USER_PSEUDONYMS, newDomainKey)
    
    // 3. Regenerate all pseudonyms (can be done in batches)
    batchSize := 1000
    offset := 0
    
    for {
        users, err := pr.userDAO.GetUsersBatch(offset, batchSize)
        if err != nil {
            return fmt.Errorf("failed to get users batch: %w", err)
        }
        
        if len(users) == 0 {
            break
        }
        
        for _, user := range users {
            // Regenerate pseudonym with new domain key
            newPseudonym := pr.ibeSystem.GeneratePseudonym(user.UserID, "default", 1)
            
            // Update user record
            err := pr.userDAO.UpdatePseudonym(user.UserID, newPseudonym)
            if err != nil {
                pr.logger.Error().Err(err).Int64("user_id", user.UserID).Msg("Failed to update pseudonym")
                continue
            }
        }
        
        offset += batchSize
        pr.logger.Info().Int("processed", offset).Msg("Pseudonym rotation progress")
    }
    
    return nil
}
```

**Overhead Analysis**:
- **CPU**: Low - Simple hash operations
- **Database**: Medium - Batch updates to user table
- **Memory**: Low - No large data structures
- **Time**: ~5-10 minutes for 1M users
- **Downtime**: None - Pseudonyms can be regenerated on-demand

### 2. User Self-Correlation Domain (Medium Overhead)

**Strategy**: Re-encryption with grace period
**Rotation Schedule**: Quarterly
**Overhead**: Moderate

```go
type SelfCorrelationRotation struct {
    ibeSystem *ibe.IBESystem
    mappingDAO *dao.IdentityMappingDAO
    logger     zerolog.Logger
}

func (scr *SelfCorrelationRotation) Rotate() error {
    // 1. Generate new domain master key
    newDomainKey := scr.generateNewDomainKey(DOMAIN_USER_CORRELATION)
    
    // 2. Create new mappings alongside existing ones
    batchSize := 500
    offset := 0
    
    for {
        mappings, err := scr.mappingDAO.GetMappingsByScope("self_correlation", offset, batchSize)
        if err != nil {
            return fmt.Errorf("failed to get mappings: %w", err)
        }
        
        if len(mappings) == 0 {
            break
        }
        
        for _, mapping := range mappings {
            // Decrypt with old key
            decrypted, err := scr.decryptWithOldKey(mapping.EncryptedRealIdentity)
            if err != nil {
                scr.logger.Error().Err(err).Msg("Failed to decrypt mapping")
                continue
            }
            
            // Re-encrypt with new key
            reEncrypted, err := scr.encryptWithNewKey(decrypted, newDomainKey)
            if err != nil {
                scr.logger.Error().Err(err).Msg("Failed to re-encrypt mapping")
                continue
            }
            
            // Create new mapping version
            err = scr.mappingDAO.CreateMappingVersion(mapping.MappingID, DOMAIN_USER_CORRELATION, reEncrypted)
            if err != nil {
                scr.logger.Error().Err(err).Msg("Failed to create mapping version")
                continue
            }
        }
        
        offset += batchSize
        scr.logger.Info().Int("processed", offset).Msg("Self-correlation rotation progress")
    }
    
    return nil
}
```

**Overhead Analysis**:
- **CPU**: Medium - AES encryption/decryption operations
- **Database**: High - Read/write operations on identity mappings
- **Memory**: Medium - Batch processing of encrypted data
- **Time**: ~30-60 minutes for 1M mappings
- **Downtime**: None - Grace period maintains access

### 3. Moderator Correlation Domain (High Overhead)

**Strategy**: Dual-key access during transition
**Rotation Schedule**: Quarterly
**Overhead**: High

```go
type ModeratorCorrelationRotation struct {
    ibeSystem *ibe.IBESystem
    mappingDAO *dao.IdentityMappingDAO
    logger     zerolog.Logger
}

func (mcr *ModeratorCorrelationRotation) Rotate() error {
    // 1. Generate new domain master key
    newDomainKey := mcr.generateNewDomainKey(DOMAIN_MOD_CORRELATION)
    
    // 2. Enable dual-key access mode
    mcr.ibeSystem.EnableDualKeyMode(DOMAIN_MOD_CORRELATION, newDomainKey)
    
    // 3. Migrate existing mappings in background
    go mcr.migrateMappingsInBackground(DOMAIN_MOD_CORRELATION, newDomainKey)
    
    return nil
}

func (mcr *ModeratorCorrelationRotation) migrateMappingsInBackground(domain string, newKey []byte) {
    batchSize := 100 // Smaller batches for correlation data
    offset := 0
    
    for {
        mappings, err := mcr.mappingDAO.GetMappingsByScope("correlation", offset, batchSize)
        if err != nil {
            mcr.logger.Error().Err(err).Msg("Failed to get correlation mappings")
            time.Sleep(5 * time.Second)
            continue
        }
        
        if len(mappings) == 0 {
            break
        }
        
        for _, mapping := range mappings {
            // Create new mapping version with new key
            err := mcr.createNewMappingVersion(mapping, newKey)
            if err != nil {
                mcr.logger.Error().Err(err).Msg("Failed to create mapping version")
                continue
            }
        }
        
        offset += batchSize
        mcr.logger.Info().Int("processed", offset).Msg("Moderator correlation migration progress")
        
        // Rate limiting to avoid overwhelming the system
        time.Sleep(100 * time.Millisecond)
    }
}
```

**Overhead Analysis**:
- **CPU**: High - Complex encryption operations
- **Database**: Very High - Heavy read/write operations
- **Memory**: High - Large encrypted data structures
- **Time**: ~2-4 hours for 1M mappings
- **Downtime**: None - Dual-key access maintains functionality

### 4. Admin/Legal Correlation Domains (Very High Overhead)

**Strategy**: Dual-key access with extended grace period
**Rotation Schedule**: Annual/Biennial
**Overhead**: Very High

```go
type AdminCorrelationRotation struct {
    ibeSystem *ibe.IBESystem
    mappingDAO *dao.IdentityMappingDAO
    auditDAO   *dao.CorrelationAuditDAO
    logger     zerolog.Logger
}

func (acr *AdminCorrelationRotation) Rotate() error {
    // 1. Generate new domain master key
    newDomainKey := acr.generateNewDomainKey(DOMAIN_ADMIN_CORRELATION)
    
    // 2. Enable dual-key access with extended grace period
    acr.ibeSystem.EnableDualKeyMode(DOMAIN_ADMIN_CORRELATION, newDomainKey)
    
    // 3. Schedule background migration with low priority
    go acr.migrateWithLowPriority(DOMAIN_ADMIN_CORRELATION, newDomainKey)
    
    return nil
}

func (acr *AdminCorrelationRotation) migrateWithLowPriority(domain string, newKey []byte) {
    batchSize := 50 // Very small batches
    offset := 0
    
    for {
        mappings, err := acr.mappingDAO.GetMappingsByScope("correlation", offset, batchSize)
        if err != nil {
            acr.logger.Error().Err(err).Msg("Failed to get admin correlation mappings")
            time.Sleep(10 * time.Second)
            continue
        }
        
        if len(mappings) == 0 {
            break
        }
        
        for _, mapping := range mappings {
            err := acr.createNewMappingVersion(mapping, newKey)
            if err != nil {
                acr.logger.Error().Err(err).Msg("Failed to create admin mapping version")
                continue
            }
        }
        
        offset += batchSize
        acr.logger.Info().Int("processed", offset).Msg("Admin correlation migration progress")
        
        // Very slow rate limiting for admin data
        time.Sleep(1 * time.Second)
    }
}
```

**Overhead Analysis**:
- **CPU**: Very High - Complex operations on sensitive data
- **Database**: Very High - Heavy operations with audit logging
- **Memory**: Very High - Large encrypted data structures
- **Time**: ~8-24 hours for 1M mappings
- **Downtime**: None - Extended grace period

## Distributed Coordination

### Rotation State Management

```go
type RotationState struct {
    Domain          string    `json:"domain"`
    Status          string    `json:"status"` // pending, in_progress, completed, failed
    StartTime       time.Time `json:"start_time"`
    EndTime         time.Time `json:"end_time,omitempty"`
    Progress        int       `json:"progress"` // 0-100
    OldKeyVersion   int       `json:"old_key_version"`
    NewKeyVersion   int       `json:"new_key_version"`
    LeaderNode      string    `json:"leader_node"`
    ErrorMessage    string    `json:"error_message,omitempty"`
}

type RotationCoordinator struct {
    etcdClient *clientv3.Client
    stateKey   string
    logger     zerolog.Logger
}

func (rc *RotationCoordinator) UpdateRotationState(state RotationState) error {
    stateJSON, err := json.Marshal(state)
    if err != nil {
        return fmt.Errorf("failed to marshal state: %w", err)
    }
    
    _, err = rc.etcdClient.Put(context.Background(), rc.stateKey, string(stateJSON))
    return err
}

func (rc *RotationCoordinator) GetRotationState() (*RotationState, error) {
    resp, err := rc.etcdClient.Get(context.Background(), rc.stateKey)
    if err != nil {
        return nil, fmt.Errorf("failed to get state: %w", err)
    }
    
    if len(resp.Kvs) == 0 {
        return nil, nil
    }
    
    var state RotationState
    err = json.Unmarshal(resp.Kvs[0].Value, &state)
    if err != nil {
        return nil, fmt.Errorf("failed to unmarshal state: %w", err)
    }
    
    return &state, nil
}
```

### Health Monitoring

```go
type RotationHealthMonitor struct {
    coordinator *RotationCoordinator
    metrics     *prometheus.Registry
    logger      zerolog.Logger
}

func (rhm *RotationHealthMonitor) MonitorRotation() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        state, err := rhm.coordinator.GetRotationState()
        if err != nil {
            rhm.logger.Error().Err(err).Msg("Failed to get rotation state")
            continue
        }
        
        if state != nil {
            // Update metrics
            rotationProgress.WithLabelValues(state.Domain).Set(float64(state.Progress))
            rotationStatus.WithLabelValues(state.Domain, state.Status).Set(1)
            
            // Check for stuck rotations
            if state.Status == "in_progress" && time.Since(state.StartTime) > 24*time.Hour {
                rhm.logger.Warn().
                    Str("domain", state.Domain).
                    Dur("duration", time.Since(state.StartTime)).
                    Msg("Rotation appears to be stuck")
            }
        }
    }
}
```

## Performance Impact Analysis

### Resource Usage Summary

| Domain | CPU Impact | Memory Impact | Database Impact | Time to Complete | Grace Period |
|--------|------------|---------------|-----------------|------------------|--------------|
| User Pseudonyms | Low | Low | Medium | 5-10 min | 30 days |
| User Self-Correlation | Medium | Medium | High | 30-60 min | 7 days |
| Moderator Correlation | High | High | Very High | 2-4 hours | 14 days |
| Admin Correlation | Very High | Very High | Very High | 8-24 hours | 30 days |
| Legal Correlation | Very High | Very High | Very High | 8-24 hours | 90 days |

### Operational Considerations

#### 1. **Scheduling Strategy**
```go
// Stagger rotations to avoid resource contention
var RotationSchedule = map[string]time.Weekday{
    DOMAIN_USER_PSEUDONYMS:   time.Sunday,    // Low impact
    DOMAIN_USER_CORRELATION:  time.Monday,    // Medium impact
    DOMAIN_MOD_CORRELATION:   time.Tuesday,   // High impact
    DOMAIN_ADMIN_CORRELATION: time.Wednesday, // Very high impact
    DOMAIN_LEGAL_CORRELATION: time.Thursday,  // Very high impact
}
```

#### 2. **Resource Throttling**
```go
type ResourceThrottler struct {
    maxCPUPercent    float64
    maxMemoryPercent float64
    maxDBConnections int
}

func (rt *ResourceThrottler) ShouldThrottle() bool {
    cpuUsage := rt.getCPUUsage()
    memoryUsage := rt.getMemoryUsage()
    dbConnections := rt.getDBConnections()
    
    return cpuUsage > rt.maxCPUPercent ||
           memoryUsage > rt.maxMemoryPercent ||
           dbConnections > rt.maxDBConnections
}
```

#### 3. **Rollback Capability**
```go
func (rc *RotationCoordinator) RollbackRotation(domain string) error {
    // 1. Stop migration process
    rc.stopMigration(domain)
    
    // 2. Revert to old domain key
    rc.ibeSystem.RevertDomainMaster(domain)
    
    // 3. Clean up partial migration data
    rc.cleanupPartialMigration(domain)
    
    // 4. Update rotation state
    state := &RotationState{
        Domain:    domain,
        Status:    "rolled_back",
        EndTime:   time.Now(),
        Progress:  0,
    }
    rc.UpdateRotationState(*state)
    
    return nil
}
```

## Monitoring and Alerting

### Key Metrics
```go
var (
    rotationProgress = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "ibe_rotation_progress",
            Help: "Progress of IBE key rotation by domain",
        },
        []string{"domain"},
    )
    
    rotationDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "ibe_rotation_duration_seconds",
            Help:    "Duration of IBE key rotation by domain",
            Buckets: prometheus.ExponentialBuckets(60, 2, 10), // 1min to 17hours
        },
        []string{"domain"},
    )
    
    rotationErrors = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "ibe_rotation_errors_total",
            Help: "Total number of IBE rotation errors by domain",
        },
        []string{"domain", "error_type"},
    )
)
```

### Alerting Rules
```yaml
# Prometheus alerting rules
groups:
  - name: ibe_rotation_alerts
    rules:
      - alert: RotationStuck
        expr: ibe_rotation_progress > 0 and time() - ibe_rotation_start_time > 3600
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "IBE rotation appears to be stuck"
          
      - alert: RotationFailed
        expr: ibe_rotation_errors_total > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "IBE rotation has failed"
```

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
- [ ] Implement leader election mechanism
- [ ] Create rotation state management
- [ ] Add basic monitoring metrics
- [ ] Implement user pseudonym rotation (lowest risk)

### Phase 2: Core Rotation (Week 3-4)
- [ ] Implement user self-correlation rotation
- [ ] Add resource throttling
- [ ] Create rollback mechanisms
- [ ] Add comprehensive logging

### Phase 3: Advanced Rotation (Week 5-6)
- [ ] Implement moderator correlation rotation
- [ ] Add dual-key access mechanisms
- [ ] Create background migration processes
- [ ] Add performance monitoring

### Phase 4: Production Ready (Week 7-8)
- [ ] Implement admin/legal correlation rotation
- [ ] Add comprehensive alerting
- [ ] Create operational runbooks
- [ ] Performance testing and optimization

## Conclusion

The automated key rotation infrastructure provides comprehensive security while addressing distributed system constraints:

1. **Leader-based coordination** prevents race conditions
2. **Domain-specific strategies** optimize for each use case
3. **Graceful degradation** maintains system availability
4. **Comprehensive monitoring** ensures operational visibility
5. **Rollback capabilities** provide safety nets

The overhead is manageable with proper scheduling and resource management, and the distributed nature of HashPost can be accommodated through coordinated rotation led by a single node. 