# tuilum

TUI luminance management - A Go package for fading the background and foreground colours of ANSI strings.

## Overview

`tuilum` provides functions to apply luminance-based fading effects to ANSI-colored text strings. It's particularly useful for creating subtle visual effects in terminal user interfaces, such as fading text to blend with terminal backgrounds or creating smooth color transitions.

## Features

- **ANSI String Processing**: Preserves existing ANSI codes while applying color transformations
- **True Color Support**: Requires truecolor-capable terminals (24-bit color)
- **Linear Color Interpolation**: Uses proper linear RGB color space for accurate fading
- **Terminal Integration**: Automatically detects terminal background/foreground colors
- **Flexible Fading Control**: Adjustable interpolation parameter for fine-grained control

## Installation

```bash
go get github.com/rmhubbert/tuilum
```

## Requirements

- Go 1.25.5 or later
- A terminal that supports truecolor (24-bit color)
- Compatible terminal environments (most modern terminals support truecolor)

## Usage

### Basic Usage

The main function is `Fade()`, which applies luminance fading to ANSI strings:

```go
package main

import (
    "fmt"
    "github.com/rmhubbert/tuilum"
)

func main() {
    // ANSI string with red text
    coloredText := "\x1b[31mHello, World!\x1b[0m"
    
    // Apply 50% fade
    faded, err := tuilum.Fade(coloredText, 0.5)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Println(faded)
}
```

### Interpolation Values

The `interpolation` parameter controls the degree of fading:

- `1.0`: No fade (original colors preserved)
- `0.5`: 50% fade (colors blended halfway with terminal background)
- `0.0`: Full fade (colors become terminal background/foreground)

### Advanced Usage with Custom Color Interpolation

For more control, you can use the `Interpolate()` function directly:

```go
package main

import (
    "fmt"
    "github.com/rmhubbert/tuilum"
)

func main() {
    // Interpolate between two colors
    result, err := tuilum.Interpolate("#ff0000", "#0000ff", 0.5)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Interpolated color: %s\n", result) // #800080 (purple)
}
```

### Working with Complex ANSI Strings

The package handles complex ANSI sequences including:

```go
package main

import (
    "fmt"
    "github.com/rmhubbert/tuilum"
)

func main() {
    // Complex ANSI with multiple colors and styles
    complexText := "\x1b[1;31;44mBold red on blue\x1b[32mGreen text\x1b[0m"
    
    // Apply fading
    faded, err := tuilum.Fade(complexText, 0.3)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Println(faded)
}
```

## API Reference

### `func Fade(content string, interpolation float64) (string, error)`

Fades the background and foreground colors of an ANSI string using the terminal's default colors.

**Parameters:**
- `content`: ANSI string to process
- `interpolation`: Fade amount (0.0 = full fade, 1.0 = no fade)

**Returns:**
- `string`: Faded ANSI string
- `error`: Error if terminal doesn't support truecolor

### `func Interpolate(hexBackground, hexForeground string, interpolation float64) (string, error)`

Interpolates between two hex colors.

**Parameters:**
- `hexBackground`: Background color in hex format (#RRGGBB)
- `hexForeground`: Foreground color in hex format (#RRGGBB)
- `interpolation`: Interpolation amount (0.0 = background, 1.0 = foreground)

**Returns:**
- `string`: Interpolated color in hex format
- `error`: Error if color formats are invalid

## Error Handling

The package returns errors in these situations:

1. **Non-truecolor terminals**: `Fade()` returns an error if the terminal doesn't support truecolor
2. **Invalid color formats**: `Interpolate()` returns errors for malformed hex color strings
3. **Interpolation clamping**: Values outside [0, 1] range are automatically clamped

```go
faded, err := tuilum.Fade(coloredText, 0.5)
if err != nil {
    // Handle error - most commonly "fade only supports truecolor terminals"
    fmt.Printf("Terminal limitation: %v\n", err)
    // Fallback: use original text
    faded = coloredText
}
```

## Color Space

The package uses linear RGB color space for interpolation, which provides more accurate color transitions than sRGB. This ensures that luminance changes appear natural and consistent across different color combinations.

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
- `github.com/lucasb-eyer/go-colorful` - Color space conversions
- `github.com/muesli/termenv` - Terminal environment detection

## Examples

### Creating a Progress Effect

```go
func showProgress() {
    text := "\x1b[32mProcessing...\x1b[0m"
    
    for i := 10; i >= 0; i-- {
        fadeAmount := float64(i) / 10.0
        faded, _ := tuilum.Fade(text, fadeAmount)
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
    subtle, _ := tuilum.Fade(hint, 0.7) // 30% faded
    fmt.Println(subtle)
}
```

## License

[Add your license information here]
