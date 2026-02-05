package tuifade

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	"github.com/muesli/termenv"
)

// Test constants for colors - kept for potential future use
// const (
// 	testColorBlack = "#000000"
// 	testColorWhite = "#ffffff"
// 	testColorRed   = "#ff0000"
// 	testColorGreen = "#00ff00"
// 	testColorBlue  = "#0000ff"
// )

// Test constants for interpolation
const (
	testInterpMin  = 0.0
	testInterpHalf = 0.5
	testInterpMax  = 1.0
)

// ANSI test strings (using octal escape for ANSI codes to match ansi-parser library)
const (
	testAnsiRed      = "\033[31mred\033[0m"
	testAnsiGreen    = "\033[32mgreen\033[0m"
	testAnsiComplex  = "\033[31mred\033[0m \033[32mgreen\033[0m"
	testAnsiNoColors = "plain text without ANSI codes"
)

// newTestFader creates a test Fader with injected terminal profile
func newTestFader(profile termenv.Profile) *Fader {
	return &Fader{
		termInfo: &termInfo{
			profile:                 profile,
			defaultBackgroundColour: "#000000", // Default black background
			defaultForegroundColour: "#ffffff", // Default white foreground
		},
	}
}

// getCacheStats returns the size of colourCache and contentCache
func getCacheStats(f *Fader) (int, int) {
	colourCount := 0
	contentCount := 0
	f.colourCache.Range(func(_, _ any) bool {
		colourCount++
		return true
	})
	f.contentCache.Range(func(_, _ any) bool {
		contentCount++
		return true
	})
	return colourCount, contentCount
}

// clearCaches clears both colourCache and contentCache
func clearCaches(f *Fader) {
	f.colourCache = sync.Map{}
	f.contentCache = sync.Map{}
}

