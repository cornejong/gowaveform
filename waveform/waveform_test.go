package waveform

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Width != 500 {
		t.Errorf("Expected width 500, got %d", config.Width)
	}

	if config.Height != 80 {
		t.Errorf("Expected height 80, got %d", config.Height)
	}

	if config.Mode != ModeDynamic {
		t.Errorf("Expected mode %s, got %s", ModeDynamic, config.Mode)
	}
}

func TestCalculationModes(t *testing.T) {
	modes := []CalculationMode{
		ModeRMS,
		ModeLUFS,
		ModePeak,
		ModeVU,
		ModeDynamic,
		ModeSmooth,
	}

	// Test with dummy samples
	samples := make([]int16, 1000)
	for i := range samples {
		samples[i] = int16(i % 100)
	}

	for _, mode := range modes {
		result := calculateLoudness(samples, 0, len(samples), mode)
		if result < 0 {
			t.Errorf("Mode %s returned negative value: %f", mode, result)
		}
	}
}

func TestNewFromSamples(t *testing.T) {
	// Create dummy samples
	samples := make([]int16, 1000)
	for i := range samples {
		samples[i] = int16(i % 1000)
	}

	config := DefaultConfig()
	config.Bars = 50

	w := NewFromSamples(samples, config)

	if w == nil {
		t.Fatal("NewFromSamples returned nil")
	}

	if len(w.Peaks) != 50 {
		t.Errorf("Expected 50 peaks, got %d", len(w.Peaks))
	}

	if w.Config.Bars != 50 {
		t.Errorf("Expected 50 bars in config, got %d", w.Config.Bars)
	}
}

func TestGenerateSVG(t *testing.T) {
	// Create dummy samples
	samples := make([]int16, 1000)
	for i := range samples {
		samples[i] = int16(i % 500)
	}

	config := DefaultConfig()
	config.Width = 400
	config.Height = 60
	config.Bars = 40

	w := NewFromSamples(samples, config)

	svgData, err := w.GenerateSVG()
	if err != nil {
		t.Fatalf("GenerateSVG failed: %v", err)
	}

	if len(svgData) == 0 {
		t.Error("Generated SVG data is empty")
	}

	// Check that it contains basic SVG structure
	svgStr := string(svgData)
	if !containsString(svgStr, "<svg") {
		t.Error("SVG data doesn't contain <svg tag")
	}

	if !containsString(svgStr, "</svg>") {
		t.Error("SVG data doesn't contain closing </svg> tag")
	}
}

func TestWriteSVG(t *testing.T) {
	// Create dummy samples
	samples := make([]int16, 500)
	for i := range samples {
		samples[i] = int16(i % 200)
	}

	config := DefaultConfig()
	config.Bars = 25

	w := NewFromSamples(samples, config)

	filename := "test_output.svg"
	defer os.Remove(filename) // Clean up

	err := w.WriteSVG(filename)
	if err != nil {
		t.Fatalf("WriteSVG failed: %v", err)
	}

	// Check that file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("SVG file was not created")
	}
}

func TestInvalidInputs(t *testing.T) {
	// Test with empty samples
	config := DefaultConfig()
	w := NewFromSamples([]int16{}, config)

	if w.Peaks != nil && len(w.Peaks) > 0 {
		t.Error("Expected no peaks for empty samples")
	}

	// Test with zero buckets
	samples := make([]int16, 100)
	config.Bars = 0
	w = NewFromSamples(samples, config)

	if w.Peaks != nil && len(w.Peaks) > 0 {
		t.Error("Expected no peaks for zero buckets")
	}
}

// Helper function since Go doesn't have strings.Contains in older versions
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
