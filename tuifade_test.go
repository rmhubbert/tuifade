package tuifade

import (
	"math"
	"strings"
	"testing"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data for various scenarios
var testColors = []struct {
	name string
	hex  string
	rgb  rbgColour
	hsl  hslColour
}{
	{
		name: "Pure Red",
		hex:  "#ff0000",
		rgb:  rbgColour{R: 255, G: 0, B: 0},
		hsl:  hslColour{H: 0, S: 100, L: 50},
	},
	{
		name: "Pure Green",
		hex:  "#00ff00",
		rgb:  rbgColour{R: 0, G: 255, B: 0},
		hsl:  hslColour{H: 43200, S: 100, L: 50},
	},
	{
		name: "Pure Blue",
		hex:  "#0000ff",
		rgb:  rbgColour{R: 0, G: 0, B: 255},
		hsl:  hslColour{H: 86400, S: 100, L: 50},
	},
	{
		name: "Pure White",
		hex:  "#ffffff",
		rgb:  rbgColour{R: 255, G: 255, B: 255},
		hsl:  hslColour{H: 0, S: 0, L: 100},
	},
	{
		name: "Pure Black",
		hex:  "#000000",
		rgb:  rbgColour{R: 0, G: 0, B: 0},
		hsl:  hslColour{H: 0, S: 0, L: 0},
	},
}

var testANSIStrings = []struct {
	name    string
	content string
}{
	{
		name:    "Simple colored text",
		content: "\x1b[31mRed text\x1b[0m",
	},
	{
		name:    "Background and foreground",
		content: "\x1b[31;42mRed on green\x1b[0m",
	},
	{
		name:    "Multiple segments",
		content: "\x1b[31mRed\x1b[32mGreen\x1b[33mYellow\x1b[0m",
	},
	{
		name:    "Styles and colors",
		content: "\x1b[1;31;44mBold red on blue\x1b[0m",
	},
	{
		name:    "No color codes",
		content: "Plain text without any ANSI codes",
	},
	{
		name:    "Empty string",
		content: "",
	},
}

// ColorsEqual compares two RGB colors with tolerance for floating point precision
func ColorsEqual(a, b rbgColour, tolerance float64) bool {
	return math.Abs(float64(a.R)-float64(b.R)) <= tolerance &&
		math.Abs(float64(a.G)-float64(b.G)) <= tolerance &&
		math.Abs(float64(a.B)-float64(b.B)) <= tolerance
}

// HexColorsEqual compares two hex color strings
func HexColorsEqual(a, b string) bool {
	rgbA, err := hexToRGB(a)
	if err != nil {
		return false
	}
	rgbB, err := hexToRGB(b)
	if err != nil {
		return false
	}
	return rgbA.R == rgbB.R && rgbA.G == rgbB.G && rgbA.B == rgbB.B
}

// TestHelperFunctions tests the helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("hexToRGB", func(t *testing.T) {
		for _, tc := range testColors {
			t.Run(tc.name, func(t *testing.T) {
				rgb, err := hexToRGB(tc.hex)
				require.NoError(t, err)
				assert.Equal(t, tc.rgb, rgb)
			})
		}
	})

	t.Run("hexToRGB_errors", func(t *testing.T) {
		testCases := []struct {
			name string
			hex  string
		}{
			{"missing #", "ff0000"},
			{"too short", "#f00"},
			{"invalid characters", "#gg0000"},
			{"empty string", ""},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := hexToRGB(tc.hex)
				assert.Error(t, err)
			})
		}
	})

	t.Run("rgbToHex", func(t *testing.T) {
		for _, tc := range testColors {
			t.Run(tc.name, func(t *testing.T) {
				hex := rgbToHex(tc.rgb)
				assert.Equal(t, tc.hex, hex)
			})
		}
	})

	t.Run("roundTripConversion", func(t *testing.T) {
		for _, tc := range testColors {
			t.Run(tc.name, func(t *testing.T) {
				rgb, err := hexToRGB(tc.hex)
				require.NoError(t, err)
				hex := rgbToHex(rgb)
				assert.Equal(t, tc.hex, hex)
			})
		}
	})

	t.Run("interpolateChannel", func(t *testing.T) {
		testCases := []struct {
			name     string
			bg       uint8
			fg       uint8
			bgWeight float64
			fgWeight float64
			expected uint8
		}{
			{"midpoint", 0, 255, 0.5, 0.5, 128},
			{"full background", 0, 255, 1.0, 0.0, 0},
			{"full foreground", 0, 255, 0.0, 1.0, 255},
			{"zero values", 0, 0, 0.5, 0.5, 0},
			{"max values", 255, 255, 0.5, 0.5, 255},
			{"rounding up", 0, 255, 0.5, 0.5, 128}, // 127.5 rounds to 128
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := interpolateChannel(tc.bg, tc.fg, tc.bgWeight, tc.fgWeight)
				assert.Equal(t, tc.expected, result)
			})
		}
	})

	t.Run("hexToHSL", func(t *testing.T) {
		for _, tc := range testColors {
			t.Run(tc.name, func(t *testing.T) {
				hsl, err := hexToHSL(tc.hex)
				require.NoError(t, err)
				// Allow small tolerance for floating point comparisons
				assert.InDelta(t, tc.hsl.H, hsl.H, 1.0, "Hue mismatch")
				assert.InDelta(t, tc.hsl.S, hsl.S, 1.0, "Saturation mismatch")
				assert.InDelta(t, tc.hsl.L, hsl.L, 1.0, "Lightness mismatch")
			})
		}
	})
}