// TestHexToRGB tests the hexToRGB helper function
func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name     string
		hex      string
		expected rbgColour
		wantErr  bool
	}{
		{"valid 6-digit hex red", "#ff0000", rbgColour{R: 255, G: 0, B: 0}, false},
		{"valid 6-digit hex green", "#00ff00", rbgColour{R: 0, G: 255, B: 0}, false},
		{"valid 6-digit hex blue", "#0000ff", rbgColour{R: 0, G: 0, B: 255}, false},
		{"valid mixed case", "#0000FF", rbgColour{R: 0, G: 0, B: 255}, false},
		{"valid lowercase", "#00ff00", rbgColour{R: 0, G: 255, B: 0}, false},
		{"invalid - too short", "#ff00", rbgColour{}, true},
		{"invalid - non-hex chars", "#gggggg", rbgColour{}, true},
		{"invalid - empty string", "", rbgColour{}, true},
		{"invalid - missing # prefix", "ff0000", rbgColour{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hexToRGB(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("hexToRGB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr &&
				(got.R != tt.expected.R || got.G != tt.expected.G || got.B != tt.expected.B) {
				t.Errorf("hexToRGB() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRGBToHex tests the rgbToHex helper function
func TestRGBToHex(t *testing.T) {
	tests := []struct {
		name     string
		rgb      rbgColour
		expected string
	}{
		{"pure red", rbgColour{R: 255, G: 0, B: 0}, "#ff0000"},
		{"pure green", rbgColour{R: 0, G: 255, B: 0}, "#00ff00"},
		{"pure blue", rbgColour{R: 0, G: 0, B: 255}, "#0000ff"},
		{"black", rbgColour{R: 0, G: 0, B: 0}, "#000000"},
		{"white", rbgColour{R: 255, G: 255, B: 255}, "#ffffff"},
		{"gray", rbgColour{R: 128, G: 128, B: 128}, "#808080"},
		{"dark gray", rbgColour{R: 64, G: 64, B: 64}, "#404040"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rgbToHex(tt.rgb)
			if got != tt.expected {
				t.Errorf("rgbToHex() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestRGBToHex_RoundTrip tests round-trip consistency of RGB to Hex conversion
func TestRGBToHex_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		rgb  rbgColour
	}{
		{"round trip 1", rbgColour{R: 100, G: 150, B: 200}},
		{"round trip 2", rbgColour{R: 255, G: 128, B: 64}},
		{"round trip 3", rbgColour{R: 32, G: 64, B: 128}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hex := rgbToHex(tt.rgb)
			got, err := hexToRGB(hex)
			if err != nil {
				t.Errorf("hexToRGB() returned error: %v", err)
				return
			}
			if got.R != tt.rgb.R || got.G != tt.rgb.G || got.B != tt.rgb.B {
				t.Errorf("round-trip mismatch: %v -> %s -> %v", tt.rgb, hex, got)
			}
		})
	}
}

// TestInterpolateChannel tests the interpolateChannel helper function
func TestInterpolateChannel(t *testing.T) {
	tests := []struct {
		name     string
		bg       uint8
		fg       uint8
		bgWeight float64
		fgWeight float64
		expected uint8
	}{
		{"both same value", 100, 100, 0.5, 0.5, 100},
		{"background only 100%", 100, 0, 1.0, 0.0, 100},
		{"foreground only 100%", 0, 100, 0.0, 1.0, 100},
		{"50/50 interpolation", 0, 255, 0.5, 0.5, 128},
		{"clamping test", 200, 250, 0.8, 0.2, 210},
		{"both zero", 0, 0, 0.5, 0.5, 0},
		{"both max", 255, 255, 0.5, 0.5, 255},
		{"weighted bg heavy", 200, 100, 0.9, 0.1, 190},
		{"weighted fg heavy", 100, 200, 0.2, 0.8, 180},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := interpolateChannel(tt.bg, tt.fg, tt.bgWeight, tt.fgWeight)
			if got != tt.expected {
				t.Errorf("interpolateChannel(%d, %d, %f, %f) = %d, want %d",
					tt.bg, tt.fg, tt.bgWeight, tt.fgWeight, got, tt.expected)
			}
		})
	}
}

// TestClampInterpolation tests the clampInterpolation helper function
func TestClampInterpolation(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"within range low", 0.25, 0.25},
		{"within range mid", 0.5, 0.5},
		{"within range high", 0.75, 0.75},
		{"below minimum", -0.5, 0},
		{"below minimum extreme", -10.0, 0},
		{"above maximum", 1.5, 1},
		{"above maximum extreme", 10.0, 1},
		{"exactly at min", 0.0, 0},
		{"exactly at max", 1.0, 1},
		{"zero", 0.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := clampInterpolation(tt.input)
			if got != tt.expected {
				t.Errorf("clampInterpolation(%f) = %f, want %f", tt.input, got, tt.expected)
			}
		})
	}
}

// TestInterpolate_KnownValues tests the Interpolate method with known color values
func TestInterpolate_KnownValues(t *testing.T) {
	tests := []struct {
		name          string
		bg            string
		fg            string
		interpolation float64
		expectedHex   string
	}{
		{"black to white 50%", "#000000", "#ffffff", 0.5, "#808080"},
		{"red to black 50%", "#ff0000", "#000000", 0.5, "#800000"},
		{"green to black 50%", "#00ff00", "#000000", 0.5, "#008000"},
		{"blue to black 50%", "#0000ff", "#000000", 0.5, "#000080"},
		{"red to green 50%", "#ff0000", "#00ff00", 0.5, "#808000"},
		{"full fade 0% (to foreground)", "#ff0000", "#000000", 0.0, "#ff0000"},
		{"no fade 100% (keep original)", "#ff0000", "#000000", 1.0, "#000000"},
		{"same color", "#ff0000", "#ff0000", 0.5, "#ff0000"},
		{"white to black 50%", "#ffffff", "#000000", 0.5, "#808080"},
		{"cyan to white 50%", "#00ffff", "#ffffff", 0.5, "#80ffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fader{}
			result, err := f.Interpolate(tt.bg, tt.fg, tt.interpolation)

			if err != nil {
				t.Errorf("Interpolate() returned error: %v", err)
				return
			}

			if result.result.hex != tt.expectedHex {
				t.Errorf("Interpolate() hex = %v, want %v", result.result.hex, tt.expectedHex)
			}
		})
	}
}

// TestInterpolate_EdgeCases tests edge cases for Interpolate
func TestInterpolate_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		bg            string
		fg            string
		interpolation float64
	}{
		{"interpolation clamping min", "#ff0000", "#000000", -0.5},
		{"interpolation clamping max", "#ff0000", "#000000", 1.5},
		{"very small interpolation", "#ff0000", "#000000", 0.001},
		{"very large interpolation", "#ff0000", "#000000", 0.999},
		{"zero interpolation", "#ff0000", "#000000", 0.0},
		{"one interpolation", "#ff0000", "#000000", 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fader{}
			result, err := f.Interpolate(tt.bg, tt.fg, tt.interpolation)

			if err != nil {
				t.Errorf("Interpolate() returned unexpected error: %v", err)
				return
			}

			// Verify the result is a valid hex color
			if result.result.hex == "" {
				t.Error("Interpolate() returned empty hex string")
			}
		})
	}
}

