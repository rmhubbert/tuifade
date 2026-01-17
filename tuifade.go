// Package tuifade provides functions for fading the background and foreground colours of an ANSI
// string.
package tuifade

import (
	"errors"
	"fmt"
	"math"
	"strings"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

type rbgColour = ansiParse.Rgb
type hslColour = ansiParse.Hsl

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
	interpolation float64,
) (string, error) {

	// Parse the input string into segments
	parsed, _ := ansiParse.Parse(content)
	builder := strings.Builder{}

	// Iterate over each segment and fade the background and foreground colours
	for _, segment := range parsed {
		// Set the colour mode based on the current profile
		segment.ColourMode = colourMode
		bgCol := termBg
		var fgCol string

		// If the background colour is set, fade it
		if segment.BgCol != nil && segment.BgCol.Hex != "" {
			if segment.BgCol.Hex != termBg {
				var err error
				bgCol, err = Interpolate(bgCol, segment.BgCol.Hex, interpolation)
				if err != nil {
					return "", err
				}
			}
		}

		// If the foreground colour is set, fade it
		if segment.FgCol != nil && segment.FgCol.Hex != "" {
			var err error
			fgCol, err = Interpolate(bgCol, segment.FgCol.Hex, interpolation)
			if err != nil {
				return "", err
			}
		} else { // If the foreground colour is not set, use the default foreground colour
			if segment.FgCol == nil {
				segment.FgCol = &ansiParse.Col{}
			}

			var err error
			fgCol, err = Interpolate(bgCol, termFg, interpolation)
			if err != nil {
				return "", err
			}

		}

		updateSegmentColours(segment, bgCol, fgCol)
		builder.WriteString(segment.String())
	}
	return builder.String(), nil
}

// updateSegment updates the background and foreground colours of a segment.
func updateSegmentColours(segment *ansiParse.StyledText, bgCol, fgCol string) {
	if segment.BgCol == nil {
		segment.BgCol = &ansiParse.Col{}
	}
	if segment.FgCol == nil {
		segment.FgCol = &ansiParse.Col{}
	}

	segment.BgCol.Hex = bgCol
	bgRgb, err := hexToRGB(bgCol)
	if err != nil {
		return
	}
	segment.BgCol.Rgb = bgRgb
	bgHsl, err := hexToHSL(bgCol)
	if err != nil {
		return
	}
	segment.BgCol.Hsl = bgHsl

	segment.FgCol.Hex = fgCol
	fgRgb, err := hexToRGB(fgCol)
	if err != nil {
		return
	}
	segment.FgCol.Rgb = fgRgb
	fgHsl, err := hexToHSL(fgCol)
	if err != nil {
		return
	}
	segment.FgCol.Hsl = fgHsl
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
func Interpolate(hexBackground, hexForeground string, interpolation float64) (string, error) {
	background, err := hexToRGB(hexBackground)
	if err != nil {
		return "", err
	}
	foreground, err := hexToRGB(hexForeground)
	if err != nil {
		return "", err
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

	return rgbToHex(rbgColour{R: r, G: g, B: b}), nil
}

// interpolateChannel performs linear interpolation for a single color channel.
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

// hexToHSL converts a hex color string to HSL.
func hexToHSL(hex string) (hslColour, error) {
	// First convert hex to RGB using existing function
	rgb, err := hexToRGB(hex)
	if err != nil {
		return hslColour{}, err
	}

	// Create colorful.Color from RGB values (normalized to 0.0-1.0 range)
	c := colorful.LinearRgb(
		float64(rgb.R)/255.0,
		float64(rgb.G)/255.0,
		float64(rgb.B)/255.0,
	)

	// Get HSL values (H: 0-360, S: 0-1, L: 0-1)
	h, s, l := c.Hsl()

	// Convert to hslColour type (H: 0-360, S: 0-100, L: 0-100)
	return hslColour{
		H: h * 360.0,
		S: s * 100.0,
		L: l * 100.0,
	}, nil
}
