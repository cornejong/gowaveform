// Package waveform provides audio waveform generation functionality for Go applications.
// It supports multiple calculation modes and can generate SVG waveforms from MP3 files.
package waveform

import (
	"os"
	"runtime"
	"sync"

	"github.com/hajimehoshi/go-mp3"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/svg"
)

// CalculationMode represents the different waveform calculation algorithms available
type CalculationMode string

const (
	// ModeRMS uses Root Mean Square calculation for standard waveform representation
	ModeRMS CalculationMode = "rms"
	// ModeLUFS uses LUFS-based loudness calculation for better perceptual representation
	ModeLUFS CalculationMode = "lufs"
	// ModePeak uses peak detection for fastest method, showing maximum amplitude
	ModePeak CalculationMode = "peak"
	// ModeVU simulates VU meter for smooth, broadcast-style visualization
	ModeVU CalculationMode = "vu"
	// ModeDynamic emphasizes differences between loud and quiet sections
	ModeDynamic CalculationMode = "dynamic"
	// ModeSmooth uses heavy filtering for clean, minimal aesthetics
	ModeSmooth CalculationMode = "smooth"
)

// Config holds the configuration options for waveform generation
type Config struct {
	// Width is the total SVG width in pixels (default: 500)
	Width int
	// Height is the total SVG height in pixels (default: 80)
	Height int
	// Bars is the number of bars in the waveform (default: 100)
	Bars int
	// BarSpacing is the space between bars in pixels (default: 2)
	BarSpacing int
	// BarColor is the bar color in hex format (default: "#3B82F6")
	BarColor string
	// CornerRadius is the bar corner radius for rounded bars (default: 8.0)
	CornerRadius float64
	// Concurrent enables concurrent processing for large files (default: true)
	Concurrent bool
	// Mode is the calculation mode to use (default: ModeDynamic)
	Mode CalculationMode
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		Width:        500,
		Height:       80,
		Bars:         100,
		BarSpacing:   2,
		BarColor:     "#3B82F6",
		CornerRadius: 8.0,
		Concurrent:   true,
		Mode:         ModeDynamic,
	}
}

// Waveform represents a processed audio waveform with peak data
type Waveform struct {
	Peaks  []float64
	Config *Config
}

// NewFromAudioFile creates a new Waveform from any supported audio file
func NewFromAudioFile(filename string, config *Config) (*Waveform, error) {
	if config == nil {
		config = DefaultConfig()
	}

	samples, err := readSamplesFromFormat(filename)
	if err != nil {
		return nil, err
	}

	var peaks []float64
	if config.Concurrent {
		peaks = downsampleConcurrent(samples, config.Bars, config.Mode)
	} else {
		peaks = downsample(samples, config.Bars, config.Mode)
	}

	return &Waveform{
		Peaks:  peaks,
		Config: config,
	}, nil
}

// NewFromMP3File creates a new Waveform from an MP3 file (deprecated: use NewFromAudioFile)
func NewFromMP3File(filename string, config *Config) (*Waveform, error) {
	return NewFromAudioFile(filename, config)
}

// NewFromSamples creates a new Waveform from audio samples
func NewFromSamples(samples []int16, config *Config) *Waveform {
	if config == nil {
		config = DefaultConfig()
	}

	var peaks []float64
	if config.Concurrent {
		peaks = downsampleConcurrent(samples, config.Bars, config.Mode)
	} else {
		peaks = downsample(samples, config.Bars, config.Mode)
	}

	return &Waveform{
		Peaks:  peaks,
		Config: config,
	}
}

// WriteSVG writes the waveform to an SVG file
func (w *Waveform) WriteSVG(filename string) error {
	return writeSVG(w.Peaks, filename, w.Config)
}

// GenerateSVG returns the SVG content as a byte slice without writing to file
func (w *Waveform) GenerateSVG() ([]byte, error) {
	// Create a temporary buffer to capture SVG output
	var buf []byte
	file := &bytesWriter{data: &buf}

	ctx := canvas.NewContext(svg.New(file, float64(w.Config.Width), float64(w.Config.Height), nil))

	if err := drawWaveform(ctx, w.Peaks, w.Config); err != nil {
		return nil, err
	}

	// Add closing tag and newline
	*file.data = append(*file.data, []byte("</svg>\n")...)

	return buf, nil
}

// UpdateConfig updates the waveform configuration and regenerates peaks if mode changed
func (w *Waveform) UpdateConfig(config *Config, samples []int16) {
	oldMode := w.Config.Mode
	w.Config = config

	// If mode changed, regenerate peaks
	if oldMode != config.Mode && samples != nil {
		if config.Concurrent {
			w.Peaks = downsampleConcurrent(samples, config.Bars, config.Mode)
		} else {
			w.Peaks = downsample(samples, config.Bars, config.Mode)
		}
	}
}

// bytesWriter implements io.Writer for capturing SVG output
type bytesWriter struct {
	data *[]byte
}