// TestInterpolateFunctionality tests the Interpolate function with normal cases
func TestInterpolateFunctionality(t *testing.T) {
	testCases := []struct {
		name          string
		background    string
		foreground    string
		interpolation float64
		expected      string
		expectError   bool
	}{
		{
			name:          "red to blue at midpoint",
			background:    "#ff0000",
			foreground:    "#0000ff",
			interpolation: 0.5,
			expected:      "#800080", // purple
			expectError:   false,
		},
		{
			name:          "no fade (interpolation = 1.0)",
			background:    "#ff0000",
			foreground:    "#0000ff",
			interpolation: 1.0,
			expected:      "#0000ff", // pure foreground
			expectError:   false,
		},
		{
			name:          "full fade (interpolation = 0.0)",
			background:    "#ff0000",
			foreground:    "#0000ff",
			interpolation: 0.0,
			expected:      "#ff0000", // pure background
			expectError:   false,
		},
		{
			name:          "black to white at midpoint",
			background:    "#000000",
			foreground:    "#ffffff",
			interpolation: 0.5,
			expected:      "#808080", // gray
			expectError:   false,
		},
		{
			name:          "green to yellow",
			background:    "#00ff00",
			foreground:    "#ffff00",
			interpolation: 0.5,
			expected:      "#80ff00",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Interpolate(tc.background, tc.foreground, tc.interpolation)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// TestInterpolateEdgeCases tests edge cases for Interpolate function
func TestInterpolateEdgeCases(t *testing.T) {
	testCases := []struct {
		name          string
		background    string
		foreground    string
		interpolation float64
		expected      string
		expectError   bool
	}{
		{
			name:          "identical colors",
			background:    "#ff0000",
			foreground:    "#ff0000",
			interpolation: 0.5,
			expected:      "#ff0000",
			expectError:   false,
		},
		{
			name:          "interpolation clamped below 0",
			background:    "#ff0000",
			foreground:    "#0000ff",
			interpolation: -1.0,
			expected:      "#ff0000", // should behave like 0.0
			expectError:   false,
		},
		{
			name:          "interpolation clamped above 1",
			background:    "#ff0000",
			foreground:    "#0000ff",
			interpolation: 2.0,
			expected:      "#0000ff", // should behave like 1.0
			expectError:   false,
		},
		{
			name:          "very close colors",
			background:    "#ff0000",
			foreground:    "#fe0000",
			interpolation: 0.5,
			expected:      "#ff0000",
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Interpolate(tc.background, tc.foreground, tc.interpolation)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

// TestInterpolateErrorHandling tests error cases for Interpolate function
func TestInterpolateErrorHandling(t *testing.T) {
	testCases := []struct {
		name          string
		background    string
		foreground    string
		interpolation float64
	}{
		{"invalid background (missing #)", "ff0000", "#00ff00", 0.5},
		{"invalid background (invalid chars)", "#gg0000", "#00ff00", 0.5},
		{"invalid background (too short)", "#f00", "#00ff00", 0.5},
		{"invalid foreground (missing #)", "#ff0000", "00ff00", 0.5},
		{"invalid foreground (invalid chars)", "#ff0000", "#gg0000", 0.5},
		{"invalid foreground (too short)", "#ff0000", "#f00", 0.5},
		{"empty background", "", "#00ff00", 0.5},
		{"empty foreground", "#ff0000", "", 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Interpolate(tc.background, tc.foreground, tc.interpolation)
			assert.Error(t, err)
		})
	}
}

// TestFadeFunctionality tests the fade function with normal cases
func TestFadeFunctionality(t *testing.T) {
	// Mock terminal info for deterministic testing
	termBg := "#000000" // black background
	termFg := "#ffffff" // white foreground
	colourMode := ansiParse.TrueColour

	t.Run("basic fade", func(t *testing.T) {
		result, err := fade("\x1b[31mRed text\x1b[0m", termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Red text")
	})

	t.Run("no fade (interpolation = 1.0)", func(t *testing.T) {
		result, err := fade("\x1b[31mRed text\x1b[0m", termBg, termFg, colourMode, 1.0)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Red text")
	})

	t.Run("full fade (interpolation = 0.0)", func(t *testing.T) {
		result, err := fade("\x1b[31mRed text\x1b[0m", termBg, termFg, colourMode, 0.0)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Red text")
	})

	t.Run("different color combinations", func(t *testing.T) {
		testCases := []struct {
			name    string
			content string
		}{
			{"red on green", "\x1b[31;42mRed on green\x1b[0m"},
			{"blue on yellow", "\x1b[34;43mBlue on yellow\x1b[0m"},
			{"magenta on cyan", "\x1b[35;46mMagenta on cyan\x1b[0m"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := fade(tc.content, termBg, termFg, colourMode, 0.5)
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			})
		}
	})

	t.Run("complex ANSI string", func(t *testing.T) {
		content := "\x1b[31mRed\x1b[32mGreen\x1b[33mYellow\x1b[0m"
		result, err := fade(content, termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Red")
		assert.Contains(t, result, "Green")
		assert.Contains(t, result, "Yellow")
	})
}

// TestFadeEdgeCases tests edge cases for fade function
func TestFadeEdgeCases(t *testing.T) {
	// Mock terminal info for deterministic testing
	termBg := "#000000"
	termFg := "#ffffff"
	colourMode := ansiParse.TrueColour

	for _, tc := range testANSIStrings {
		t.Run(tc.name, func(t *testing.T) {
			result, err := fade(tc.content, termBg, termFg, colourMode, 0.5)
			require.NoError(t, err)
			// Empty string input still returns ANSI codes (reset sequence)
			// so we just verify it doesn't error
			assert.NotEmpty(t, result)
		})
	}

	t.Run("unicode characters", func(t *testing.T) {
		content := "\x1b[31mHello 世界 🌍\x1b[0m"
		result, err := fade(content, termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("very long ANSI string", func(t *testing.T) {
		content := "\x1b[31m" + strings.Repeat("x", 1000) + "\x1b[0m"
		result, err := fade(content, termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})
}

// TestFadeErrorHandling tests error cases for fade function
func TestFadeErrorHandling(t *testing.T) {
	// Mock terminal info for deterministic testing
	termBg := "#000000"
	termFg := "#ffffff"
	colourMode := ansiParse.TrueColour

	// Negative interpolation should be clamped, not error
	result, err := fade("\x1b[31mRed text\x1b[0m", termBg, termFg, colourMode, -1.0)
	require.NoError(t, err)
	assert.NotEmpty(t, result)

	// Interpolation > 1 should be clamped, not error
	result, err = fade("\x1b[31mRed text\x1b[0m", termBg, termFg, colourMode, 2.0)
	require.NoError(t, err)
	assert.NotEmpty(t, result)
}

// TestIntegration tests complete color processing pipeline
func TestIntegration(t *testing.T) {
	// Mock terminal info for deterministic testing
	termBg := "#000000"
	termFg := "#ffffff"
	colourMode := ansiParse.TrueColour

	t.Run("basic pipeline", func(t *testing.T) {
		// Test a complete flow: ANSI string -> fade -> verify output
		content := "\x1b[31mRed text\x1b[0m"
		result, err := fade(content, termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "Red text")
	})

	t.Run("multiple fade operations", func(t *testing.T) {
		content := "\x1b[31mRed text\x1b[0m"

		// First fade
		result1, err := fade(content, termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)

		// Second fade on the result
		result2, err := fade(result1, termBg, termFg, colourMode, 0.5)
		require.NoError(t, err)
		assert.NotEmpty(t, result2)
	})
}

// BenchmarkFade benchmarks the fade function
func BenchmarkFade(b *testing.B) {
	// Mock terminal info for deterministic benchmarking
	termBg := "#000000"
	termFg := "#ffffff"
	colourMode := ansiParse.TrueColour

	content := "\x1b[31mRed text\x1b[32mGreen text\x1b[33mYellow text\x1b[0m"

	for b.Loop() {
		_, _ = fade(content, termBg, termFg, colourMode, 0.5)
	}
}

// BenchmarkInterpolate benchmarks the Interpolate function
func BenchmarkInterpolate(b *testing.B) {
	background := "#ff0000"
	foreground := "#0000ff"

	for b.Loop() {
		_, _ = Interpolate(background, foreground, 0.5)
	}
}
