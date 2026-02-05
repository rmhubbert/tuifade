package tuifade

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	"github.com/muesli/termenv"
)

// generateANSIString generates ANSI strings of specified length with mixed colored and plain text segments.
func generateANSIString(length int) string {
	result := make([]string, 0)
	currentLen := 0

	for currentLen < length {
		// Generate random color (30-36 are standard foreground colors)
		color := fmt.Sprintf("\x1b[%dm", 30+rand.Intn(7))
		text := "sample text "
		reset := "\x1b[0m"

		segment := color + text + reset
		if len(segment) > length-currentLen {
			segment = segment[:length-currentLen]
		}

		result = append(result, segment)
		currentLen += len(segment)
	}

	return strings.Join(result, "")
}

// generateComplexANSI generates ANSI strings with a specific number of color change segments.
func generateComplexANSI(segmentCount int) string {
	result := make([]string, 0)

	for i := 0; i < segmentCount; i++ {
		// Use sequential colors to ensure predictable segment count
		color := fmt.Sprintf("\x1b[%dm", 30+i%7)
		text := fmt.Sprintf("segment%d ", i)
		reset := "\x1b[0m"
		result = append(result, color+text+reset)
	}

	return strings.Join(result, "")
}

// newTestFader creates a Fader instance with a consistent test profile.
// This function is defined in tuifade_test.go and is reused here for consistency.

// BenchmarkFade is the main table-driven benchmark that tests performance across different content sizes.
func BenchmarkFade(b *testing.B) {
	testCases := []struct {
		name    string
		content string
	}{
		{"short", generateANSIString(100)},
		{"medium", generateANSIString(1000)},
		{"long", generateANSIString(10000)},
		{"very_long", generateANSIString(100000)},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			f := newTestFader(termenv.TrueColor)
			for i := 0; i < b.N; i++ {
				_, _ = f.Fade(tc.content, 0.5)
			}
		})
	}
}

// BenchmarkFade_Memory is a memory allocation benchmark to identify high-allocation areas.
func BenchmarkFade_Memory(b *testing.B) {
	content := generateANSIString(1000)
	f := newTestFader(termenv.TrueColor)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = f.Fade(content, 0.5)
	}
}

// BenchmarkInterpolate benchmarks specifically for the color interpolation performance.
func BenchmarkInterpolate(b *testing.B) {
	testCases := []struct {
		name          string
		bg, fg        string
		interpolation float64
	}{
		{"simple_gradient", "#000000", "#ffffff", 0.5},
		{"complex_gradient", "#ff0000", "#00ff00", 0.3},
		{"edge_case_min", "#ffffff", "#000000", 0.0},
		{"edge_case_max", "#000000", "#ffffff", 1.0},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			f := &Fader{}
			for i := 0; i < b.N; i++ {
				_, _ = f.Interpolate(tc.bg, tc.fg, tc.interpolation)
			}
		})
	}
}

// BenchmarkANSIParse benchmarks specifically for the ANSI parsing performance.
func BenchmarkANSIParse(b *testing.B) {
	testCases := []struct {
		name    string
		content string
	}{
		{"simple", generateComplexANSI(1)},
		{"complex", generateComplexANSI(10)},
		{"very_complex", generateComplexANSI(100)},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = ansiParse.Parse(tc.content)
			}
		})
	}
}

// BenchmarkFade_CacheWarm tests performance with pre-warmed cache.
func BenchmarkFade_CacheWarm(b *testing.B) {
	content := generateANSIString(1000)
	f := newTestFader(termenv.TrueColor)

	// Pre-warm caches
	for i := 0; i < 10; i++ {
		_, _ = f.Fade(content, 0.5)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = f.Fade(content, 0.5)
	}
}

// BenchmarkFade_CacheCold tests performance without cache warming.
func BenchmarkFade_CacheCold(b *testing.B) {
	content := generateANSIString(1000)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		f := newTestFader(termenv.TrueColor)
		_, _ = f.Fade(content, 0.5)
	}
}

// BenchmarkFade_Concurrent tests performance under concurrent access.
func BenchmarkFade_Concurrent(b *testing.B) {
	content := generateANSIString(1000)
	f := newTestFader(termenv.TrueColor)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = f.Fade(content, 0.5)
		}
	})
}

// BenchmarkFade_SegmentCount tests performance scaling with segment count.
func BenchmarkFade_SegmentCount(b *testing.B) {
	segmentCounts := []int{1, 5, 10, 50, 100}

	for _, count := range segmentCounts {
		content := generateComplexANSI(count)
		b.Run(fmt.Sprintf("segments_%d", count), func(b *testing.B) {
			f := newTestFader(termenv.TrueColor)
			for i := 0; i < b.N; i++ {
				_, _ = f.Fade(content, 0.5)
			}
		})
	}
}

// BenchmarkFade_InterpolationValues tests performance with different interpolation values.
func BenchmarkFade_InterpolationValues(b *testing.B) {
	interpolations := []float64{0.0, 0.25, 0.5, 0.75, 1.0}
	content := generateANSIString(1000)

	for _, interp := range interpolations {
		b.Run(fmt.Sprintf("interp_%.2f", interp), func(b *testing.B) {
			f := newTestFader(termenv.TrueColor)
			for i := 0; i < b.N; i++ {
				_, _ = f.Fade(content, interp)
			}
		})
	}
}

// BenchmarkGenerateContentCacheKey benchmarks cache key generation performance.
func BenchmarkGenerateContentCacheKey(b *testing.B) {
	testCases := []struct {
		name    string
		content string
	}{
		{"short", "Hello World"},
		{"medium", generateANSIString(100)},
		{"long", generateANSIString(1000)},
		{"very_long", generateANSIString(10000)},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = generateContentCacheKey(tc.content)
			}
		})
	}
}

// BenchmarkGenerateContentCacheKey_Comparison compares xxHash vs base64 performance.
func BenchmarkGenerateContentCacheKey_Comparison(b *testing.B) {
	content := generateANSIString(1000)

	b.Run("xxHash", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = generateContentCacheKey(content)
		}
	})

	b.Run("base64", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = base64.StdEncoding.EncodeToString([]byte(content))
		}
	})
}