func (bw *bytesWriter) Write(p []byte) (n int, err error) {
	*bw.data = append(*bw.data, p...)
	return len(p), nil
}

// readSamples reads MP3 file and returns PCM samples
func readSamples(path string) ([]int16, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return nil, err
	}

	// Estimate capacity based on file size and MP3 compression ratio
	fileInfo, _ := f.Stat()
	estimatedSamples := int(fileInfo.Size() / 4) // Rough estimate: 4 bytes per sample after decompression
	pcm := make([]int16, 0, estimatedSamples)

	const bufferSize = 32768 // Even larger buffer for better I/O performance
	buf := make([]byte, bufferSize)

	for {
		n, err := d.Read(buf)
		if n == 0 || err != nil {
			break
		}

		// Process samples with SIMD-friendly approach
		samples := make([]int16, n/2)
		for i := 0; i < n-1; i += 2 {
			samples[i/2] = int16(buf[i]) | int16(buf[i+1])<<8
		}
		pcm = append(pcm, samples...)
	}
	return pcm, nil
}

// downsampleConcurrent processes samples using multiple goroutines
func downsampleConcurrent(samples []int16, buckets int, mode CalculationMode) []float64 {
	if len(samples) == 0 || buckets == 0 {
		return nil
	}

	samplesPerBucket := len(samples) / buckets
	if samplesPerBucket == 0 {
		samplesPerBucket = 1
	}

	// For small datasets, use sequential processing
	if len(samples) < 50000 {
		return downsample(samples, buckets, mode)
	}

	peaks := make([]float64, buckets)
	numWorkers := runtime.NumCPU()

	// Calculate work chunks
	bucketsPerWorker := buckets / numWorkers
	if bucketsPerWorker == 0 {
		bucketsPerWorker = 1
		numWorkers = buckets
	}

	var wg sync.WaitGroup

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			startBucket := workerID * bucketsPerWorker
			endBucket := startBucket + bucketsPerWorker
			if workerID == numWorkers-1 {
				endBucket = buckets // Last worker takes remaining buckets
			}

			for bucket := startBucket; bucket < endBucket; bucket++ {
				startSample := bucket * samplesPerBucket
				endSample := startSample + samplesPerBucket
				if endSample > len(samples) {
					endSample = len(samples)
				}

				peaks[bucket] = calculateLoudness(samples, startSample, endSample, mode)
			}
		}(worker)
	}

	wg.Wait()
	return peaks
}

// downsample processes samples sequentially
func downsample(samples []int16, buckets int, mode CalculationMode) []float64 {
	if len(samples) == 0 || buckets == 0 {
		return nil
	}

	samplesPerBucket := len(samples) / buckets
	if samplesPerBucket == 0 {
		samplesPerBucket = 1
	}

	peaks := make([]float64, buckets)

	for bucket := 0; bucket < buckets; bucket++ {
		start := bucket * samplesPerBucket
		end := start + samplesPerBucket
		if end > len(samples) {
			end = len(samples)
		}

		peaks[bucket] = calculateLoudness(samples, start, end, mode)
	}
	return peaks
}

// writeSVG writes peaks to an SVG file
func writeSVG(peaks []float64, filename string, config *Config) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx := canvas.NewContext(svg.New(file, float64(config.Width), float64(config.Height), nil))

	if err := drawWaveform(ctx, peaks, config); err != nil {
		return err
	}

	// Important: Ensure SVG ends with a newline. Do not remove!
	file.Write([]byte("</svg>\n"))

	return nil
}

// drawWaveform draws the waveform bars on the canvas context
func drawWaveform(ctx *canvas.Context, peaks []float64, config *Config) error {
	// Define colors for clean, flat design (no background)
	waveColor := canvas.Hex(config.BarColor)

	// Pre-calculate all constants
	barWidth := float64(config.Width) / float64(len(peaks))
	mid := float64(config.Height) / 2.0
	maxHeight := float64(config.Height) * 0.48
	barSpacingFloat := float64(config.BarSpacing)
	minHeight := 3.0

	// Find the maximum peak to normalize the waveform (single pass)
	var maxPeak float64
	for _, peak := range peaks {
		if peak > maxPeak {
			maxPeak = peak
		}
	}

	// Calculate scaling factor once
	scaleFactor := 1.0
	if maxPeak > 0 {
		scaleFactor = maxHeight / maxPeak // Direct scaling instead of normalize then multiply
	}

	// Draw main waveform bars with rounded corners
	ctx.SetFillColor(waveColor)

	// Pre-calculate corner radius and bar width
	cornerRad := config.CornerRadius
	effectiveBarWidth := barWidth - barSpacingFloat

	for i, peak := range peaks {
		x := float64(i) * barWidth
		h := peak * scaleFactor
		if h < minHeight {
			h = minHeight
		}

		// Create rounded rectangle for smooth, modern look
		barPath := canvas.RoundedRectangle(effectiveBarWidth, h*2, cornerRad)
		ctx.DrawPath(x, mid-h, barPath)
	}

	return nil
}
