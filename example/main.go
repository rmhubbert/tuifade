package main

import (
	"fmt"

	"github.com/goforj/godump"
	"github.com/rmhubbert/tuifade"
)

func main() {
	// ANSI string with red text
	colouredText := "\x1b[31mHello, World!\x1b[0m, this is a test"
	godump.Dump(colouredText)

	// Apply 50% fade
	faded, err := tuifade.Fade(colouredText, 0.5)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	godump.Dump(faded)
	fmt.Println(faded)
}