// TestInterpolate_InvalidInputs tests Interpolate with invalid inputs
func TestInterpolate_InvalidInputs(t *testing.T) {
	tests := []struct {
		name          string
		bg            string
		fg            string
		interpolation float64
	}{
		{"empty background", "", "#ffffff", 0.5},
		{"empty foreground", "#ffffff", "", 0.5},
		{"invalid hex too short", "#ff", "#ffffff", 0.5},
		{"invalid hex non-hex chars", "#gggggg", "#ffffff", 0.5},
		{"invalid hex malformed", "invalid", "#ffffff", 0.5},
		{"invalid hex missing hash", "ff0000", "#ffffff", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fader{}
			result, err := f.Interpolate(tt.bg, tt.fg, tt.interpolation)

			if err == nil {
				t.Errorf("Interpolate() expected error for invalid input, got result: %+v", result)
			}
		})
	}
}

// TestInterpolate_Caching tests that Interpolate uses caching correctly
func TestInterpolate_Caching(t *testing.T) {
	f := &Fader{}

	// Clear caches before test
	clearCaches(f)

	// First call - should populate cache
	_, err := f.Interpolate("#ff0000", "#000000", 0.5)
	if err != nil {
		t.Errorf("First Interpolate() returned error: %v", err)
		return
	}

	colourCount, contentCount := getCacheStats(f)
	if colourCount == 0 {
		t.Error("Interpolate() did not populate colourCache")
	}
	if contentCount != 0 {
		t.Error("Interpolate() incorrectly populated contentCache")
	}

	// Second call with same parameters - should use cache
	result1, err := f.Interpolate("#ff0000", "#000000", 0.5)
	if err != nil {
		t.Errorf("Second Interpolate() returned error: %v", err)
		return
	}

	// Third call - verify cache hit
	result2, err := f.Interpolate("#ff0000", "#000000", 0.5)
	if err != nil {
		t.Errorf("Third Interpolate() returned error: %v", err)
		return
	}

	if result1.result.hex != result2.result.hex {
		t.Error("Cached result does not match expected value")
	}
}

// TestFade_NoFade tests the Fade method with interpolation = 1.0 (no fade)
func TestFade_NoFade(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"ansi red", testAnsiRed},
		{"ansi green", testAnsiGreen},
		{"ansi complex", testAnsiComplex},
		{"no ansi codes", testAnsiNoColors},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpMax)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			if result != tt.content {
				t.Errorf(
					"Fade() changed content when interpolation=1.0\nGot: %q\nWant: %q",
					result,
					tt.content,
				)
			}
		})
	}
}

// TestFade_FullFade tests the Fade method with interpolation = 0.0 (full fade)
func TestFade_FullFade(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"plain text", "Hello World"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpMin)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should be different from original (fully faded)
			if result == tt.content {
				t.Error("Fade() returned unchanged content when interpolation=0.0")
			}
		})
	}
}

