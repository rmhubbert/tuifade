package main

import (
	"fmt"
	"strings"

	"github.com/goforj/godump"
	"github.com/rmhubbert/tuifade"
)

func main() {
	d := godump.NewDumper(godump.WithoutHeader())
	fade := 1.0
	colouredText := "\x1b[31mHello, World!\x1b[0m, this is a test. \x1b[33;45mThe end\x1b[0m."
	fmt.Println(colouredText)
	d.Dump(colouredText)

	faded, err := tuifade.Fade(colouredText, fade)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(faded)
	d.Dump(faded)

	repeated := strings.Repeat(colouredText, 3)
	fadedRepeated, err := tuifade.Fade(repeated, fade)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(fadedRepeated)
	d.Dump(fadedRepeated)

}
