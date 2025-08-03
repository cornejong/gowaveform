# 🎵 GoWaveform

A Go-based tool for generating SVG waveform visualizations from MP3 audio files. Features multiple calculation modes, concurrent processing, and a beautiful interactive HTML showcase.

![GoWaveform Demo](https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

## ✨ Features

- **🚀 High Performance**: Optimized with concurrent processing and SIMD-friendly algorithms
- **🎨 Multiple Calculation Modes**: 6 different visualization modes for various aesthetic preferences
- **📱 Interactive Showcase**: Beautiful HTML player with real-time waveform progress
- **⚡ Fast Processing**: Efficient MP3 decoding and SVG generation
- **🎛️ Customizable**: Adjustable width, height, colors, bar count, and styling

## 🛠️ Installation

```bash
# Clone the repository
git clone https://github.com/cornejong/gowaveform.git
cd gowaveform

# Install dependencies
go mod tidy

# Build the binary
go build -o gowaveform main.go
```

## 🚀 Quick Start

### Basic Usage

```bash
# Generate a waveform with default settings
./gowaveform input.mp3 output.svg

# Custom dimensions and styling
./gowaveform -width 800 -height 120 -bars 200 -color "#FF6B6B" input.mp3 output.svg
```

### Advanced Usage

```bash
# Use different calculation modes
./gowaveform -mode lufs input.mp3 waveform-lufs.svg
./gowaveform -mode smooth input.mp3 waveform-smooth.svg

# Customize appearance
./gowaveform \
  -width 1000 \
  -height 150 \
  -bars 300 \
  -spacing 1 \
  -color "#8B5CF6" \
  -radius 4.0 \
  input.mp3 output.svg
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
./gowaveform -mode rms audio.mp3 standard.svg

# Dramatic, perceptually-weighted visualization
./gowaveform -mode lufs audio.mp3 dramatic.svg

# Clean, minimal aesthetic
./gowaveform -mode smooth audio.mp3 minimal.svg

# Professional broadcast style
./gowaveform -mode vu audio.mp3 broadcast.svg
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

- **MP3 Decoding**: Uses `hajimehoshi/go-mp3` for reliable decoding
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

## 🎨 Integration Examples

### Web Integration

```html
<!-- Embed generated SVG -->
<div class="waveform-container">
    <svg><!-- Your generated waveform --></svg>
    <div class="playhead"></div>
</div>
```

### Programmatic Usage

```go
// Use as a library
import "github.com/cornejong/gowaveform"

samples, err := readSamples("audio.mp3")
peaks := downsampleConcurrent(samples, 200)
writeSVG(peaks, "output.svg")
```

## 🔄 Batch Processing

Process multiple files efficiently:

```bash
#!/bin/bash
# Batch process all MP3 files
for file in *.mp3; do
    ./gowaveform -mode dynamic "$file" "${file%.mp3}.svg"
done
```

## 🚧 Roadmap

- [ ] Additional audio format support (FLAC, WAV, OGG)
- [ ] Real-time streaming waveform generation
- [ ] Advanced colorization options
- [ ] PNG/WebP output formats
- [ ] REST API server mode
- [ ] Docker containerization

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
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [hajimehoshi/go-mp3](https://github.com/hajimehoshi/go-mp3) - Excellent MP3 decoding library
- [tdewolff/canvas](https://github.com/tdewolff/canvas) - Powerful SVG generation toolkit
- Quake III - Fast square root algorithm inspiration

## 📞 Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/cornejong/gowaveform/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/cornejong/gowaveform/discussions)
- 📧 **Contact**: Create an issue for any questions

---

<div align="center">
Made with ❤️ and Go

**[⭐ Star this repo](https://github.com/cornejong/gowaveform) if you find it useful!**
</div>
