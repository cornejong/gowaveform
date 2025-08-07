package waveform

import "unsafe"

// calculateLoudness calculates loudness based on the selected mode
func calculateLoudness(samples []int16, start, end int, mode CalculationMode) float64 {
	switch mode {
	case ModeRMS:
		return calculateRMS(samples, start, end)
	case ModeLUFS:
		return calculateLUFS(samples, start, end)
	case ModePeak:
		return calculatePeak(samples, start, end)
	case ModeVU:
		return calculateVU(samples, start, end)
	case ModeDynamic:
		return calculateDynamic(samples, start, end)
	case ModeSmooth:
		return calculateSmooth(samples, start, end)
	default:
		// Default to LUFS for unknown modes
		return calculateLUFS(samples, start, end)
	}
}

// calculateLUFS implements LUFS-based loudness calculation for better perceptual representation
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

// calculateRMS implements traditional RMS calculation for standard waveform representation
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

// calculatePeak implements peak detection - fastest method, shows maximum amplitude
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

// calculateVU implements VU meter simulation - smooth, broadcast-style visualization
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

// calculateDynamic implements dynamic range emphasis - highlights differences between loud and quiet
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

// calculateSmooth implements smooth mode - heavily filtered for clean, minimal aesthetics
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

// fastSqrt implements fast approximate square root using bit manipulation (Quake III algorithm variant)
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
