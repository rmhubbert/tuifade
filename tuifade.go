// Package tuifade provides functions for fading the background and foreground colours of an ANSI
// string.
package tuifade

import (
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

// colourCache provides thread-safe caching of colour conversions
type colourCache struct {
	rgb map[string]rbgColour
	hsl map[string]hslColour
	mu  sync.RWMutex
}

// global cache instance
var globalColourCache = &colourCache{
	rgb: make(map[string]rbgColour),
	hsl: make(map[string]hslColour),
}

// interpolationCache stores computed interpolation results internally
type interpolationCache struct {
	cache map[string]string
	mu    sync.RWMutex
}

// Global instance of the interpolation cache
var globalInterpolationCache = &interpolationCache{
	cache: make(map[string]string),
}

// generateCacheKey creates a simple key for interpolation caching
func generateCacheKey(background, foreground string, interpolation float64) string {
	return fmt.Sprintf("%s_%s_%.6f", background, foreground, interpolation)
}

// get retrieves a cached result or returns false if not found
func (c *interpolationCache) get(key string) (string, bool) {
	c.mu.RLock()
	val, ok := c.cache[key]
	c.mu.RUnlock()
	return val, ok
}

// set stores a computed result in the cache
func (c *interpolationCache) set(key, value string) {
	c.mu.Lock()
	c.cache[key] = value
	c.mu.Unlock()
}

// GlobalInterpolationCache returns a pointer to the global interpolation cache for testing
func GlobalInterpolationCache() *interpolationCache {
	return globalInterpolationCache
}

// getRGB retrieves cached RGB conversion or computes and stores it
func (c *colourCache) getRGB(hex string) (rbgColour, error) {
	c.mu.RLock()
	if rgb, ok := c.rgb[hex]; ok {
		c.mu.RUnlock()
		return rgb, nil
	}
	c.mu.RUnlock()

	// Compute and cache
	c.mu.Lock()
	defer c.mu.Unlock()
	// Double-check after acquiring write lock
	if rgb, ok := c.rgb[hex]; ok {
		return rgb, nil
	}

	rgb, err := hexToRGB(hex)
	if err != nil {
		return rbgColour{}, err
	}
	c.rgb[hex] = rgb
	return rgb, nil
}

// getHSL retrieves cached HSL conversion or computes and stores it
func (c *colourCache) getHSL(hex string) (hslColour, error) {
	c.mu.RLock()
	if hsl, ok := c.hsl[hex]; ok {
		c.mu.RUnlock()
		return hsl, nil
	}
	c.mu.RUnlock()

	// Get RGB first (this may acquire its own lock, but we don't hold any lock yet)
	rgb, err := c.getRGB(hex)
	if err != nil {
		return hslColour{}, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if hsl, ok := c.hsl[hex]; ok {
		return hsl, nil
	}

	// Convert RGB to HSL
	h, s, l := rgbToHSL(rgb)

	// Convert to hslColour type (H: 0-360, S: 0-100, L: 0-100)
	result := hslColour{
		H: h * 360.0,
		S: s * 100.0,
		L: l * 100.0,
	}
	c.hsl[hex] = result
	return result, nil
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
	interpolation float64,
) (string, error) {

	// Parse the input string into segments
	parsed, _ := ansiParse.Parse(content)

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
				err = updateSegmentBackgroundColours(segment, bgCol)
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

			err = updateSegmentForegroundColours(segment, fgCol)
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

			err = updateSegmentForegroundColours(segment, fgCol)
			if err != nil {
				return "", err
			}
		}

	}
	return ansiParse.String(parsed), nil
}

// updateSegmentForegroundColours updates the foreground colours of a segment.
func updateSegmentForegroundColours(segment *ansiParse.StyledText, fgCol string) error {
	if segment.FgCol == nil {
		segment.FgCol = &ansiParse.Col{}
	}

	segment.FgCol.Hex = fgCol
	fgRgb, err := globalColourCache.getRGB(fgCol)
	if err != nil {
		return err
	}
	segment.FgCol.Rgb = fgRgb

	fgHsl, err := globalColourCache.getHSL(fgCol)
	if err != nil {
		return err
	}
	segment.FgCol.Hsl = fgHsl

	return nil
}

// updateSegment updates the background colours of a segment. It will do nothing if the segment
// has no background colour.
func updateSegmentBackgroundColours(segment *ansiParse.StyledText, bgCol string) error {
	if segment.BgCol == nil {
		return nil
	}

	segment.BgCol.Hex = bgCol
	bgRgb, err := globalColourCache.getRGB(bgCol)
	if err != nil {
		return err
	}
	segment.BgCol.Rgb = bgRgb

	bgHsl, err := globalColourCache.getHSL(bgCol)
	if err != nil {
		return err
	}
	segment.BgCol.Hsl = bgHsl

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
func Interpolate(hexBackground, hexForeground string, interpolation float64) (string, error) {
	// Check cache first
	key := generateCacheKey(hexBackground, hexForeground, interpolation)
	if result, ok := globalInterpolationCache.get(key); ok {
		return result, nil
	}

	// Original interpolation logic
	background, err := globalColourCache.getRGB(hexBackground)
	if err != nil {
		return "", err
	}
	foreground, err := globalColourCache.getRGB(hexForeground)
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

	result := rgbToHex(rbgColour{R: r, G: g, B: b})

	// Store result in cache
	globalInterpolationCache.set(key, result)

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

// hexToHSL converts a hex colour string to HSL.
func hexToHSL(hex string) (hslColour, error) {
	rgb, err := globalColourCache.getRGB(hex)
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
