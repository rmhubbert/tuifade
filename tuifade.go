// Package tuifade provides functions for fading the background and foreground colours of an ANSI
// string.
package tuifade

import (
	"errors"
	"fmt"
	"math"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

type rbgColour = ansiParse.Rgb
type hslColour = ansiParse.Hsl

type InterpolationResult struct {
	Hex string
	Rgb rbgColour
	Hsl hslColour
}

type Interpolation struct {
	OriginalForeground string
	OriginalBackground string
	Interpolated       float64
	Result             InterpolationResult
}

// Fade fades the background and foreground colours of an ANSI string.
//
// If no background colour is specified, the default background colour is used. If no foreground
// colour is specified, the default foreground colour is used. The interpolation parameter
// controls the degree of fade. A value of 1 will result in no fade, while a value of 0
// will result in a fully faded string.
//
// If the current terminal does not support truecolor, the original content, plus an error is
// returned.
func Fade(content string, interpolation float64) (string, error) {
	termOutput := termenv.DefaultOutput()
	profile := termOutput.EnvColorProfile()

	if profile != termenv.TrueColor {
		return content, errors.New("fade only supports truecolor terminals")
	}

	termBg := fmt.Sprintf("%s", termOutput.BackgroundColor())
	termFg := fmt.Sprintf("%s", termOutput.ForegroundColor())
	colourMode := colourModeFromProfile(profile)

	return fade(content, termBg, termFg, colourMode, interpolation)
}

// fade fades the background and foreground colours of an ANSI string.
func fade(
	content, termBg, termFg string,
	colourMode ansiParse.ColourMode,
	interpolationAmount float64,
) (string, error) {

	// Parse the input string into segments
	originalSegments, _ := ansiParse.Parse(
		content,
		ansiParse.WithDefaultBackgroundColor(termBg),
		ansiParse.WithDefaultForegroundColor(termFg),
	)
	fadedSegments := []*ansiParse.StyledText{}

	// Iterate over each segment and fade the background and foreground colours
	for _, originalSegment := range originalSegments {
		segment := &ansiParse.StyledText{}
		segment.Label = originalSegment.Label
		segment.Style = originalSegment.Style
		segment.ColourMode = colourMode
		segment.Offset = originalSegment.Offset
		segment.Len = originalSegment.Len
		if originalSegment.FgCol != nil {
			segment.FgCol = originalSegment.FgCol
		}
		if originalSegment.BgCol != nil {
			segment.BgCol = originalSegment.BgCol
		}

		bgCol := termBg
		// If the background colour is set, fade it
		if originalSegment.BgCol != nil && originalSegment.BgCol.Hex != "" {
			if originalSegment.BgCol.Hex != termBg {
				var err error
				interpolation, err := Interpolate(
					bgCol,
					originalSegment.BgCol.Hex,
					interpolationAmount,
				)
				if err != nil {
					return "", err
				}
				err = updateSegmentBackgroundColours(segment, interpolation.Result)
				if err != nil {
					return "", err
				}
				bgCol = interpolation.Result.Hex
			}
		}
		// If the foreground colour is set, fade it
		if originalSegment.FgCol != nil && originalSegment.FgCol.Hex != "" {
			interpolation, err := Interpolate(bgCol, originalSegment.FgCol.Hex, interpolationAmount)
			if err != nil {
				return "", err
			}

			err = updateSegmentForegroundColours(segment, interpolation.Result)
			if err != nil {
				return "", err
			}
		} else { // If the foreground colour is not set, use the default foreground colour
			interpolation, err := Interpolate(bgCol, termFg, interpolationAmount)
			if err != nil {
				return "", err
			}

			err = updateSegmentForegroundColours(segment, interpolation.Result)
			if err != nil {
				return "", err
			}
		}
		fadedSegments = append(fadedSegments, segment)
	}
	return ansiParse.String(fadedSegments), nil
}

// updateSegmentForegroundColours updates the foreground colours of a segment.
func updateSegmentForegroundColours(
	segment *ansiParse.StyledText,
	colours InterpolationResult,
) error {
	if segment.FgCol == nil {
		segment.FgCol = &ansiParse.Col{}
	}

	segment.FgCol = &ansiParse.Col{
		Id:   segment.FgCol.Id,
		Name: segment.FgCol.Name,
		Hex:  colours.Hex,
		Rgb: ansiParse.Rgb{
			R: colours.Rgb.R,
			G: colours.Rgb.G,
			B: colours.Rgb.B,
		},
		Hsl: ansiParse.Hsl{
			H: colours.Hsl.H,
			S: colours.Hsl.S,
			L: colours.Hsl.L,
		},
	}

	return nil
}

// updateSegment updates the background colours of a segment. It will do nothing if the segment
// has no background colour.
func updateSegmentBackgroundColours(
	segment *ansiParse.StyledText,
	colours InterpolationResult,
) error {
	if segment.BgCol == nil {
		return nil
	}

	segment.BgCol = &ansiParse.Col{
		Id:   segment.FgCol.Id,
		Name: segment.FgCol.Name,
		Hex:  colours.Hex,
		Rgb: ansiParse.Rgb{
			R: colours.Rgb.R,
			G: colours.Rgb.G,
			B: colours.Rgb.B,
		},
		Hsl: ansiParse.Hsl{
			H: colours.Hsl.H,
			S: colours.Hsl.S,
			L: colours.Hsl.L,
		},
	}

	return nil
}

// colourModeFromProfile returns the appropriate ansiParse.ColourMode based on the given
// termenv profile.
func colourModeFromProfile(profile termenv.Profile) ansiParse.ColourMode {
	if profile == termenv.TrueColor {
		return ansiParse.TrueColour
	}
	if profile == termenv.ANSI256 {
		return ansiParse.TwoFiveSix
	}
	return ansiParse.Default
}

// Interpolate interpolates the background and foreground colours of an ANSI string.
//
// The interpolation parameter controls the degree of fade. A value of 1 will result in no fade,
// while a value of 0 will result in a fully faded string.
func Interpolate(
	hexBackground, hexForeground string,
	interpolation float64,
) (Interpolation, error) {
	result := Interpolation{
		OriginalForeground: hexForeground,
		OriginalBackground: hexBackground,
		Interpolated:       interpolation,
		Result:             InterpolationResult{},
	}

	background, err := hexToRGB(hexBackground)
	if err != nil {
		return result, err
	}
	foreground, err := hexToRGB(hexForeground)
	if err != nil {
		return result, err
	}

	// Clamp interpolation value to valid range [0, 1]
	if interpolation < 0 {
		interpolation = 0
	} else if interpolation > 1 {
		interpolation = 1
	}

	// Calculate interpolation weights
	bgWeight := 1 - interpolation
	fgWeight := interpolation
	// Interpolate each RGB channel
	r := interpolateChannel(background.R, foreground.R, bgWeight, fgWeight)
	g := interpolateChannel(background.G, foreground.G, bgWeight, fgWeight)
	b := interpolateChannel(background.B, foreground.B, bgWeight, fgWeight)

	result.Result.Hex = rgbToHex(rbgColour{R: r, G: g, B: b})
	result.Result.Rgb = rbgColour{R: r, G: g, B: b}
	h, s, l := rgbToHSL(rbgColour{R: r, G: g, B: b})
	result.Result.Hsl = hslColour{H: h, S: s, L: l}

	return result, nil
}

// interpolateChannel performs linear interpolation for a single colour channel.
func interpolateChannel(bg, fg uint8, bgWeight, fgWeight float64) uint8 {
	bgValue := float64(bg)
	fgValue := float64(fg)
	result := bgValue*bgWeight + fgValue*fgWeight
	return uint8(math.Round(result))
}

// rgbToHex converts an rbgColour to a hex string.
func rgbToHex(rgb rbgColour) string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

// hexToRGB converts a hex string to an rbgColour.
func hexToRGB(hex string) (rbgColour, error) {
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return rbgColour{}, err
	}
	return rbgColour{R: r, G: g, B: b}, nil
}

