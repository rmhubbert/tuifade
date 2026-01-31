// Package tuifade provides functions for fading the background and foreground colours of an ANSI
// string.
package tuifade

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"sync"

	ansiParse "github.com/leaanthony/go-ansi-parser"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/termenv"
)

type rbgColour = ansiParse.Rgb
type hslColour = ansiParse.Hsl

type Fader struct {
	colourCache  sync.Map
	contentCache sync.Map
	termInfo     *termInfo
}

type termInfo struct {
	defaultOutput           *termenv.Output
	defaultBackgroundColour string
	defaultForegroundColour string
	profile                 termenv.Profile
}

type InterpolationResult struct {
	hex string
	rgb rbgColour
	hsl hslColour
}

type Interpolation struct {
	originalForeground string
	originalBackground string
	interpolated       float64
	result             InterpolationResult
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
func (f *Fader) Fade(content string, interpolation float64) (string, error) {
	termInfo := f.generateTermInfo()
	if termInfo.profile != termenv.TrueColor {
		return content, errors.New("fade only supports truecolor terminals")
	}

	contentCacheKey := generateContentCacheKey(content)
	if cached, ok := f.contentCache.Load(contentCacheKey); ok {
		return cached.(string), nil
	}

	i := clampInterpolation(interpolation)
	if i == 1 {
		return content, nil
	}

	faded, err := f.fade(
		content,
		termInfo.defaultBackgroundColour,
		termInfo.defaultForegroundColour,
		colourModeFromProfile(termInfo.profile),
		i,
	)
	if err != nil {
		return "", err
	}
	f.contentCache.Store(contentCacheKey, faded)
	return faded, nil
}

// fade fades the background and foreground colours of an ANSI string.
func (f *Fader) fade(
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
				interpolation, err := f.Interpolate(
					bgCol,
					originalSegment.BgCol.Hex,
					interpolationAmount,
				)
				if err != nil {
					return "", err
				}
				err = updateSegmentBackgroundColours(segment, interpolation.result)
				if err != nil {
					return "", err
				}
				bgCol = interpolation.result.hex
			}
		}
		// If the foreground colour is set, fade it
		if originalSegment.FgCol != nil && originalSegment.FgCol.Hex != "" {
			interpolation, err := f.Interpolate(
				bgCol,
				originalSegment.FgCol.Hex,
				interpolationAmount,
			)
			if err != nil {
				return "", err
			}
			err = updateSegmentForegroundColours(segment, interpolation.result)
			if err != nil {
				return "", err
			}
		} else { // If the foreground colour is not set, use the default foreground colour
			interpolation, err := f.Interpolate(bgCol, termFg, interpolationAmount)
			if err != nil {
				return "", err
			}
			err = updateSegmentForegroundColours(segment, interpolation.result)
			if err != nil {
				return "", err
			}
		}
		fadedSegments = append(fadedSegments, segment)
	}
	return ansiParse.String(fadedSegments), nil
}

// Interpolate interpolates the background and foreground colours of an ANSI string.
//
// The interpolation parameter controls the degree of fade. A value of 1 will result in no fade,
// while a value of 0 will result in a fully faded string.
func (f *Fader) Interpolate(
	hexBackground, hexForeground string,
	interpolation float64,
) (Interpolation, error) {
	cacheKey := generateColourCacheKey(hexBackground, hexForeground, interpolation)
	if cached, ok := f.colourCache.Load(cacheKey); ok {
		return cached.(Interpolation), nil
	}

	result := Interpolation{
		originalForeground: hexForeground,
		originalBackground: hexBackground,
		interpolated:       interpolation,
		result:             InterpolationResult{},
	}

	background, err := hexToRGB(hexBackground)
	if err != nil {
		return result, err
	}
	foreground, err := hexToRGB(hexForeground)
	if err != nil {
		return result, err
	}

	// Calculate interpolation weights
	bgWeight := 1 - interpolation
	fgWeight := interpolation
	// Interpolate each RGB channel
	r := interpolateChannel(background.R, foreground.R, bgWeight, fgWeight)
	g := interpolateChannel(background.G, foreground.G, bgWeight, fgWeight)
	b := interpolateChannel(background.B, foreground.B, bgWeight, fgWeight)

	result.result.hex = rgbToHex(rbgColour{R: r, G: g, B: b})
	result.result.rgb = rbgColour{R: r, G: g, B: b}
	result.result.hsl = rgbToHSL(rbgColour{R: r, G: g, B: b})

	f.colourCache.Store(cacheKey, result)

	return result, nil
}

func (f *Fader) generateTermInfo() *termInfo {
	if f.termInfo != nil {
		return f.termInfo
	}

	termOutput := termenv.DefaultOutput()
	profile := termOutput.EnvColorProfile()
	return &termInfo{
		defaultOutput:           termOutput,
		defaultBackgroundColour: fmt.Sprintf("%s", termOutput.BackgroundColor()),
		defaultForegroundColour: fmt.Sprintf("%s", termOutput.ForegroundColor()),
		profile:                 profile,
	}
}

// generateColourCacheKey generates a cache key for the given parameters.
func generateColourCacheKey(termBg, termFg string, interpolation float64) string {
	return fmt.Sprintf("%s-%s-%f", termBg, termFg, interpolation)
}

func generateContentCacheKey(content string) string {
	return base64.StdEncoding.EncodeToString([]byte(content))
}

// clampInterpolation clamps the interpolation value to the range [0, 1].
func clampInterpolation(interpolation float64) float64 {
	if interpolation < 0 {
		return 0
	} else if interpolation > 1 {
		return 1
	}
	return interpolation
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
		Hex:  colours.hex,
		Rgb: ansiParse.Rgb{
			R: colours.rgb.R,
			G: colours.rgb.G,
			B: colours.rgb.B,
		},
		Hsl: ansiParse.Hsl{
			H: colours.hsl.H,
			S: colours.hsl.S,
			L: colours.hsl.L,
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
		Id:   segment.BgCol.Id,
		Name: segment.BgCol.Name,
		Hex:  colours.hex,
		Rgb: ansiParse.Rgb{
			R: colours.rgb.R,
			G: colours.rgb.G,
			B: colours.rgb.B,
		},
		Hsl: ansiParse.Hsl{
			H: colours.hsl.H,
			S: colours.hsl.S,
			L: colours.hsl.L,
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
func rgbToHSL(rgb rbgColour) hslColour {
	// Create colorful.Color from RGB values (normalized to 0.0-1.0 range)
	c := colorful.LinearRgb(
		float64(rgb.R)/255.0,
		float64(rgb.G)/255.0,
		float64(rgb.B)/255.0,
	)

	h, s, l := c.Hsl()
	return hslColour{
		H: h,
		S: s,
		L: l,
	}
}