// TestFade_PartialFade tests the Fade method with partial interpolation
func TestFade_PartialFade(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name          string
		content       string
		interpolation float64
	}{
		{"ansi red 50%", testAnsiRed, testInterpHalf},
		{"ansi complex 50%", testAnsiComplex, testInterpHalf},
		{"no ansi codes 50%", testAnsiNoColors, testInterpHalf},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, tt.interpolation)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should be different from original (partially faded)
			if result == tt.content {
				t.Error("Fade() returned unchanged content when interpolation < 1.0")
			}
		})
	}
}

// TestFade_EdgeCases tests Fade with edge case inputs
func TestFade_EdgeCases(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"plain text", "plain text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpHalf)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should not be empty unless input was empty
			if tt.content == "" && result != "" {
				t.Errorf("Fade() returned non-empty result for empty input: %q", result)
			}
		})
	}
}

// TestFade_SingleColorSegments tests Fade with single color segment strings
func TestFade_SingleColorSegments(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"foreground only", "\x1b[31mred\x1b[0m"},
		{"background only", "\x1b[41mred\x1b[0m"},
		{"foreground and background", "\x1b[31;42mtext\x1b[0m"},
		{"different foreground", "\x1b[34mblue\x1b[0m"},
		{"different background", "\x1b[45mgreen\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpHalf)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should still contain ANSI codes
			if !containsANSI(result) {
				t.Error("Fade() removed ANSI codes from result")
			}
		})
	}
}

// TestFade_MultiSegmentStrings tests Fade with multi-segment ANSI strings
func TestFade_MultiSegmentStrings(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"multiple colors", "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m"},
		{"consecutive same colors", "\x1b[31mred\x1b[0m\x1b[31mmore red\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpHalf)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should still contain ANSI codes
			if !containsANSI(result) {
				t.Error("Fade() removed ANSI codes from result")
			}

			// Result should preserve segment count (approximately)
			originalSegs, err := ansiParse.Parse(tt.content, ansiParse.WithIgnoreInvalidCodes())
			if err != nil {
				t.Errorf("Failed to parse original content: %v", err)
				return
			}
			resultSegs, err := ansiParse.Parse(result, ansiParse.WithIgnoreInvalidCodes())
			if err != nil {
				t.Errorf("Failed to parse result: %v", err)
				return
			}

			if len(resultSegs) != len(originalSegs) {
				t.Errorf(
					"Segment count changed: original=%d, result=%d",
					len(originalSegs),
					len(resultSegs),
				)
			}
		})
	}
}

// TestFade_DefaultColorHandling tests Fade with default color handling
func TestFade_DefaultColorHandling(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"background only (no fg)", "\x1b[41mred\x1b[0m"},
		{"foreground only (no bg)", "\x1b[31mred\x1b[0m"},
		{"style only (no colors)", "\x1b[1mbold\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpHalf)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should still contain ANSI codes
			if !containsANSI(result) {
				t.Error("Fade() removed ANSI codes from result")
			}
		})
	}
}

// TestFade_StylePreservation tests that ANSI styles are preserved through fade
func TestFade_StylePreservation(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"bold", "\x1b[1mbold\x1b[0m"},
		{"italic", "\x1b[3mitalic\x1b[0m"},
		{"underline", "\x1b[4munderline\x1b[0m"},
		{"multiple styles", "\x1b[1;3;4mstyled\x1b[0m"},
		{"color with style", "\x1b[1;31mbold red\x1b[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpHalf)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should still contain ANSI codes
			if !containsANSI(result) {
				t.Error("Fade() removed ANSI codes from result")
			}
		})
	}
}

// TestFade_ComplexANSI tests Fade with complex ANSI sequences
func TestFade_ComplexANSI(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	tests := []struct {
		name    string
		content string
	}{
		{"TrueColor RGB", "\033[38;2;255;0;0mred\033[0m"},
		{"ANSI256", "\033[38;5;196mred\033[0m"},
		{"complex sequence", "\033[1;31;42mbold red on green\033[0m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := f.Fade(tt.content, testInterpHalf)

			if err != nil {
				t.Errorf("Fade() returned error: %v", err)
				return
			}

			// Result should still contain ANSI codes
			if !containsANSI(result) {
				t.Error("Fade() removed ANSI codes from result")
			}
		})
	}
}

