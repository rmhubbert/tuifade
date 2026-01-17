# tuifade

TUI luminance management - A Go package for fading the background and foreground colours of ANSI strings.

## Overview

`tuifade` provides functions to apply luminance-based fading effects to ANSI-coloured text strings. It's particularly useful for creating subtle visual effects in terminal user interfaces, such as fading text to blend with terminal backgrounds or creating smooth colour transitions.

## Features

- **ANSI String Processing**: Preserves existing ANSI codes while applying colour transformations
- **True Color Support**: Requires truecolour-capable terminals (24-bit colour)
- **Linear Color Interpolation**: Uses proper linear RGB colour space for accurate fading
- **Terminal Integration**: Automatically detects terminal background/foreground colours
- **Flexible Fading Control**: Adjustable interpolation parameter for fine-grained control

## Installation

```bash
go get github.com/rmhubbert/tuifade
```

## Requirements

- Go 1.25.5 or later
- A terminal that supports truecolour (24-bit colour)
- Compatible terminal environments (most modern terminals support truecolour)

## Usage

### Basic Usage

The main function is `Fade()`, which applies luminance fading to ANSI strings:

```go
package main

import (
    "fmt"
    "github.com/rmhubbert/tuifade"
)

func main() {
    // ANSI string with red text
    colouredText := "\x1b[31mHello, World!\x1b[0m"
    
    // Apply 50% fade
    faded, err := tuifade.Fade(colouredText, 0.5)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Println(faded)
}
```

### Interpolation Values

The `interpolation` parameter controls the degree of fading:

- `1.0`: No fade (original colours preserved)
- `0.5`: 50% fade (colours blended halfway with terminal background)
- `0.0`: Full fade (colours become terminal background/foreground)

### Advanced Usage with Custom Color Interpolation

For more control, you can use the `Interpolate()` function directly:

```go
package main

import (
    "fmt"
    "github.com/rmhubbert/tuifade"
)

func main() {
    // Interpolate between two colours
    result, err := tuifade.Interpolate("#ff0000", "#0000ff", 0.5)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Interpolated colour: %s\n", result) // #800080 (purple)
}
```

### Working with Complex ANSI Strings

The package handles complex ANSI sequences including:

```go
package main

import (
    "fmt"
    "github.com/rmhubbert/tuifade"
)

func main() {
    // Complex ANSI with multiple colours and styles
    complexText := "\x1b[1;31;44mBold red on blue\x1b[32mGreen text\x1b[0m"
    
    // Apply fading
    faded, err := tuifade.Fade(complexText, 0.3)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Println(faded)
}
```

## API Reference

### `func Fade(content string, interpolation float64) (string, error)`

Fades the background and foreground colours of an ANSI string using the terminal's default colours.

**Parameters:**
- `content`: ANSI string to process
- `interpolation`: Fade amount (0.0 = full fade, 1.0 = no fade)

**Returns:**
- `string`: Faded ANSI string
- `error`: Error if terminal doesn't support truecolour

### `func Interpolate(hexBackground, hexForeground string, interpolation float64) (string, error)`

Interpolates between two hex colours.

**Parameters:**
- `hexBackground`: Background colour in hex format (#RRGGBB)
- `hexForeground`: Foreground colour in hex format (#RRGGBB)
- `interpolation`: Interpolation amount (0.0 = background, 1.0 = foreground)

**Returns:**
- `string`: Interpolated colour in hex format
- `error`: Error if colour formats are invalid

## Error Handling

The package returns errors in these situations:

1. **Non-truecolour terminals**: `Fade()` returns an error if the terminal doesn't support truecolour
2. **Invalid colour formats**: `Interpolate()` returns errors for malformed hex colour strings
3. **Interpolation clamping**: Values outside [0, 1] range are automatically clamped

```go
faded, err := tuifade.Fade(colouredText, 0.5)
if err != nil {
    // Handle error - most commonly "fade only supports truecolour terminals"
    fmt.Printf("Terminal limitation: %v\n", err)
    // Fallback: use original text
    faded = colouredText
}
```

## Colour Space

The package uses linear RGB colour space for interpolation, which provides more accurate colour transitions than sRGB. This ensures that luminance changes appear natural and consistent across different colour combinations.

## Testing

Run the test suite:

```bash
go test
```

Run benchmarks:

```bash
go test -bench=.
```

## Dependencies

- `github.com/leaanthony/go-ansi-parser` - ANSI string parsing
- `github.com/lucasb-eyer/go-colourful` - Color space conversions
- `github.com/muesli/termenv` - Terminal environment detection

## Examples

### Creating a Progress Effect

```go
func showProgress() {
    text := "\x1b[32mProcessing...\x1b[0m"
    
    for i := 10; i >= 0; i-- {
        fadeAmount := float64(i) / 10.0
        faded, _ := tuifade.Fade(text, fadeAmount)
        fmt.Printf("\r%s", faded)
        time.Sleep(200 * time.Millisecond)
    }
    fmt.Println()
}
```

### Subtle UI Elements

```go
func createSubtleHint() {
    hint := "\x1b[36mPress 'q' to quit\x1b[0m"
    subtle, _ := tuifade.Fade(hint, 0.7) // 30% faded
    fmt.Println(subtle)
}
```

## License

tuifade is licensed under the MIT License. See the LICENSE file for more information.
