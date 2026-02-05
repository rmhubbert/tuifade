# Benchmark Tests for tuifade

This document provides instructions on how to use and interpret the benchmark tests for the tuifade library.

## Overview

The benchmark suite helps identify performance bottlenecks in the `Fade` function, particularly when processing long content strings. The benchmarks measure:

- **Cache key generation performance** (base64 vs xxHash)
- **ANSI parsing performance** at various content sizes
- **Color interpolation performance**
- **Memory allocations** and their impact
- **Cache behavior** (warm vs cold)
- **Concurrent access performance**

## Running Benchmarks

### Run All Benchmarks

```bash
go test -bench=. -benchmem
```

### Run Specific Benchmark

```bash
# Run Fade benchmark with medium content
go test -bench="Fade/medium" -benchmem

# Run cache key generation benchmarks
go test -bench=GenerateContentCacheKey -benchmem

# Run interpolation benchmarks
go test -bench=Interpolate -benchmem
```

### Run with Specific Content Size

```bash
# Short content (100 chars)
go test -bench="Fade/short" -benchmem

# Medium content (1000 chars)
go test -bench="Fade/medium" -benchmem

# Long content (10000 chars)
go test -bench="Fade/long" -benchmem

# Very long content (100000 chars)
go test -bench="Fade/very_long" -benchmem
```

## Benchmark Output

The benchmark output follows Go's standard format:

```
BenchmarkFade/short-8          10000    85000 ns/op   288 B/op    2 allocs/op
BenchmarkFade/medium-8          5000    582000 ns/op  2816 B/op   2 allocs/op
BenchmarkFade/long-8             100    50562000 ns/op 82232 B/op  1441 allocs/op
```

### Understanding the Output

| Column | Description |
|--------|-------------|
| `BenchmarkFade/short-8` | Benchmark name with subtest and GOMAXPROCS (8 CPUs) |
| `10000` | Number of iterations (b.N) |
| `85000 ns/op` | Nanoseconds per operation (lower is better) |
| `288 B/op` | Bytes allocated per operation (lower is better) |
| `2 allocs/op` | Memory allocations per operation (lower is better) |

## Profiling

### CPU Profiling

Generate a CPU profile to identify which functions consume the most CPU time:

```bash
# Profile Fade function
go test -bench="Fade" -benchmem -cpuprofile=cpu.out -run=^$

# Profile cache key generation
go test -bench=GenerateContentCacheKey -benchmem -cpuprofile=cpu_cache.out -run=^$
```

Analyze the profile:

```bash
# Interactive analysis
go tool pprof cpu.out

# Top functions by CPU usage
go tool pprof -top cpu.out

# Generate flame graph (requires Graphviz)
go tool pprof -http=localhost:8080 cpu.out
```

### Memory Profiling

Generate a memory profile to identify memory allocation hotspots:

```bash
# Profile memory allocations
go test -bench="Fade" -benchmem -memprofile=mem.out -run=^$

# Profile with memory allocation reporting
go test -bench="Fade" -benchmem -memprofile=mem.out -benchmem -run=^$
```

Analyze the profile:

```bash
# Interactive analysis
go tool pprof mem.out

# Top functions by memory allocation
go tool pprof -top mem.out

# Generate flame graph
go tool pprof -http=localhost:8080 mem.out
```

### Comparing Profiles

Compare performance before and after optimizations:

```bash
# Compare CPU profiles
go tool pprof -top cpu_before.out cpu_after.out

# Compare memory profiles
go tool pprof -top mem_before.out mem_after.out

# Visual comparison (requires Graphviz)
go tool pprof -http=localhost:8080 cpu_before.out cpu_after.out
```

## Benchmark Functions

### Main Performance Benchmarks

#### `BenchmarkFade`

Table-driven benchmark testing performance across different content sizes:

- **short**: 100 characters
- **medium**: 1000 characters
- **long**: 10000 characters
- **very_long**: 100000 characters

**Purpose**: Identify scaling behavior and bottlenecks with increasing content size.

#### `BenchmarkFade_Memory`

Memory allocation benchmark with detailed allocation reporting.

**Purpose**: Measure memory efficiency and identify high-allocation areas.

#### `BenchmarkInterpolate`

Color interpolation performance benchmarks:

- **simple_gradient**: Black to white, 50% interpolation
- **complex_gradient**: Red to green, 30% interpolation
- **edge_case_min**: White to black, 0% interpolation
- **edge_case_max**: Black to white, 100% interpolation

**Purpose**: Measure color interpolation performance with different color pairs.

#### `BenchmarkANSIParse`

ANSI parsing performance benchmarks:

