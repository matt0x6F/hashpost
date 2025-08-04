# HashPost Crypto Benchmarks

## Overview

This directory contains comprehensive benchmark documentation and implementation guides for HashPost's cryptographic operations. The benchmarks use production-ready configurations with realistic key sizes and comprehensive mocking to provide accurate performance insights.

## 📚 Documentation

### 1. [Crypto Performance Analysis](./crypto-performance-analysis.md)
**Comprehensive analysis of benchmark results** including detailed performance metrics, insights, bottlenecks, scalability analysis, and production readiness assessment.

### 2. [Benchmark Implementation Guide](./benchmark-implementation-guide.md)
**Technical guide for implementing and running benchmarks** including architecture overview, implementation methodology, running benchmarks, interpreting results, troubleshooting, and CI/CD integration.

### 3. [Performance Quick Reference](./performance-quick-reference.md)
**Quick reference card** with key performance metrics, top performers, optimization opportunities, and monitoring thresholds.

## 🚀 Quick Start

### Run All Benchmarks
```bash
make benchmark
```

### Run Specific Benchmark Suites
```bash
# IBE System benchmarks
go test ./internal/ibe -bench=. -benchmem -run=^$ -count=1

# Authentication benchmarks
go test ./internal/api/handlers -bench="Benchmark(Password|JWT|SHA256|RandomBytes|Hex)" -benchmem -run=^$ -count=1

# Handler-level benchmarks
go test ./internal/api/handlers -bench=BenchmarkAuthHandler -benchmem -run=^$ -count=1

# Correlation benchmarks
go test ./internal/api/handlers -bench=BenchmarkIBE -benchmem -run=^$ -count=1
```

## 🔧 Implementation Details

### Production-Ready Configuration
- **IBE Domain Keys**: 256-bit (32-byte) production keys
- **JWT Secret**: 256-bit production secret
- **Password Hashing**: SHA-256 (intentionally slow)
- **Mock Dependencies**: Fully mocked DAOs and external services

### Benchmark Architecture
1. **IBE System Benchmarks** (`internal/ibe/ibe_benchmark_test.go`)
2. **Authentication Benchmarks** (`internal/api/handlers/auth_benchmark_test.go`)
3. **Handler-Level Benchmarks** (`internal/api/handlers/auth_handler_benchmark_test.go`)
4. **Correlation Benchmarks** (`internal/api/handlers/correlation_benchmark_test.go`)

## 🎯 Key Insights

### Performance Categories
- **High Performance (>1M ops/sec)**: Hex operations, random bytes, SHA-256, parallel pseudonyms
- **Medium Performance (100K-1M ops/sec)**: JWT operations, IBE correlation, identity decryption
- **Low Performance (<100 ops/sec)**: Password operations (intentional for security)

### Top Performers
- **Fastest**: Hex decoding (23.7M ops/sec)
- **Most Scalable**: Parallel pseudonym generation (5x improvement)
- **Most Memory Efficient**: Random bytes (16 B/op)
- **Security Focused**: Password operations (intentionally slow)

### Optimization Opportunities
1. **JWT Validation**: Implement caching (2-3x improvement expected)
2. **Parallel Processing**: Leverage parallel pseudonym generation (5x improvement)
3. **Memory Pooling**: Reduce allocations for IBE operations (10-20% improvement)
4. **Role Key Caching**: Cache validated role keys (3-5x improvement)

## 🎯 Production Readiness

### ✅ Strengths
- **High throughput**: Most operations achieve 1M+ ops/sec
- **Memory efficient**: Minimal memory usage per operation
- **Security focused**: Appropriate performance trade-offs for security
- **Parallel ready**: Significant gains with parallel processing
- **Realistic configurations**: Production key sizes and settings

### 🔧 Areas for Improvement
- **JWT validation**: Implement caching for better performance
- **Password operations**: Consider distributed processing
- **Memory management**: Implement object pooling for frequent allocations

## 🔍 Monitoring and Alerting

### Performance Thresholds
- **Warning**: Operations > 1ms/op
- **Critical**: Operations > 10ms/op
- **Exception**: Password operations (intentionally slow)

### Memory Thresholds
- **Warning**: Operations > 1KB/op
- **Critical**: Operations > 10KB/op

## 🤝 Contributing

When adding new benchmarks:

1. **Use production configurations** with realistic key sizes
2. **Implement comprehensive mocking** for isolated testing
3. **Follow the benchmark patterns** in existing files
4. **Include memory allocation metrics** with `-benchmem`
5. **Document performance expectations** and trade-offs

## 📝 License

This documentation is part of the HashPost project and follows the same license terms. 