// rgbToHSL converts an rbgColour to HSL without re-parsing hex string.
func rgbToHSL(rgb rbgColour) (h, s, l float64) {
	// Create colorful.Color from RGB values (normalized to 0.0-1.0 range)
	c := colorful.LinearRgb(
		float64(rgb.R)/255.0,
		float64(rgb.G)/255.0,
		float64(rgb.B)/255.0,
	)

	// Get HSL values (H: 0-360, S: 0-1, L: 0-1)
	return c.Hsl()
}

// // hexToHSL converts a hex colour string to HSL.
// func hexToHSL(hex string) (hslColour, error) {
// 	rgb, err := globalColourCache.getRGB(hex)
// 	if err != nil {
// 		return hslColour{}, err
// 	}
//
// 	// Create colorful.Color from RGB values (normalized to 0.0-1.0 range)
// 	c := colorful.LinearRgb(
// 		float64(rgb.R)/255.0,
// 		float64(rgb.G)/255.0,
// 		float64(rgb.B)/255.0,
// 	)
//
// 	// Get HSL values (H: 0-360, S: 0-1, L: 0-1)
// 	h, s, l := c.Hsl()
//
// 	// Convert to hslColour type (H: 0-360, S: 0-100, L: 0-100)
// 	return hslColour{
// 		H: h * 360.0,
// 		S: s * 100.0,
// 		L: l * 100.0,
// 	}, nil
// }
