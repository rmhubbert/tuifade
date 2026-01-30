package tuifade

import (
	"math"
	"testing"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	termenv "github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertAlmostEqual checks if two floats are close, using a tolerance of 0.01 for RGB/Hex values
func assertAlmostEqual(t *testing.T, a, b float64) {
	t.Helper()
	// Use 0.01 as a reasonable tolerance for RGB values (0-255 range)
	tolerance := 0.01
	delta := math.Abs(a - b)
	if delta > tolerance {
		t.Errorf("Values close: %f (expected %f)", a, b)
	}
}

// TestRgbToHex tests the rgbToHex conversion function
func TestRgbToHex(t *testing.T) {
	tests := []struct {
		name     string
		rgb      rbgColour
		expected string
	}{
		{
			name:     "Pure Red",
			rgb:      rbgColour{R: 255, G: 0, B: 0},
			expected: "#ff0000",
		},
		{
			name:     "Pure Green",
			rgb:      rbgColour{R: 0, G: 255, B: 0},
			expected: "#00ff00",
		},
		{
			name:     "Pure Blue",
			rgb:      rbgColour{R: 0, G: 0, B: 255},
			expected: "#0000ff",
		},
		{
			name:     "Pure White",
			rgb:      rbgColour{R: 255, G: 255, B: 255},
			expected: "#ffffff",
		},
		{
			name:     "Pure Black",
			rgb:      rbgColour{R: 0, G: 0, B: 0},
			expected: "#000000",
		},
		{
			name:     "Mixed Color",
			rgb:      rbgColour{R: 128, G: 128, B: 128},
			expected: "#808080",
		},
		{
			name:     "Low Brightness",
			rgb:      rbgColour{R: 10, G: 10, B: 10},
			expected: "#0a0a0a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rgbToHex(tt.rgb)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHexToRGB tests the hexToRGB conversion function
func TestHexToRGB(t *testing.T) {
	tests := []struct {
		name      string
		hex       string
		expected  rbgColour
		wantError bool
	}{
		{
			name:      "Valid hex with # prefix",
			hex:       "#ff0000",
			expected:  rbgColour{R: 255, G: 0, B: 0},
			wantError: false,
		},
		{
			name:      "Valid hex without # prefix",
			hex:       "ff0000",
			expected:  rbgColour{R: 255, G: 0, B: 0},
			wantError: true, // hexToRGB expects # prefix
		},
		{
			name:      "Pure White",
			hex:       "#ffffff",
			expected:  rbgColour{R: 255, G: 255, B: 255},
			wantError: false,
		},
		{
			name:      "Pure Black",
			hex:       "#000000",
			expected:  rbgColour{R: 0, G: 0, B: 0},
			wantError: false,
		},
		{
			name:      "Invalid hex - too short",
			hex:       "#f00",
			expected:  rbgColour{},
			wantError: true,
		},
		{
			name:      "Invalid hex - non-hex characters",
			hex:       "#gggggg",
			expected:  rbgColour{},
			wantError: true,
		},
		{
			name:      "Empty string",
			hex:       "",
			expected:  rbgColour{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hexToRGB(tt.hex)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// TestRgbToHSL tests the rgbToHSL conversion function
func TestRgbToHSL(t *testing.T) {
	tests := []struct {
		name     string
		rgb      rbgColour
		expected struct {
			h, s, l float64
		}
	}{
		{
			name:     "Pure Red",
			rgb:      rbgColour{R: 255, G: 0, B: 0},
			expected: struct{ h, s, l float64 }{h: 0, s: 1.0, l: 0.5},
		},
		{
			name:     "Pure Green",
			rgb:      rbgColour{R: 0, G: 255, B: 0},
			expected: struct{ h, s, l float64 }{h: 120, s: 1.0, l: 0.5},
		},
		{
			name:     "Pure Blue",
			rgb:      rbgColour{R: 0, G: 0, B: 255},
			expected: struct{ h, s, l float64 }{h: 240, s: 1.0, l: 0.5},
		},
		{
			name:     "Pure White",
			rgb:      rbgColour{R: 255, G: 255, B: 255},
			expected: struct{ h, s, l float64 }{h: 0, s: 0, l: 1.0},
		},
		{
			name:     "Pure Black",
			rgb:      rbgColour{R: 0, G: 0, B: 0},
			expected: struct{ h, s, l float64 }{h: 0, s: 0, l: 0.0},
		},
		{
			name:     "Mixed Color",
			rgb:      rbgColour{R: 128, G: 128, B: 128},
			expected: struct{ h, s, l float64 }{h: 0, s: 0, l: 0.736647},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s, l := rgbToHSL(tt.rgb)

			// Use tolerance for floating-point comparison
			assertAlmostEqual(t, h, tt.expected.h)
			assertAlmostEqual(t, s, tt.expected.s)
			assertAlmostEqual(t, l, tt.expected.l)
		})
	}
}

// TestInterpolateChannel tests the interpolateChannel function
func TestInterpolateChannel(t *testing.T) {
	tests := []struct {
		name     string
		bg, fg   uint8
		bgWeight float64
		fgWeight float64
		expected uint8
	}{
		{
			name:     "Midpoint interpolation",
			bg:       0,
			fg:       255,
			bgWeight: 0.5,
			fgWeight: 0.5,
			expected: 128,
		},
		{
			name:     "Full background",
			bg:       255,
			fg:       0,
			bgWeight: 1.0,
			fgWeight: 0.0,
			expected: 255,
		},
		{
			name:     "Full foreground",
			bg:       0,
			fg:       255,
			bgWeight: 0.0,
			fgWeight: 1.0,
			expected: 255,
		},
		{
			name:     "75% background, 25% foreground",
			bg:       255,
			fg:       0,
			bgWeight: 0.75,
			fgWeight: 0.25,
			expected: 191,
		},
		{
			name:     "25% background, 75% foreground",
			bg:       0,
			fg:       255,
			bgWeight: 0.25,
			fgWeight: 0.75,
			expected: 191,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpolateChannel(tt.bg, tt.fg, tt.bgWeight, tt.fgWeight)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestInterpolate tests the Interpolate function
func TestInterpolate(t *testing.T) {
	tests := []struct {
		name          string
		bgHex, fgHex  string
		interpolation float64
		expectedHex   string
		wantError     bool
		errorMsg      string
	}{
		{
			name:          "Midpoint interpolation",
			bgHex:         "#ff0000",
			fgHex:         "#0000ff",
			interpolation: 0.5,
			expectedHex:   "#800080",
			wantError:     false,
		},
		{
			name:          "No fade (interpolation = 1.0)",
			bgHex:         "#ff0000",
			fgHex:         "#0000ff",
			interpolation: 1.0,
			expectedHex:   "#0000ff",
			wantError:     false,
		},
		{
			name:          "Full fade (interpolation = 0.0)",
			bgHex:         "#ff0000",
			fgHex:         "#0000ff",
			interpolation: 0.0,
			expectedHex:   "#ff0000",
			wantError:     false,
		},
		{
			name:          "25% interpolation",
			bgHex:         "#000000",
			fgHex:         "#ffffff",
			interpolation: 0.25,
			expectedHex:   "#404040",
			wantError:     false,
		},
		{
			name:          "75% interpolation",
			bgHex:         "#000000",
			fgHex:         "#ffffff",
			interpolation: 0.75,
			expectedHex:   "#bfbfbf",
			wantError:     false,
		},
		{
			name:          "Invalid hex - non-hex characters",
			bgHex:         "#gggggg",
			fgHex:         "#ffffff",
			interpolation: 0.5,
			wantError:     true,
			errorMsg:      "expected integer",
		},
		{
			name:          "Invalid hex - too short",
			bgHex:         "#f00",
			fgHex:         "#ffffff",
			interpolation: 0.5,
			wantError:     true,
			errorMsg:      "EOF",
		},
		{
			name:          "Empty background hex",
			bgHex:         "",
			fgHex:         "#ffffff",
			interpolation: 0.5,
			wantError:     true,
			errorMsg:      "unexpected EOF",
		},
		{
			name:          "Interpolation value > 1 (should clamp to 1)",
			bgHex:         "#ff0000",
			fgHex:         "#0000ff",
			interpolation: 1.5,
			expectedHex:   "#0000ff",
			wantError:     false,
		},
		{
			name:          "Interpolation value < 0 (should clamp to 0)",
			bgHex:         "#ff0000",
			fgHex:         "#0000ff",
			interpolation: -0.5,
			expectedHex:   "#ff0000",
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Interpolate(tt.bgHex, tt.fgHex, tt.interpolation)

			if tt.wantError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedHex, result.result.hex)

				// Verify RGB values
				// bgRgb, _ := hexToRGB(tt.bgHex)
				// fgRgb, _ := hexToRGB(tt.fgHex)
				// Calculate expected RGB using interpolation
				bgRgb, bgErr := hexToRGB(tt.bgHex)
				fgRgb, fgErr := hexToRGB(tt.fgHex)
				if bgErr != nil || fgErr != nil {
					t.Fatal("Failed to parse hex colors for interpolation test")
				}

				// Calculate expected RGB using interpolation (with clamping)
				interpolation := tt.interpolation
				if interpolation < 0 {
					interpolation = 0
				} else if interpolation > 1 {
					interpolation = 1
				}
				bgWeight := 1.0 - interpolation
				fgWeight := interpolation
				expectedR := interpolateChannel(bgRgb.R, fgRgb.R, bgWeight, fgWeight)
				expectedG := interpolateChannel(bgRgb.G, fgRgb.G, bgWeight, fgWeight)
				expectedB := interpolateChannel(bgRgb.B, fgRgb.B, bgWeight, fgWeight)

				assert.Equal(t, expectedR, result.result.rgb.R)
				assert.Equal(t, expectedG, result.result.rgb.G)
				assert.Equal(t, expectedB, result.result.rgb.B)

				// Verify HSL values are calculated correctly
				h, s, l := rgbToHSL(result.result.rgb)
				assertAlmostEqual(t, h, result.result.hsl.H)
				assertAlmostEqual(t, s, result.result.hsl.S)
				assertAlmostEqual(t, l, result.result.hsl.L)
			}
		})
	}
}

// TestUpdateSegmentForegroundColours tests the updateSegmentForegroundColours function
func TestUpdateSegmentForegroundColours(t *testing.T) {
	tests := []struct {
		name     string
		segment  *ansiParse.StyledText
		colours  interpolationResult
		expected *ansiParse.StyledText
	}{
		{
			name: "Segment with existing foreground color",
			segment: &ansiParse.StyledText{
				FgCol: &ansiParse.Col{
					Hex: "#ff0000",
					Rgb: ansiParse.Rgb{R: 255, G: 0, B: 0},
				},
			},
			colours: interpolationResult{
				hex: "#00ff00",
				rgb: rbgColour{R: 0, G: 255, B: 0},
				hsl: hslColour{H: 120, S: 1.0, L: 0.5},
			},
			expected: &ansiParse.StyledText{
				FgCol: &ansiParse.Col{
					Hex: "#00ff00",
					Rgb: ansiParse.Rgb{R: 0, G: 255, B: 0},
					Hsl: ansiParse.Hsl{H: 120, S: 1.0, L: 0.5},
				},
			},
		},
		{
			name: "Segment without foreground color",
			segment: &ansiParse.StyledText{
				FgCol: nil,
			},
			colours: interpolationResult{
				hex: "#0000ff",
				rgb: rbgColour{R: 0, G: 0, B: 255},
				hsl: hslColour{H: 240, S: 1.0, L: 0.5},
			},
			expected: &ansiParse.StyledText{
				FgCol: &ansiParse.Col{
					Hex: "#0000ff",
					Rgb: ansiParse.Rgb{R: 0, G: 0, B: 255},
					Hsl: ansiParse.Hsl{H: 240, S: 1.0, L: 0.5},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the segment to avoid modifying the original
			segmentCopy := &ansiParse.StyledText{}
			if tt.segment.FgCol != nil {
				segmentCopy.FgCol = &ansiParse.Col{
					Id:   tt.segment.FgCol.Id,
					Name: tt.segment.FgCol.Name,
					Hex:  tt.segment.FgCol.Hex,
					Rgb:  tt.segment.FgCol.Rgb,
					Hsl:  tt.segment.FgCol.Hsl,
				}
			}

			err := updateSegmentForegroundColours(segmentCopy, tt.colours)
			require.NoError(t, err)

			assert.Equal(t, tt.expected.FgCol.Hex, segmentCopy.FgCol.Hex)
			assert.Equal(t, tt.expected.FgCol.Rgb.R, segmentCopy.FgCol.Rgb.R)
			assert.Equal(t, tt.expected.FgCol.Rgb.G, segmentCopy.FgCol.Rgb.G)
			assert.Equal(t, tt.expected.FgCol.Rgb.B, segmentCopy.FgCol.Rgb.B)
			assertAlmostEqual(t, tt.expected.FgCol.Hsl.H, segmentCopy.FgCol.Hsl.H)
			assertAlmostEqual(t, tt.expected.FgCol.Hsl.S, segmentCopy.FgCol.Hsl.S)
			assertAlmostEqual(t, tt.expected.FgCol.Hsl.L, segmentCopy.FgCol.Hsl.L)
		})
	}
}

// TestUpdateSegmentBackgroundColours tests the updateSegmentBackgroundColours function
func TestUpdateSegmentBackgroundColours(t *testing.T) {
	tests := []struct {
		name      string
		segment   *ansiParse.StyledText
		colours   interpolationResult
		expected  *ansiParse.StyledText
		wantError bool
	}{
		{
			name: "Segment with existing background color",
			segment: &ansiParse.StyledText{
				BgCol: &ansiParse.Col{
					Hex: "#ff0000",
					Rgb: ansiParse.Rgb{R: 255, G: 0, B: 0},
				},
			},
			colours: interpolationResult{
				hex: "#00ff00",
				rgb: rbgColour{R: 0, G: 255, B: 0},
				hsl: hslColour{H: 120, S: 1.0, L: 0.5},
			},
			expected: &ansiParse.StyledText{
				BgCol: &ansiParse.Col{
					Hex: "#00ff00",
					Rgb: ansiParse.Rgb{R: 0, G: 255, B: 0},
					Hsl: ansiParse.Hsl{H: 120, S: 1.0, L: 0.5},
				},
			},
			wantError: false,
		},
		{
			name: "Segment without background color",
			segment: &ansiParse.StyledText{
				BgCol: nil,
			},
			colours: interpolationResult{
				hex: "#0000ff",
				rgb: rbgColour{R: 0, G: 0, B: 255},
				hsl: hslColour{H: 240, S: 1.0, L: 0.5},
			},
			expected: &ansiParse.StyledText{
				BgCol: nil,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the segment to avoid modifying the original
			segmentCopy := &ansiParse.StyledText{}
			if tt.segment.BgCol != nil {
				segmentCopy.BgCol = &ansiParse.Col{
					Id:   tt.segment.BgCol.Id,
					Name: tt.segment.BgCol.Name,
					Hex:  tt.segment.BgCol.Hex,
					Rgb:  tt.segment.BgCol.Rgb,
					Hsl:  tt.segment.BgCol.Hsl,
				}
			}

			err := updateSegmentBackgroundColours(segmentCopy, tt.colours)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.expected.BgCol != nil {
					assert.Equal(t, tt.expected.BgCol.Hex, segmentCopy.BgCol.Hex)
					assert.Equal(t, tt.expected.BgCol.Rgb.R, segmentCopy.BgCol.Rgb.R)
					assert.Equal(t, tt.expected.BgCol.Rgb.G, segmentCopy.BgCol.Rgb.G)
					assert.Equal(t, tt.expected.BgCol.Rgb.B, segmentCopy.BgCol.Rgb.B)
					assertAlmostEqual(t, tt.expected.BgCol.Hsl.H, segmentCopy.BgCol.Hsl.H)
					assertAlmostEqual(t, tt.expected.BgCol.Hsl.S, segmentCopy.BgCol.Hsl.S)
					assertAlmostEqual(t, tt.expected.BgCol.Hsl.L, segmentCopy.BgCol.Hsl.L)
				}
			}
		})
	}
}

// TestColourModeFromProfile tests the colourModeFromProfile function
func TestColourModeFromProfile(t *testing.T) {
	tests := []struct {
		name     string
		profile  termenv.Profile
		expected ansiParse.ColourMode
	}{
		{
			name:     "TrueColor profile",
			profile:  termenv.TrueColor,
			expected: ansiParse.TrueColour,
		},
		{
			name:     "ANSI256 profile",
			profile:  termenv.ANSI256,
			expected: ansiParse.TwoFiveSix,
		},
		{
			name:     "ANSI profile",
			profile:  termenv.ANSI,
			expected: ansiParse.Default,
		},
		{
			name:     "ASCII profile",
			profile:  termenv.Ascii,
			expected: ansiParse.Default,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := colourModeFromProfile(tt.profile)
			assert.Equal(t, tt.expected, result)
		})
	}
}
