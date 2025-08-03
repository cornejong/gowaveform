package main

import (
	"flag"
	"log"
	"os"
	"runtime"
	"sync"
	"unsafe"

	"github.com/hajimehoshi/go-mp3"
	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/svg"
)

// LUFS-based loudness calculation for better perceptual representation
// This applies psychoacoustic weighting and provides more dramatic differences
func calculateLUFS(samples []int16, start, end int) float64 {
	if end <= start {
		return 0
	}

	bucketSize := end - start
	const invMaxSample = 1.0 / 32768.0

	// Pre-emphasis filter coefficients (simulating K-weighting)
	// This emphasizes frequencies that humans are more sensitive to
	var sum float64
	var prevSample float64

	for i := start; i < end; i++ {
		// Convert to float and normalize
		sample := float64(samples[i]) * invMaxSample

		// Simple high-pass pre-emphasis (approximates K-weighting)
		// Emphasizes mid frequencies where human hearing is most sensitive
		filtered := sample - 0.85*prevSample
		prevSample = sample

		// Apply psychoacoustic scaling - emphasize differences
		// Use a power function to exaggerate dynamic range
		absFiltered := filtered
		if absFiltered < 0 {
			absFiltered = -absFiltered
		}

		// Apply non-linear scaling to emphasize louder parts more
		// This creates more dramatic differences between quiet and loud sections
		scaled := absFiltered * absFiltered * (1.0 + absFiltered*0.5)
		sum += scaled
	}

	if bucketSize > 0 {
		meanSquare := sum / float64(bucketSize)
		// Convert to LUFS-like scale with exaggerated dynamics
		lufs := fastSqrt(meanSquare)

		// Apply additional dynamic enhancement
		// Quiet parts become quieter, loud parts become louder
		if lufs > 0.1 {
			lufs = lufs * lufs * 2.0 // Emphasize loud parts more
		} else {
			lufs = lufs * 0.5 // Make quiet parts even quieter
		}

		return lufs
	}

	return 0
}

// Traditional RMS calculation for standard waveform representation
func calculateRMS(samples []int16, start, end int) float64 {
	if end <= start {
		return 0
	}

	bucketSize := end - start
	const invMaxSample = 1.0 / 32768.0

	var sum float64

	// Unroll loop by 8 for better performance
	i := start
	for ; i <= end-8; i += 8 {
		val1 := float64(samples[i]) * invMaxSample
		val2 := float64(samples[i+1]) * invMaxSample
		val3 := float64(samples[i+2]) * invMaxSample
		val4 := float64(samples[i+3]) * invMaxSample
		val5 := float64(samples[i+4]) * invMaxSample
		val6 := float64(samples[i+5]) * invMaxSample
		val7 := float64(samples[i+6]) * invMaxSample
		val8 := float64(samples[i+7]) * invMaxSample
		sum += val1*val1 + val2*val2 + val3*val3 + val4*val4 + val5*val5 + val6*val6 + val7*val7 + val8*val8
	}

	// Handle remaining samples
	for ; i < end; i++ {
		val := float64(samples[i]) * invMaxSample
		sum += val * val
	}

	if bucketSize > 0 {
		return fastSqrt(sum / float64(bucketSize))
	}

	return 0
}

// Peak detection - fastest method, shows maximum amplitude
func calculatePeak(samples []int16, start, end int) float64 {
	if end <= start {
		return 0
	}

	const invMaxSample = 1.0 / 32768.0
	var maxVal float64

	for i := start; i < end; i++ {
		val := float64(samples[i]) * invMaxSample
		if val < 0 {
			val = -val
		}
		if val > maxVal {
			maxVal = val
		}
	}

	return maxVal
}

// VU meter simulation - smooth, broadcast-style visualization
func calculateVU(samples []int16, start, end int) float64 {
	if end <= start {
		return 0
	}

	bucketSize := end - start
	const invMaxSample = 1.0 / 32768.0

	var sum float64

	// VU meters have a specific time constant and weighting
	for i := start; i < end; i++ {
		val := float64(samples[i]) * invMaxSample
		// Apply VU-style smoothing (less aggressive than RMS)
		sum += val * val * 0.8 // Slight compression for VU characteristics
	}

	if bucketSize > 0 {
		vu := fastSqrt(sum / float64(bucketSize))
		// Apply VU meter ballistics (smooth response)
		return vu * 1.2 // Slight boost for better visualization
	}

	return 0
}

// Dynamic range emphasis - highlights differences between loud and quiet
func calculateDynamic(samples []int16, start, end int) float64 {
	if end <= start {
		return 0
	}

	bucketSize := end - start
	const invMaxSample = 1.0 / 32768.0

	var sum, variance float64
	var mean float64

	// First pass: calculate mean
	for i := start; i < end; i++ {
		val := float64(samples[i]) * invMaxSample
		if val < 0 {
			val = -val
		}
		mean += val
	}
	mean /= float64(bucketSize)

	// Second pass: calculate variance (measure of dynamic range)
	for i := start; i < end; i++ {
		val := float64(samples[i]) * invMaxSample
		if val < 0 {
			val = -val
		}
		diff := val - mean
		variance += diff * diff
		sum += val * val
	}

	if bucketSize > 0 {
		rms := fastSqrt(sum / float64(bucketSize))
		dynamicFactor := fastSqrt(variance / float64(bucketSize))

		// Combine RMS with dynamic range factor
		// High variance = more dynamic = emphasized
		return rms * (1.0 + dynamicFactor*2.0)
	}

	return 0
}

