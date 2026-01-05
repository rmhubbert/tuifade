package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/rmhubbert/tuilum"
)

func main() {
	output := termenv.DefaultOutput()
	foregroundColour := output.ForegroundColor()
	backgroundColour := output.BackgroundColor()
	darkTheme := output.HasDarkBackground()
	profile := output.EnvColorProfile()

	fmt.Printf("Foreground: %s\n", foregroundColour)
	fmt.Printf("Background: %s\n", backgroundColour)
	fmt.Printf("Dark Theme: %t\n", darkTheme)
	fmt.Printf("Profile: %s\n", profile.Name())

	fgString := fmt.Sprintf("%s", foregroundColour)
	bgString := fmt.Sprintf("%s", backgroundColour)
	fadedFg, _ := tuilum.Fade(fgString, bgString, 0.7)
	var style = lipgloss.NewStyle().Foreground(lipgloss.Color(fadedFg))

	content := "Hello, \x1b[2;37;41mWorld!"
	// parsed, _ := ansiParse.Parse(content)
	//
	// for _, segment := range parsed {
	// 	if segment.Label == "" {
	// 		continue
	// 	}
	// 	bgHex := segment.BgCol.Hex
	// 	fgHex := segment.FgCol.Hex
	//
	// }

	fmt.Println(style.Render(content))
}