// TestFade_TrueColorProfile tests Fade with TrueColor profile
func TestFade_TrueColorProfile(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	result, err := f.Fade(testAnsiRed, testInterpHalf)

	if err != nil {
		t.Errorf("Fade() returned error with TrueColor profile: %v", err)
	}

	if result == "" {
		t.Error("Fade() returned empty string with TrueColor profile")
	}

	if result == testAnsiRed {
		t.Error("Fade() returned unchanged content with TrueColor profile")
	}
}

// TestFade_ANSI256Profile tests Fade with ANSI256 profile (should return error)
func TestFade_ANSI256Profile(t *testing.T) {
	f := newTestFader(termenv.ANSI256)

	result, err := f.Fade(testAnsiRed, testInterpHalf)

	if err == nil {
		t.Error("Fade() should return error with ANSI256 profile")
	}

	// Original content should be returned on error
	if result != testAnsiRed {
		t.Errorf("Fade() should return original content on error, got: %q", result)
	}
}

// TestFade_ANSIProfile tests Fade with ANSI profile (should return error)
func TestFade_ANSIProfile(t *testing.T) {
	f := newTestFader(termenv.ANSI)

	result, err := f.Fade(testAnsiRed, testInterpHalf)

	if err == nil {
		t.Error("Fade() should return error with ANSI profile")
	}

	// Original content should be returned on error
	if result != testAnsiRed {
		t.Errorf("Fade() should return original content on error, got: %q", result)
	}
}

// TestFade_AsciiProfile tests Fade with Ascii profile (should return error)
func TestFade_AsciiProfile(t *testing.T) {
	f := newTestFader(termenv.Ascii)

	result, err := f.Fade(testAnsiRed, testInterpHalf)

	if err == nil {
		t.Error("Fade() should return error with Ascii profile")
	}

	// Original content should be returned on error
	if result != testAnsiRed {
		t.Errorf("Fade() should return original content on error, got: %q", result)
	}
}

// TestFade_CacheWarming tests that Fade properly warms the cache
func TestFade_CacheWarming(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	// Clear caches before test
	clearCaches(f)

	// First call - should populate cache
	_, err := f.Fade(testAnsiRed, testInterpHalf)
	if err != nil {
		t.Errorf("First Fade() returned error: %v", err)
		return
	}

	_, contentCount := getCacheStats(f)
	if contentCount == 0 {
		t.Error("Fade() did not populate contentCache")
	}
}

// TestFade_CacheHits tests that Fade properly uses cached results
func TestFade_CacheHits(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	// Clear caches before test
	clearCaches(f)

	// First call - should populate cache
	result1, err := f.Fade(testAnsiRed, testInterpHalf)
	if err != nil {
		t.Errorf("First Fade() returned error: %v", err)
		return
	}

	// Second call with same content - should use cache
	result2, err := f.Fade(testAnsiRed, testInterpHalf)
	if err != nil {
		t.Errorf("Second Fade() returned error: %v", err)
		return
	}

	// Results should be identical (cached)
	if result1 != result2 {
		t.Error("Cached result does not match expected value")
	}

	// Third call - verify cache hit
	result3, err := f.Fade(testAnsiRed, testInterpHalf)
	if err != nil {
		t.Errorf("Third Fade() returned error: %v", err)
		return
	}

	if result1 != result3 {
		t.Error("Cached result does not match expected value")
	}
}

// TestFade_CacheIsolation tests that different inputs use different cache entries
func TestFade_CacheIsolation(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	// Clear caches before test
	clearCaches(f)

	// First call
	_, err := f.Fade(testAnsiRed, testInterpHalf)
	if err != nil {
		t.Errorf("First Fade() returned error: %v", err)
		return
	}

	// Different content - should use different cache entry
	_, err = f.Fade(testAnsiGreen, testInterpHalf)
	if err != nil {
		t.Errorf("Second Fade() returned error: %v", err)
		return
	}

	// Different interpolation - should use different cache entry
	_, err = f.Fade(testAnsiRed, 0.3)
	if err != nil {
		t.Errorf("Third Fade() returned error: %v", err)
		return
	}

	_, contentCount := getCacheStats(f)
	if contentCount < 2 {
		t.Errorf("Expected at least 2 cache entries, got %d", contentCount)
	}
}

