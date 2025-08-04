# Benchmark Implementation Guide

## Overview

This guide explains how to implement, run, and interpret cryptographic benchmarks for HashPost. The benchmarks use production-ready configurations with realistic key sizes and comprehensive mocking.

## Benchmark Architecture

### Key Components

1. **IBE System Benchmarks** (`internal/ibe/ibe_benchmark_test.go`)
   - Pseudonym generation and validation
   - Correlation key operations
   - Identity encryption/decryption
   - Fingerprint generation

2. **Authentication Benchmarks** (`internal/api/handlers/auth_benchmark_test.go`)
   - Core crypto operations (SHA-256, JWT, random bytes)
   - Password hashing and verification
   - Hex encoding/decoding

3. **Handler-Level Benchmarks** (`internal/api/handlers/auth_handler_benchmark_test.go`)
   - Complete authentication flows
   - Registration and login processes
   - Token management operations

4. **Correlation Benchmarks** (`internal/api/handlers/correlation_benchmark_test.go`)
   - IBE correlation operations
   - Identity mapping operations
   - Fingerprint correlation

## Implementation Methodology

### 1. Production-Ready Configuration

#### IBE System Setup
```go
// Generate realistic 32-byte domain master keys (AES-256 equivalent)
domainMasters := make(map[string][]byte)
domains := []string{
    DOMAIN_USER_PSEUDONYMS,
    DOMAIN_USER_CORRELATION,
    DOMAIN_MOD_CORRELATION,
    DOMAIN_ADMIN_CORRELATION,
    DOMAIN_LEGAL_CORRELATION,
}
for _, domain := range domains {
    key := make([]byte, 32) // 32 bytes for 256-bit key
    if _, err := rand.Read(key); err != nil {
        panic(fmt.Sprintf("failed to generate random key for domain %s: %v", domain, err))
    }
    domainMasters[domain] = key
}

return NewIBESystemWithOptions(IBEOptions{
    DomainMasters: domainMasters,
    KeyVersion:    1,
    Salt:          "benchmark_fingerprint_salt_v1",
})
```

#### JWT Configuration
```go
// Use realistic 256-bit JWT secret (32 bytes)
secret := "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long"
```

### 2. Comprehensive Mocking Strategy

#### Mock DAO Setup
```go
// Create mock DAOs with realistic behavior
mockUserDAO := &mocks.MockUserDAO{}
mockPseudonymDAO := &mocks.MockPseudonymDAO{}
mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}

// Set up realistic mock responses
mockUserDAO.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(&dbmodels.User{
    UserID: 1,
    Email:  "test@example.com",
}, nil)

// Mock that no user exists with this email (for registration)
mockUserDAO.On("GetUserByEmail", mock.Anything, mock.Anything).Return(nil, nil)
```

#### Configuration Mocking
```go
// Create config with production JWT secret and disabled email validation
cfg := &config.Config{
    JWT: config.JWTConfig{
        Secret:     "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long",
        Expiration: 24 * time.Hour,
    },
    Email: config.EmailConfig{
        Provider: "", // Disable email service for benchmarking
        Validation: config.EmailValidationConfig{
            Enabled: false, // Disable email validation for benchmarking
        },
    },
}
```

### 3. Benchmark Structure

