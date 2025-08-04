# Performance Quick Reference

## 🚀 Top Performers

| Operation | Performance | Throughput | Memory | Category |
|-----------|-------------|------------|---------|----------|
| **Hex Decoding** | 42.26 ns/op | 23.7M ops/sec | 32 B/op | Core Crypto |
| **Random Bytes (16)** | 53.54 ns/op | 18.7M ops/sec | 16 B/op | Core Crypto |
| **Hex Encoding** | 68.11 ns/op | 14.7M ops/sec | 128 B/op | Core Crypto |
| **SHA-256 Hashing** | 141.9 ns/op | 7.0M ops/sec | 128 B/op | Core Crypto |
| **Parallel Pseudonym** | 39.62 ns/op | 25.2M ops/sec | 144 B/op | IBE System |

## ⚡ High Performance (>1M ops/sec)

| Operation | Performance | Throughput | Memory |
|-----------|-------------|------------|---------|
| Random Bytes (32) | 81.39 ns/op | 12.3M ops/sec | 32 B/op |
| Random Bytes (64) | 135.6 ns/op | 7.4M ops/sec | 64 B/op |
| Random Bytes (128) | 237.1 ns/op | 4.2M ops/sec | 128 B/op |
| Basic Pseudonym | 193.5 ns/op | 5.5M ops/sec | 144 B/op |
| Generate Fingerprint | 172.7 ns/op | 5.8M ops/sec | 144 B/op |
| Decrypt Identity | 477.6 ns/op | 2.1M ops/sec | 1376 B/op |
| Correlation Keys | 362-409 ns/op | 2.4-2.8M ops/sec | 272 B/op |

## 🔐 Medium Performance (100K-1M ops/sec)

| Operation | Performance | Throughput | Memory |
|-----------|-------------|------------|---------|
| JWT Creation | 3.4 μs/op | 294K ops/sec | 3311 B/op |
| JWT Validation | 5.7 μs/op | 175K ops/sec | 3568 B/op |
| Encrypt Identity | 959.4 ns/op | 1.0M ops/sec | 1649 B/op |
| Enhanced Pseudonym | 431.5 ns/op | 2.3M ops/sec | 416 B/op |

## 🛡️ Security Operations (<100 ops/sec)

| Operation | Performance | Throughput | Memory | Notes |
|-----------|-------------|------------|---------|-------|
| Password Hashing | 43.4 ms/op | 23 ops/sec | 5268 B/op | Intentional |
| Password Verification | 43.6 ms/op | 23 ops/sec | 5244 B/op | Intentional |

## 📊 Memory Efficiency

### Most Memory Efficient
1. **Random Bytes (16)**: 16 B/op
2. **Hex Decoding**: 32 B/op
3. **Random Bytes (32)**: 32 B/op
4. **SHA-256**: 128 B/op
5. **Hex Encoding**: 128 B/op

### Memory Usage by Category
- **Core Crypto**: 16-128 B/op
- **IBE Operations**: 144-416 B/op
- **JWT Operations**: 1-3KB/op
- **Password Operations**: 5KB/op

## 🔄 Parallel Processing Gains

| Operation | Sequential | Parallel | Improvement |
|-----------|------------|----------|-------------|
| Pseudonym Generation | 193.5 ns/op | 39.62 ns/op | **5x faster** |

## 🎯 Performance Targets

### Optimization Opportunities
1. **JWT Validation**: 5.7μs → 2μs (3x improvement with caching)
2. **Role Key Validation**: 409.8ns → 100ns (4x improvement with caching)
3. **Pseudonym Generation**: 193.5ns → 39.62ns (5x improvement with parallel)

### Production Thresholds
- **Warning**: >1ms/op
- **Critical**: >10ms/op
- **Exception**: Password operations (intentionally slow)

## 🔍 Monitoring Points

### High-Frequency Operations
- **Pseudonym generation**: 5.5M ops/sec baseline
- **SHA-256 hashing**: 7.0M ops/sec baseline
- **Random bytes**: 18.7M ops/sec baseline

### Authentication Bottlenecks
- **JWT validation**: 175K ops/sec (potential caching target)
- **Password operations**: 23 ops/sec (security-constrained)

### Memory Watch Points
- **JWT operations**: >3KB/op (highest memory usage)
- **IBE operations**: 144-416 B/op (efficient)
- **Core crypto**: 16-128 B/op (very efficient)

## 🚨 Alert Thresholds

### Performance Alerts
```
Warning:  >1ms/op
Critical: >10ms/op
```

### Memory Alerts
```
Warning:  >1KB/op
Critical: >10KB/op
```

### Throughput Alerts
```
Warning:  <100K ops/sec (except passwords)
Critical: <10K ops/sec (except passwords)
```

## 📝 Quick Commands

### Run All Benchmarks
```bash
make benchmark
```

### Run Specific Suites
```bash
# IBE System
go test ./internal/ibe -bench=. -benchmem -run=^$ -count=1

# Core Crypto
go test ./internal/api/handlers -bench="Benchmark(Password|JWT|SHA256|RandomBytes|Hex)" -benchmem -run=^$ -count=1

# Handler Level
go test ./internal/api/handlers -bench=BenchmarkAuthHandler -benchmem -run=^$ -count=1
```

### Performance Analysis
```bash
# Compare with baseline
benchcmp baseline.txt current.txt

# Profile specific operation
go test -bench=BenchmarkOperation -cpuprofile=cpu.prof
``` 