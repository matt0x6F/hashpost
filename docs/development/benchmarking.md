# HashPost Crypto Operation Benchmarks

This document describes the benchmarking setup for HashPost's crypto operations, including IBE (Identity-Based Encryption) operations, authentication handlers, and correlation handlers.

## Overview

The benchmarking system is designed to measure the performance of various cryptographic operations used throughout the HashPost application. This helps identify performance bottlenecks and optimize critical crypto operations.

## Benchmark Categories

### 1. IBE (Identity-Based Encryption) Operations

Located in `internal/ibe/ibe_benchmark_test.go`, these benchmarks measure:

- **Pseudonym Generation**: Creating pseudonymous identities
- **Correlation Key Generation**: Generating time-bounded correlation keys
- **Identity Encryption/Decryption**: Encrypting and decrypting real identities
- **Fingerprint Generation**: Creating cryptographic fingerprints
- **Role Key Operations**: Generating and validating role keys

### 2. Authentication Handler Operations

Located in `internal/api/handlers/auth_benchmark_test.go`, these benchmarks measure:

- **Password Hashing**: bcrypt operations with different costs
- **JWT Operations**: Token creation, validation, and parsing
- **SHA256 Hashing**: Cryptographic hashing operations
- **Random Number Generation**: Cryptographically secure random bytes
- **Hex Encoding/Decoding**: Data encoding operations

### 3. Correlation Handler Operations

Located in `internal/api/handlers/correlation_benchmark_test.go`, these benchmarks measure:

- **Identity Correlation**: Correlating pseudonyms with real identities
- **Fingerprint Correlation**: Correlating fingerprints with pseudonyms
- **Domain-Specific Operations**: Operations across different IBE domains
- **Version-Specific Operations**: Multi-version key operations

## Running Benchmarks

### Quick Start

```bash
# Run all benchmarks
make benchmark

# Run specific benchmark categories
make benchmark-ibe
make benchmark-auth
make benchmark-correlation

# Run all benchmarks with memory profiling
make benchmark-all
```

### Individual Benchmark Files

```bash
# Run IBE benchmarks only
go test ./internal/ibe -bench=. -benchmem -run=^$

# Run auth handler benchmarks only
go test ./internal/api/handlers -bench=Benchmark.*Auth -benchmem -run=^$

# Run correlation handler benchmarks only
go test ./internal/api/handlers -bench=Benchmark.*Correlation -benchmem -run=^$
```

### Benchmark Options

```bash
# Run with specific iterations
go test ./internal/ibe -bench=. -benchtime=10s -benchmem

# Run with CPU profiling
go test ./internal/ibe -bench=. -cpuprofile=cpu.prof

# Run with memory profiling
go test ./internal/ibe -bench=. -memprofile=mem.prof

# Run specific benchmark
go test ./internal/ibe -bench=BenchmarkIBESystem_GeneratePseudonym -benchmem
```

## Benchmark Results Interpretation

### Key Metrics

- **ns/op**: Nanoseconds per operation
- **B/op**: Bytes allocated per operation
- **allocs/op**: Number of allocations per operation

### Performance Guidelines

#### IBE Operations
- **Pseudonym Generation**: Should complete in < 1ms
- **Correlation Key Generation**: Should complete in < 100μs
- **Identity Encryption**: Should complete in < 5ms
- **Identity Decryption**: Should complete in < 5ms

#### Authentication Operations
- **Password Hashing**: bcrypt cost 12 should complete in < 100ms
- **JWT Creation**: Should complete in < 1ms
- **JWT Validation**: Should complete in < 1ms

#### Correlation Operations
- **Fingerprint Generation**: Should complete in < 1ms
- **Correlation Key Generation**: Should complete in < 100μs

## Benchmark Categories

### 1. Basic Operations
- Simple function calls with minimal setup
- Measures raw performance of crypto primitives

### 2. Parallel Operations
- Tests concurrent access patterns
- Identifies thread safety and contention issues

### 3. Different Input Sizes
- Tests performance with varying data sizes
- Helps identify scaling characteristics

### 4. Error Conditions
- Tests performance of error handling paths
- Ensures error cases don't have performance issues

## Benchmark Best Practices

### 1. Setup and Teardown
- Use `b.ResetTimer()` to exclude setup time
- Keep setup minimal and focused

### 2. Realistic Data
- Use realistic input data sizes
- Test with actual production-like data

### 3. Memory Allocation
- Monitor memory allocations with `-benchmem`
- Look for excessive allocations

### 4. CPU Usage
- Use CPU profiling to identify bottlenecks
- Focus optimization efforts on hot paths

## Performance Optimization Tips

### IBE Operations
1. **Key Caching**: Cache frequently used domain keys
2. **Batch Operations**: Process multiple operations together
3. **Async Processing**: Use goroutines for parallel operations

### Authentication Operations
1. **JWT Caching**: Cache parsed JWT tokens
2. **Password Hashing**: Use appropriate bcrypt cost
3. **Token Validation**: Optimize validation logic

### Correlation Operations
1. **Key Pre-computation**: Pre-compute correlation keys
2. **Domain Separation**: Use appropriate domains for operations
3. **Version Management**: Efficiently handle key versions

## Monitoring and Alerting

### Performance Thresholds
- Set up alerts for benchmark regressions
- Monitor performance trends over time
- Track memory usage patterns

### Continuous Integration
- Run benchmarks in CI pipeline
- Fail builds on significant regressions
- Track performance metrics over time

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Check for memory leaks in benchmark setup
   - Use `-memprofile` to identify allocations

2. **Slow Performance**
   - Profile with `-cpuprofile`
   - Check for unnecessary allocations
   - Verify crypto library versions

3. **Inconsistent Results**
   - Run benchmarks multiple times
   - Check for system load
   - Use `-count` flag for multiple runs

### Debugging Commands

```bash
# Run with verbose output
go test ./internal/ibe -bench=. -v

# Run specific benchmark with count
go test ./internal/ibe -bench=BenchmarkIBESystem_GeneratePseudonym -count=5

# Run with race detection
go test ./internal/ibe -bench=. -race

# Run with coverage
go test ./internal/ibe -bench=. -cover
```

## Future Enhancements

### Planned Improvements
1. **Automated Performance Regression Testing**
2. **Performance Dashboard Integration**
3. **Real-world Load Testing**
4. **Memory Usage Optimization**

### New Benchmark Categories
1. **Database Crypto Operations**
2. **Network Crypto Operations**
3. **Concurrent Access Patterns**
4. **Memory Pressure Testing**

## Contributing

When adding new benchmarks:

1. **Follow Naming Convention**: `BenchmarkCategory_Operation`
2. **Include Memory Profiling**: Use `-benchmem` flag
3. **Add Documentation**: Document the benchmark purpose
4. **Test Realistic Scenarios**: Use production-like data
5. **Include Error Cases**: Test error handling performance

### Example Benchmark Structure

```go
func BenchmarkCategory_Operation(b *testing.B) {
    // Setup
    setup := createBenchmarkSetup()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Operation being benchmarked
        result, err := setup.Operation()
        if err != nil {
            b.Fatal(err)
        }
        _ = result
    }
}
```

## References

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Go Benchmarking Guide](https://golang.org/pkg/testing/#hdr-Benchmarks)
- [Performance Profiling](https://golang.org/pkg/runtime/pprof/)
- [Memory Profiling](https://golang.org/pkg/runtime/pprof/#WriteHeapProfile) 