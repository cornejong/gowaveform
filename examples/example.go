// Package waveform provides comprehensive documentation and examples.
package main

import (
	"fmt"
	"log"

	"github.com/cornejong/gowaveform/waveform"
)

// ExampleBasicUsage demonstrates the simplest way to use the library
func ExampleBasicUsage() {
	// Create waveform with default settings - now supports multiple formats
	w, err := waveform.NewFromAudioFile("example.mp3", nil) // or .wav, .flac, .ogg, .aiff, .opus
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Save as SVG
	err = w.WriteSVG("basic_waveform.svg")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Basic waveform generated successfully")
}

// ExampleCustomConfig shows how to use custom configuration
func ExampleCustomConfig() {
	config := &waveform.Config{
		Width:        1000,
		Height:       150,
		Bars:         250,
		BarSpacing:   1,
		BarColor:     "#E74C3C",
		CornerRadius: 6.0,
		Concurrent:   true,
		Mode:         waveform.ModeLUFS,
	}

	w, err := waveform.NewFromAudioFile("example.mp3", config)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	err = w.WriteSVG("custom_waveform.svg")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Println("Custom waveform generated successfully")
}

// ExampleAllModes demonstrates all available calculation modes
func ExampleAllModes() {
	modes := []struct {
		mode waveform.CalculationMode
		desc string
	}{
		{waveform.ModeRMS, "Standard RMS calculation"},
		{waveform.ModeLUFS, "Perceptual loudness"},
		{waveform.ModePeak, "Peak detection"},
		{waveform.ModeVU, "VU meter style"},
		{waveform.ModeDynamic, "Dynamic range emphasis"},
		{waveform.ModeSmooth, "Smooth, minimal aesthetics"},
	}

	for _, m := range modes {
		config := waveform.DefaultConfig()
		config.Mode = m.mode
		config.BarColor = "#3498DB"

		w, err := waveform.NewFromAudioFile("example.mp3", config)
		if err != nil {
			log.Printf("Error with mode %s: %v", m.mode, err)
			continue
		}

		filename := fmt.Sprintf("waveform_%s.svg", m.mode)
		err = w.WriteSVG(filename)
		if err != nil {
			log.Printf("Error writing %s: %v", filename, err)
			continue
		}

		fmt.Printf("Generated %s (%s)\n", filename, m.desc)
	}
}

// ExampleMemoryGeneration shows how to generate SVG in memory
func ExampleMemoryGeneration() {
	config := waveform.DefaultConfig()
	config.Mode = waveform.ModeSmooth
	config.BarColor = "#9B59B6"

	w, err := waveform.NewFromAudioFile("example.mp3", config)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	// Generate SVG data without writing to file
	svgData, err := w.GenerateSVG()
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("Generated SVG in memory: %d bytes\n", len(svgData))

	// You can now use svgData for HTTP responses, embedding, etc.
	// For example: w.Header().Set("Content-Type", "image/svg+xml"); w.Write(svgData)
}

func main() {
	fmt.Println("Running GoWaveform Library Examples...")
	fmt.Println("=====================================")

	fmt.Println("\n1. Basic Usage:")
	ExampleBasicUsage()

	fmt.Println("\n2. Custom Configuration:")
	ExampleCustomConfig()

	fmt.Println("\n3. All Calculation Modes:")
	ExampleAllModes()

	fmt.Println("\n4. Memory Generation:")
	ExampleMemoryGeneration()

	fmt.Println("\nAll examples completed!")
}
