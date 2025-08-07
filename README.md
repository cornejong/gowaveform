# 🎵 GoWaveform

A fast, concurrent Go library and CLI tool for generating beautiful SVG waveforms from audio files. Supports MP3, WAV, FLAC, OGG, AIFF, and Opus formats with multiple calculation modes, concurrent processing, and both library and CLI interfaces.

## ✨ Features

- **🎵 Multi-Format Support**: Works with MP3, WAV, FLAC, OGG, AIFF, and Opus audio files
- **🚀 High Performance**: Optimized with concurrent processing and SIMD-friendly algorithms
- **📚 Library + CLI**: Use as a Go library in your projects or as a standalone CLI tool
- **🎨 Multiple Calculation Modes**: 6 different visualization modes for various aesthetic preferences
- **📱 Interactive Showcase**: Beautiful HTML player with real-time waveform progress
- **⚡ Fast Processing**: Efficient MP3 decoding and SVG generation
- **🎛️ Customizable**: Adjustable width, height, colors, bar count, and styling
- **🔧 Flexible API**: Easy-to-use library interface with sensible defaults

## 🛠️ Installation

### As a Library

```bash
go get -u github.com/cornejong/gowaveform
```

### CLI Tool

```bash
# Clone the repository
git clone https://github.com/cornejong/gowaveform.git
cd gowaveform

# Install dependencies
go mod tidy

# Build the binary
go build -o gowaveform main.go
```

## 🚀 Usage

### Library Usage

#### Basic Example

```go
package main

import (
    "log"
    "github.com/cornejong/gowaveform/waveform"
)

func main() {
    // Create waveform with default settings - works with MP3, WAV, FLAC, OGG, AIFF, Opus
    w, err := waveform.NewFromAudioFile("audio.mp3", nil) // or .wav, .flac, .ogg, .aiff, .opus
    if err != nil {
        log.Fatal(err)
    }
    
    // Save as SVG
    err = w.WriteSVG("waveform.svg")
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Custom Configuration

```go
package main

import (
    "log"
    "github.com/cornejong/gowaveform/waveform"
)

func main() {
    // Create custom configuration
    config := &waveform.Config{
        Width:        800,
        Height:       120,
        Bars:         200,
        BarSpacing:   1,
        BarColor:     "#FF6B6B",
        CornerRadius: 10.0,
        Concurrent:   true,
        Mode:         waveform.ModeDynamic,
    }
    
    // Generate waveform
    w, err := waveform.NewFromAudioFile("audio.wav", config) // supports .mp3, .wav, .flac, .ogg, .aiff, .opus
    if err != nil {
        log.Fatal(err)
    }
    
    // Write to file
    err = w.WriteSVG("custom_waveform.svg")
    if err != nil {
        log.Fatal(err)
    }
}
```

#### Generate SVG in Memory

```go
package main

import (
    "fmt"
    "log"
    "github.com/cornejong/gowaveform/waveform"
)

