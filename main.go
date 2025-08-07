package main

import (
	"flag"
	"log"
	"os"

	"github.com/cornejong/gowaveform/waveform"
)

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

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatalf("Usage: %s [options] input.{mp3|wav|flac|ogg|aiff|opus} output.svg\n", os.Args[0])
	}

	// Convert string mode to CalculationMode
	var mode waveform.CalculationMode
	switch *calcMode {
	case "rms":
		mode = waveform.ModeRMS
	case "lufs":
		mode = waveform.ModeLUFS
	case "peak":
		mode = waveform.ModePeak
	case "vu":
		mode = waveform.ModeVU
	case "dynamic":
		mode = waveform.ModeDynamic
	case "smooth":
		mode = waveform.ModeSmooth
	default:
		log.Fatalf("Invalid mode '%s'. Valid modes are: rms, lufs, peak, vu, dynamic, smooth\n", *calcMode)
	}

	inputFile := flag.Arg(0)
	outputFile := flag.Arg(1)

	// Create configuration from CLI flags
	config := &waveform.Config{
		Width:        *outputWidth,
		Height:       *outputHeight,
		Bars:         *bars,
		BarSpacing:   *barSpacing,
		BarColor:     *barColor,
		CornerRadius: *cornerRadius,
		Concurrent:   *concurrent,
		Mode:         mode,
	}

	// Generate waveform using the library
	w, err := waveform.NewFromAudioFile(inputFile, config)
	if err != nil {
		log.Fatalf("Failed to read audio file: %v\n", err)
	}

	if err := w.WriteSVG(outputFile); err != nil {
		log.Fatalf("Failed to write SVG: %v\n", err)
	}

	log.Printf("Waveform generated using %s mode: %s\n", *calcMode, outputFile)
}