#### Basic Benchmark Pattern
```go
func BenchmarkOperationName(b *testing.B) {
    // Setup
    handler := createBenchmarkHandler()
    
    // Input preparation
    input := &models.InputType{
        Body: models.InputBody{
            // Test data
        },
    }
    
    // Reset timer to exclude setup
    b.ResetTimer()
    
    // Benchmark loop
    for i := 0; i < b.N; i++ {
        _, err := handler.Operation(context.Background(), input)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### Sub-benchmark Pattern
```go
func BenchmarkOperation_DifferentInputs(b *testing.B) {
    inputs := []string{"short", "medium", "long"}
    
    for _, input := range inputs {
        b.Run(fmt.Sprintf("%d_chars", len(input)), func(b *testing.B) {
            // Benchmark implementation
        })
    }
}
```

## Running Benchmarks

### 1. Individual Benchmark Suites

#### IBE System Benchmarks
```bash
go test ./internal/ibe -bench=. -benchmem -run=^$ -count=1
```

#### Authentication Benchmarks
```bash
go test ./internal/api/handlers -bench="Benchmark(Password|JWT|SHA256|RandomBytes|Hex)" -benchmem -run=^$ -count=1
```

#### Handler-Level Benchmarks
```bash
go test ./internal/api/handlers -bench=BenchmarkAuthHandler -benchmem -run=^$ -count=1
```

#### Correlation Benchmarks
```bash
go test ./internal/api/handlers -bench=BenchmarkIBE -benchmem -run=^$ -count=1
```

### 2. Complete Benchmark Suite
```bash
make benchmark
```

### 3. Benchmark Options

#### Standard Options
- `-bench=.`: Run all benchmarks
- `-benchmem`: Include memory allocation statistics
- `-run=^$`: Skip unit tests, run only benchmarks
- `-count=1`: Run each benchmark once (faster)

#### Advanced Options
- `-benchtime=10s`: Set benchmark duration
- `-cpu=1,2,4,8`: Test with different CPU counts
- `-parallel=4`: Set parallel benchmark count

## Interpreting Results

### 1. Performance Metrics

#### Time per Operation
```
BenchmarkOperation-16    1000000    1234 ns/op
```
- **1000000**: Number of operations run
- **1234 ns/op**: Average time per operation
- **-16**: Number of CPU cores used

#### Memory Usage
```
BenchmarkOperation-16    1000000    1234 ns/op    1234 B/op    12 allocs/op
```
- **1234 B/op**: Bytes allocated per operation
- **12 allocs/op**: Number of allocations per operation

### 2. Throughput Calculation

#### Operations per Second
```
Throughput = 1,000,000,000 / (ns_per_operation)
```

Example:
- **1234 ns/op** = 810,000 ops/sec
- **100 ns/op** = 10,000,000 ops/sec

## Benchmark Best Practices

### 1. Setup and Teardown

#### Proper Timer Reset
```go
func BenchmarkOperation(b *testing.B) {
    // Expensive setup
    handler := createExpensiveHandler()
    
    // Reset timer after setup
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

#### Parallel Benchmarks
```go
func BenchmarkOperation_Parallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            // Thread-safe benchmark code
        }
    })
}
```

### 2. Realistic Data

#### Production Key Sizes
```go
// Use 256-bit keys for production realism
key := make([]byte, 32) // 32 bytes = 256 bits
rand.Read(key)
```

#### Realistic Input Sizes
```go
// Test with various input sizes
inputs := []string{
    "short",                                    // 5 chars
    "medium_length_input",                      // 22 chars
    "very_long_input_with_many_characters",    // 56 chars
}
```

### 3. Error Handling

#### Proper Error Checking
```go
for i := 0; i < b.N; i++ {
    result, err := operation(input)
    if err != nil {
        b.Fatal(err) // Stop benchmark on error
    }
    // Use result to prevent optimization
    _ = result
}
```

## Troubleshooting

### 1. Common Issues

#### Missing Mocks
```
Error: mock: Unexpected Method Call
```
**Solution**: Add missing mock expectations in setup functions.

#### Configuration Errors
```
Error: email validation failed
```
**Solution**: Disable email validation in benchmark config.

#### Memory Leaks
```
Error: too many allocations
```
**Solution**: Use object pooling or reduce allocations in hot paths.

### 2. Performance Debugging

#### High Memory Usage
- Check for unnecessary allocations
- Consider object pooling
- Review string concatenation patterns

#### Slow Operations
- Profile with `go test -cpuprofile`
- Check for blocking operations
- Review algorithm complexity

### 3. Benchmark Stability

#### Inconsistent Results
- Run with `-count=5` for multiple iterations
- Check for external dependencies
- Ensure deterministic input data

#### High Variance
- Increase benchmark time with `-benchtime=10s`
- Check for garbage collection impact
- Review parallel benchmark setup

## Integration with CI/CD

### 1. Automated Benchmarking

#### GitHub Actions
```yaml
- name: Run Benchmarks
  run: |
    make benchmark
    # Parse and store results
```

#### Performance Regression Detection
```bash
# Compare with baseline
go test -bench=. -benchmem | tee current.txt
benchcmp baseline.txt current.txt
```

### 2. Performance Monitoring

#### Threshold Alerts
```go
// Set performance thresholds
const (
    MaxOperationTime = 1 * time.Millisecond
    MaxMemoryUsage  = 1024 // bytes
)
```

#### Continuous Monitoring
- Track benchmark results over time
- Alert on performance regressions
- Maintain performance baselines

## Conclusion

This benchmark implementation provides:

1. **Production-ready configurations** with realistic key sizes
2. **Comprehensive mocking** for isolated performance testing
3. **Detailed metrics** for performance analysis
4. **Scalable architecture** for adding new benchmarks
5. **CI/CD integration** for continuous monitoring

The benchmarks serve as a foundation for:
- Performance optimization decisions
- Capacity planning
- Regression detection
- Production readiness validation

Follow these guidelines to maintain accurate, reliable, and actionable benchmark results. 