func main() {
    config := waveform.DefaultConfig()
    config.Mode = waveform.ModeVU
    config.BarColor = "#4ECDC4"
    
    w, err := waveform.NewFromAudioFile("audio.flac", config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get SVG data as bytes (useful for web servers)
    svgData, err := w.GenerateSVG()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("SVG size: %d bytes\n", len(svgData))
    // Use svgData for HTTP responses, etc.
}
```

### CLI Usage

#### Basic Usage

```bash
# Generate a waveform with default settings - supports multiple formats
./gowaveform input.mp3 output.svg    # MP3
./gowaveform input.wav output.svg    # WAV  
./gowaveform input.flac output.svg   # FLAC
./gowaveform input.ogg output.svg    # OGG
./gowaveform input.aiff output.svg   # AIFF
./gowaveform input.opus output.svg   # Opus

# Custom dimensions and styling
./gowaveform -width 800 -height 120 -bars 200 -color "#FF6B6B" input.wav output.svg
```

#### Advanced Usage

```bash
# Use different calculation modes with various formats
./gowaveform -mode lufs input.flac waveform-lufs.svg
./gowaveform -mode smooth input.wav waveform-smooth.svg
./gowaveform -mode vu input.ogg waveform-vu.svg
./gowaveform -mode rms input.aiff waveform-rms.svg

# Customize appearance
./gowaveform \
  -width 1000 \
  -height 150 \
  -bars 300 \
  -spacing 1 \
  -color "#8B5CF6" \
  -radius 4.0 \
  input.flac output.svg
```

## 🎛️ Calculation Modes

GoWaveform offers 6 distinct calculation modes, each optimized for different visual styles:

| Mode | Description | Best For |
|------|-------------|----------|
| **`rms`** | Traditional root-mean-square calculation | Standard waveform representation |
| **`lufs`** | Perceptual loudness with psychoacoustic weighting | Dramatic differences, broadcast-style |
| **`peak`** | Maximum amplitude detection | Fast processing, showing peak levels |
| **`vu`** | Broadcast-style VU meter simulation | Smooth, professional visualization |
| **`dynamic`** | Emphasizes differences between loud/quiet sections | Highlighting dynamic range |
| **`smooth`** | Heavily filtered for clean aesthetics | Minimal, modern design |

### Mode Examples

```bash
# Traditional balanced waveform
./gowaveform -mode rms audio.wav standard.svg

# Dramatic, perceptually-weighted visualization  
./gowaveform -mode lufs audio.flac dramatic.svg

# Clean, minimal aesthetic
./gowaveform -mode smooth audio.ogg minimal.svg

# Professional broadcast style
./gowaveform -mode vu audio.aiff broadcast.svg
```

## ⚙️ Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-width` | `500` | Total SVG width in pixels |
| `-height` | `80` | Total SVG height in pixels |
| `-bars` | `100` | Number of bars in waveform |
| `-spacing` | `2` | Space between bars in pixels |
| `-color` | `#3B82F6` | Bar color (hex format) |
| `-radius` | `8.0` | Bar corner radius for rounded edges |
| `-mode` | `dynamic` | Calculation mode (see modes above) |
| `-concurrent` | `true` | Enable concurrent processing |

## 🎮 Interactive Showcase

The repository includes a beautiful HTML showcase that demonstrates all the waveform modes with an interactive audio player.

### Features:
- **Real-time Progress**: Waveform changes color as audio plays
- **Visual Playhead**: Moving indicator shows exact playback position
- **Opacity Transitions**: Played portion at 100% opacity, unplayed at 50%
- **Multiple Tracks**: Compare different calculation modes side-by-side
- **Responsive Design**: Works on desktop and mobile devices

### Setup Showcase:

```bash
# Generate example waveforms
./gowaveform -mode dynamic example.mp3 test.svg
./gowaveform -mode lufs example2.mp3 test2.svg
./gowaveform -mode smooth example3.mp3 test3.svg

# Open showcase.html in your browser
open showcase.html
```

## 🔧 Technical Details

### Performance Optimizations

- **Concurrent Processing**: Automatic multi-core utilization for large files
- **SIMD-Friendly Algorithms**: Optimized mathematical operations
- **Memory Efficiency**: Pre-allocated buffers and minimal allocations
- **Fast Square Root**: Quake III-style bit manipulation for speed

### Audio Processing

- **Multi-Format Decoding**: Native support for MP3, WAV, FLAC, OGG, AIFF, and Opus formats
- **Sample Processing**: 16-bit PCM processing with configurable bucket sizes
- **Dynamic Range**: Intelligent normalization preserving audio characteristics

### SVG Generation

- **Vector Graphics**: Clean, scalable SVG output
- **Modern Styling**: Rounded corners and flat design
- **Optimized Output**: Minimal file size with clean markup

## 📊 Performance Benchmarks

Typical performance on modern hardware:

| File Size | Duration | Bars | Processing Time |
|-----------|----------|------|-----------------|
| 3MB | 3 minutes | 100 | ~50ms |
| 8MB | 8 minutes | 200 | ~120ms |
| 15MB | 15 minutes | 500 | ~300ms |

*Benchmarks on MacBook Pro M1, concurrent processing enabled*

## Programmatic Usage

```go
// Use as a library with any supported format
import "github.com/cornejong/gowaveform/waveform"

config := waveform.DefaultConfig()
w, err := waveform.NewFromAudioFile("audio.flac", config) // .mp3, .wav, .flac, .ogg, .aiff
err = w.WriteSVG("output.svg")
```

## 🔄 Batch Processing

Process multiple files efficiently:

```bash
#!/bin/bash
# Batch process all audio files
for file in *.{mp3,wav,flac,ogg,aiff,opus}; do
    [ -f "$file" ] && ./gowaveform -mode dynamic "$file" "${file%.*}.svg"
done
```

## 🚧 Roadmap

- [x] Additional audio format support (✅ **COMPLETED:** FLAC, WAV, OGG, AIFF, Opus)
- [ ] Real-time streaming waveform generation
- [ ] Advanced colorization options
- [ ] PNG/WebP output formats
- [ ] REST API server mode

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Development Setup

```bash
# Clone and setup
git clone https://github.com/cornejong/gowaveform.git
cd gowaveform
go mod tidy

# Run tests
go test ./...

# Build and test
go build -o gowaveform main.go
./gowaveform example.mp3 test.svg

# Test with different formats (if available)
# ./gowaveform example.wav test-wav.svg
# ./gowaveform example.flac test-flac.svg  
# ./gowaveform example.ogg test-ogg.svg
# ./gowaveform example.aiff test-aiff.svg
# ./gowaveform example.opus test-opus.svg
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [hajimehoshi/go-mp3](https://github.com/hajimehoshi/go-mp3) - Excellent MP3 decoding library
- [go-audio/wav](https://github.com/go-audio/wav) - Reliable WAV file support
- [go-audio/aiff](https://github.com/go-audio/aiff) - AIFF format support
- [mewkiz/flac](https://github.com/mewkiz/flac) - High-quality FLAC decoder
- [jfreymuth/oggvorbis](https://github.com/jfreymuth/oggvorbis) - Clean OGG Vorbis implementation
- [pion/opus](https://github.com/pion/opus) - Opus audio codec implementation
- [tdewolff/canvas](https://github.com/tdewolff/canvas) - Powerful SVG generation toolkit
- Quake III - Fast square root algorithm inspiration

---

<div align="center">
Made with ❤️ and Go

**[⭐ Star this repo](https://github.com/cornejong/gowaveform) if you find it useful!**
</div>
