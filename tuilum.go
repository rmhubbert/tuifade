package tuilum

import (
	"fmt"
	"math"
)

type rbgColour struct {
	Red   uint8
	Green uint8
	Blue  uint8
}

func Fade(hexBackground, hexForeground string, interpolation float64) (string, error) {
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
	r := interpolateChannel(background.Red, foreground.Red, bgWeight, fgWeight)
	g := interpolateChannel(background.Green, foreground.Green, bgWeight, fgWeight)
	b := interpolateChannel(background.Blue, foreground.Blue, bgWeight, fgWeight)

	return rgbToHex(rbgColour{r, g, b}), nil
}

// interpolateChannel performs linear interpolation for a single color channel.
func interpolateChannel(bg, fg uint8, bgWeight, fgWeight float64) uint8 {
	bgValue := float64(bg)
	fgValue := float64(fg)
	result := bgValue*bgWeight + fgValue*fgWeight
	return uint8(math.Round(result))
}

func rgbToHex(rgb rbgColour) string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.Red, rgb.Green, rgb.Blue)
}

func hexToRGB(hex string) (rbgColour, error) {
	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return rbgColour{}, err
	}
	return rbgColour{r, g, b}, nil
}