// Smooth mode - heavily filtered for clean, minimal aesthetics
func calculateSmooth(samples []int16, start, end int) float64 {
	if end <= start {
		return 0
	}

	bucketSize := end - start
	const invMaxSample = 1.0 / 32768.0

	var sum float64
	var smoothedPrev float64 = 0.0
	const smoothingFactor = 0.95 // Heavy smoothing

	for i := start; i < end; i++ {
		val := float64(samples[i]) * invMaxSample
		if val < 0 {
			val = -val
		}

		// Apply exponential smoothing
		smoothedPrev = smoothingFactor*smoothedPrev + (1.0-smoothingFactor)*val
		sum += smoothedPrev * smoothedPrev
	}

	if bucketSize > 0 {
		smooth := fastSqrt(sum / float64(bucketSize))
		// Additional gentle compression for ultra-smooth appearance
		return smooth * 0.8
	}

	return 0
}

// Calculate loudness based on selected mode
func calculateLoudness(samples []int16, start, end int, mode string) float64 {
	switch mode {
	case "rms":
		return calculateRMS(samples, start, end)
	case "lufs":
		return calculateLUFS(samples, start, end)
	case "peak":
		return calculatePeak(samples, start, end)
	case "vu":
		return calculateVU(samples, start, end)
	case "dynamic":
		return calculateDynamic(samples, start, end)
	case "smooth":
		return calculateSmooth(samples, start, end)
	default:
		// Default to LUFS for unknown modes
		return calculateLUFS(samples, start, end)
	}
}

// Fast approximate square root using bit manipulation (Quake III algorithm variant)
func fastSqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Use bit manipulation for faster square root approximation
	i := *(*uint64)(unsafe.Pointer(&x))
	i = 0x5fe6eb50c7b537a9 - (i >> 1) // Magic number for double precision
	y := *(*float64)(unsafe.Pointer(&i))
	// One Newton-Raphson iteration for better accuracy
	y = y * (1.5 - 0.5*x*y*y)
	return 1.0 / y
}

// Command-line flags with default values
var (
	outputWidth  = flag.Int("width", 500, "Total SVG width in pixels")
	outputHeight = flag.Int("height", 80, "Total SVG height in pixels")
	bars         = flag.Int("bars", 100, "Number of bars in waveform")
	barSpacing   = flag.Int("spacing", 2, "Space between bars")
	barColor     = flag.String("color", "#3B82F6", "Bar color (hex)")
	cornerRadius = flag.Float64("radius", 8.0, "Bar corner radius")
	concurrent   = flag.Bool("concurrent", true, "Use concurrent processing for large files")
	calcMode     = flag.String("mode", "dynamic", "Calculation mode: 'rms', 'lufs', 'peak', 'vu', 'dynamic', 'smooth'")
)

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

func downsampleConcurrent(samples []int16, buckets int) []float64 {
	if len(samples) == 0 || buckets == 0 {
		return nil
	}

	samplesPerBucket := len(samples) / buckets
	if samplesPerBucket == 0 {
		samplesPerBucket = 1
	}

	// For small datasets, use sequential processing
	if len(samples) < 50000 { // Lowered threshold for better concurrency benefits
		return downsample(samples, buckets)
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

				// Use selected calculation mode
				peaks[bucket] = calculateLoudness(samples, startSample, endSample, *calcMode)
			}
		}(worker)
	}

	wg.Wait()
	return peaks
}

func downsample(samples []int16, buckets int) []float64 {
	if len(samples) == 0 || buckets == 0 {
		return nil
	}

	samplesPerBucket := len(samples) / buckets
	if samplesPerBucket == 0 {
		samplesPerBucket = 1
	}

	// Pre-allocate the exact size needed
	peaks := make([]float64, buckets)

	for bucket := 0; bucket < buckets; bucket++ {
		start := bucket * samplesPerBucket
		end := start + samplesPerBucket
		if end > len(samples) {
			end = len(samples)
		}

		// Use selected calculation mode
		peaks[bucket] = calculateLoudness(samples, start, end, *calcMode)
	}
	return peaks
}

func writeSVG(peaks []float64, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	ctx := canvas.NewContext(svg.New(file, float64(*outputWidth), float64(*outputHeight), nil))

	// Define colors for clean, flat design (no background)
	waveColor := canvas.Hex(*barColor)

	// Pre-calculate all constants
	barWidth := float64(*outputWidth) / float64(len(peaks))
	mid := float64(*outputHeight) / 2.0
	maxHeight := float64(*outputHeight) * 0.48
	barSpacingFloat := float64(*barSpacing)
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
	cornerRad := *cornerRadius
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

	// Important: Ensure SVG ends with a newline. Do not remove!
	file.Write([]byte("</svg>\n"))

	return nil
}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatalf("Usage: %s [options] input.mp3 output.svg\n", os.Args[0])
	}

	// Validate calculation mode
	validModes := map[string]bool{
		"rms":     true,
		"lufs":    true,
		"peak":    true,
		"vu":      true,
		"dynamic": true,
		"smooth":  true,
	}
	if !validModes[*calcMode] {
		log.Fatalf("Invalid mode '%s'. Valid modes are: rms, lufs, peak, vu, dynamic, smooth\n", *calcMode)
	}

	inputFile := flag.Arg(0)
	outputFile := flag.Arg(1)

	samples, err := readSamples(inputFile)
	if err != nil {
		log.Fatalf("Failed to read MP3: %v\n", err)
	}

	var peaks []float64
	if *concurrent {
		peaks = downsampleConcurrent(samples, *bars)
	} else {
		peaks = downsample(samples, *bars)
	}

	if err := writeSVG(peaks, outputFile); err != nil {
		log.Fatalf("Failed to write SVG: %v\n", err)
	}

	log.Printf("Waveform generated using %s mode: %s\n", *calcMode, outputFile)
}