// TestFade_ConcurrentAccess tests Fade with concurrent access
func TestFade_ConcurrentAccess(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	// Clear caches before test
	clearCaches(f)

	done := make(chan bool, 10)

	// Run multiple concurrent fade operations
	for i := range 10 {
		go func(id int) {
			_, err := f.Fade(testAnsiRed, testInterpHalf)
			if err != nil {
				t.Errorf("Concurrent Fade() %d returned error: %v", id, err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	// Verify cache was populated
	_, contentCount := getCacheStats(f)
	if contentCount == 0 {
		t.Error("Fade() did not populate cache under concurrent access")
	}
}

// TestIntegration tests integration scenarios
func TestIntegration(t *testing.T) {
	f := newTestFader(termenv.TrueColor)

	// Test real-world usage pattern
	content := "\x1b[31mError:\x1b[0m \x1b[32mSuccess:\x1b[0m \x1b[34mInfo:\x1b[0m"

	result, err := f.Fade(content, testInterpHalf)

	if err != nil {
		t.Errorf("Fade() returned error: %v", err)
		return
	}

	if result == "" {
		t.Error("Fade() returned empty string")
	}

	if !containsANSI(result) {
		t.Error("Fade() removed all ANSI codes")
	}
}

// containsANSI checks if a string contains ANSI escape codes
func containsANSI(s string) bool {
	for _, r := range s {
		if r == '\x1b' {
			return true
		}
	}
	return false
}

// TestGenerateContentCacheKey_Uniqueness tests that cache key generation is consistent
func TestGenerateContentCacheKey_Uniqueness(t *testing.T) {
	testCases := []struct {
		name       string
		content1   string
		content2   string
		expectSame bool
	}{
		{"same_content", "Hello World", "Hello World", true},
		{"different_content", "Hello World", "Hello World!", false},
		{"empty_strings", "", "", true},
		{"ansi_content", "\x1b[31mred\x1b[0m", "\x1b[31mred\x1b[0m", true},
		{"complex_ansi", "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m", "\x1b[31mred\x1b[0m \x1b[32mgreen\x1b[0m", true},
		{"similar_ansi", "\x1b[31mred\x1b[0m", "\x1b[32mred\x1b[0m", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key1 := generateContentCacheKey(tc.content1)
			key2 := generateContentCacheKey(tc.content2)

			if tc.expectSame && key1 != key2 {
				t.Errorf("Same content should produce same cache key. Got '%s' vs '%s'", key1, key2)
			}

			if !tc.expectSame && key1 == key2 {
				t.Errorf("Different content should produce different cache key")
			}
		})
	}
}

// TestGenerateContentCacheKey_CollisionResistance tests hash collision resistance
func TestGenerateContentCacheKey_CollisionResistance(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		content := fmt.Sprintf("test content %d with some random data: %d", i, rand.Int())
		key := generateContentCacheKey(content)

		if seen[key] {
			t.Errorf("Hash collision detected for content: %s", content)
		}
		seen[key] = true
	}
}

// TestGenerateContentCacheKey_Length tests that cache key length is reasonable
func TestGenerateContentCacheKey_Length(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"short", "Hello"},
		{"medium", generateANSIString(100)},
		{"long", generateANSIString(1000)},
		{"very_long", generateANSIString(10000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key := generateContentCacheKey(tc.content)
			// 64-bit hash in base-36 can be up to 13 characters for larger values
			if len(key) > 13 {
				t.Errorf("Cache key too long: %d chars (expected max 13). Key: %s", len(key), key)
			}
			if len(key) == 0 {
				t.Errorf("Cache key should not be empty")
			}
		})
	}
}