- **simple**: 1 ANSI color change
- **complex**: 10 ANSI color changes
- **very_complex**: 100 ANSI color changes

**Purpose**: Measure ANSI parsing performance with varying complexity.

### Cache Behavior Benchmarks

#### `BenchmarkFade_CacheWarm`

Tests performance with pre-warmed caches (10 warmup iterations before benchmarking).

**Purpose**: Measure performance of cached operations.

#### `BenchmarkFade_CacheCold`

Tests performance without cache warming (fresh cache for each iteration).

**Purpose**: Measure first-run performance and cache effectiveness.

### Concurrent Access Benchmarks

#### `BenchmarkFade_Concurrent`

Tests performance under concurrent access using `b.RunParallel`.

**Purpose**: Measure thread safety and concurrent performance.

### Specialized Benchmarks

#### `BenchmarkFade_SegmentCount`

Performance scaling with ANSI segment count:

- **segments_1**: 1 segment
- **segments_5**: 5 segments
- **segments_10**: 10 segments
- **segments_50**: 50 segments
- **segments_100**: 100 segments

**Purpose**: Identify how parsing complexity affects performance.

#### `BenchmarkFade_InterpolationValues`

Performance with different interpolation values:

- **interp_0.00**: No fade
- **interp_0.25**: 25% fade
- **interp_0.50**: 50% fade
- **interp_0.75**: 75% fade
- **interp_1.00**: Full fade

**Purpose**: Measure interpolation value impact on performance.

#### `BenchmarkGenerateContentCacheKey`

Cache key generation performance:

- **short**: 11 characters
- **medium**: 100 characters
- **long**: 1000 characters
- **very_long**: 10000 characters

**Purpose**: Measure cache key generation efficiency at different content sizes.

#### `BenchmarkGenerateContentCacheKey_Comparison`

Direct comparison between xxHash and base64 implementations.

**Purpose**: Quantify performance improvement from optimization.

## Interpreting Results

### Identifying Bottlenecks

1. **High ns/op**: Function is slow - look for computational inefficiencies
2. **High B/op**: High memory allocation - look for unnecessary allocations
3. **High allocs/op**: Many allocations - consider object pooling or pre-allocation

### Expected Performance Characteristics

#### Cache Key Generation (Optimized with xxHash)

- **Expected**: ~70-100 ns/op
- **Memory**: ~16 B/op, 1 alloc/op
- **Cache key size**: 13 characters (base-36 64-bit hash)

#### ANSI Parsing

- **Expected**: Scales with content size and segment count
- **Memory**: Scales with content size
- **Bottleneck**: Primary source of memory allocations for large content

#### Color Interpolation

- **Expected**: Fast (<1000 ns/op)
- **Memory**: Minimal allocations
- **Bottleneck**: Color space conversions (RGB↔HSL)

### Performance Regression Detection

1. Run benchmarks on baseline commit
2. Save baseline results: `go test -bench=. -benchmem > baseline.txt`
3. Make changes
4. Run benchmarks: `go test -bench=. -benchmem > current.txt`
5. Compare: `benchstat baseline.txt current.txt`

Install `benchstat`:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

## Troubleshooting

### Benchmarks Not Running

- Ensure you're using the `-bench` flag with a pattern
- Use `-benchmem` to enable memory profiling
- Use `-run=^$` to skip tests

### High Memory Usage

- Check ANSI parsing for unnecessary allocations
- Profile with `go tool pprof mem.out`
- Look for repeated allocations in hot paths

### Slow Performance

- Profile with `go tool pprof cpu.out`
- Check for expensive operations in hot paths
- Consider caching for repeated operations

## Best Practices

1. **Run benchmarks multiple times** to account for variance
2. **Use `-benchmem`** to track memory allocations
3. **Profile before and after optimizations** to quantify improvements
4. **Test with realistic content sizes** for your use case
5. **Monitor cache behavior** to ensure caching is effective
6. **Check concurrent performance** if your application is multi-threaded

## Example Workflow

```bash
# 1. Establish baseline
go test -bench=. -benchmem > baseline.txt

# 2. Make optimizations

# 3. Run benchmarks after changes
go test -bench=. -benchmem > current.txt

# 4. Compare results
benchstat baseline.txt current.txt

# 5. Profile to understand changes
go test -bench="Fade/medium" -benchmem -cpuprofile=cpu.out -run=^$
go tool pprof -http=localhost:8080 cpu.out
```

## Related Files

- `tuifade.go` - Main library code
- `tuifade_test.go` - Unit tests
- `tuifade_bench_test.go` - Benchmark tests
- `go.mod` - Dependencies (includes xxhash for optimized cache keys)
