package waveform

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-audio/aiff"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/go-mp3"
	"github.com/jfreymuth/oggvorbis"
	"github.com/mewkiz/flac"
	"github.com/pion/opus"
)

// AudioFormat represents the supported audio formats
type AudioFormat int

const (
	FormatMP3 AudioFormat = iota
	FormatWAV
	FormatFLAC
	FormatOGG
	FormatAIFF
	FormatOpus
	FormatUnknown
)

// String returns the string representation of the audio format
func (f AudioFormat) String() string {
	switch f {
	case FormatMP3:
		return "MP3"
	case FormatWAV:
		return "WAV"
	case FormatFLAC:
		return "FLAC"
	case FormatOGG:
		return "OGG"
	case FormatAIFF:
		return "AIFF"
	case FormatOpus:
		return "Opus"
	default:
		return "Unknown"
	}
}

// DetectFormat determines the audio format from the file extension
func DetectFormat(filename string) AudioFormat {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3":
		return FormatMP3
	case ".wav":
		return FormatWAV
	case ".flac":
		return FormatFLAC
	case ".ogg":
		return FormatOGG
	case ".aiff", ".aif":
		return FormatAIFF
	case ".opus":
		return FormatOpus
	default:
		return FormatUnknown
	}
}

// AudioDecoder interface for unified audio decoding
type AudioDecoder interface {
	Read([]byte) (int, error)
	SampleRate() int
	NumChannels() int
	Close() error
}

// MP3Decoder wraps go-mp3 decoder
type MP3Decoder struct {
	decoder *mp3.Decoder
	file    *os.File
}

func (d *MP3Decoder) Read(buf []byte) (int, error) {
	return d.decoder.Read(buf)
}

func (d *MP3Decoder) SampleRate() int {
	return d.decoder.SampleRate()
}

func (d *MP3Decoder) NumChannels() int {
	return 2 // MP3 is typically stereo
}

func (d *MP3Decoder) Close() error {
	return d.file.Close()
}

// WAVDecoder wraps go-audio/wav decoder
type WAVDecoder struct {
	decoder *wav.Decoder
	file    *os.File
	buffer  *audio.IntBuffer
}

func (d *WAVDecoder) Read(buf []byte) (int, error) {
	// Read PCM data using IntBuffer
	n, err := d.decoder.PCMBuffer(d.buffer)
	if err != nil && err != io.EOF {
		return 0, err
	}

	if n == 0 {
		return 0, io.EOF
	}

	// Convert int samples to int16 bytes
	bytesWritten := 0
	samples := d.buffer.Data
	for i := 0; i < len(samples) && bytesWritten < len(buf)-1; i++ {
		sample := int16(samples[i])
		buf[bytesWritten] = byte(sample)
		buf[bytesWritten+1] = byte(sample >> 8)
		bytesWritten += 2
	}

	return bytesWritten, err
}

func (d *WAVDecoder) SampleRate() int {
	return int(d.decoder.SampleRate)
}

func (d *WAVDecoder) NumChannels() int {
	return int(d.decoder.NumChans)
}

func (d *WAVDecoder) Close() error {
	return d.file.Close()
}

// FLACDecoder wraps mewkiz/flac decoder
type FLACDecoder struct {
	stream   *flac.Stream
	file     *os.File
	buffer   []int32
	pos      int
	finished bool
}

func (d *FLACDecoder) Read(buf []byte) (int, error) {
	if d.finished {
		return 0, io.EOF
	}

	bytesWritten := 0

	for bytesWritten < len(buf)-1 {
		// If we need more samples, read next frame
		if d.pos >= len(d.buffer) {
			frame, err := d.stream.ParseNext()
			if err != nil {
				if err == io.EOF {
					d.finished = true
				}
				return bytesWritten, err
			}

			// Get samples from first channel (convert to mono for simplicity)
			d.buffer = frame.Subframes[0].Samples
			d.pos = 0
		}

		// Convert samples to bytes
		for d.pos < len(d.buffer) && bytesWritten < len(buf)-1 {
			sample := int16(d.buffer[d.pos])
			buf[bytesWritten] = byte(sample)
			buf[bytesWritten+1] = byte(sample >> 8)
			bytesWritten += 2
			d.pos++
		}
	}

	return bytesWritten, nil
}

func (d *FLACDecoder) SampleRate() int {
	return int(d.stream.Info.SampleRate)
}

func (d *FLACDecoder) NumChannels() int {
	return int(d.stream.Info.NChannels)
}

func (d *FLACDecoder) Close() error {
	return d.file.Close()
}

// OGGDecoder wraps jfreymuth/oggvorbis decoder
type OGGDecoder struct {
	reader *oggvorbis.Reader
	file   *os.File
	format *oggvorbis.Format
}

func (d *OGGDecoder) Read(buf []byte) (int, error) {
	// Read float32 samples
	floatBuf := make([]float32, len(buf)/4) // Assuming stereo, 2 bytes per sample
	n, err := d.reader.Read(floatBuf)
	if err != nil {
		return 0, err
	}

	// Convert float32 to int16 bytes
	bytesWritten := 0
	for i := 0; i < n && bytesWritten < len(buf)-1; i++ {
		sample := int16(floatBuf[i] * 32767)
		buf[bytesWritten] = byte(sample)
		buf[bytesWritten+1] = byte(sample >> 8)
		bytesWritten += 2
	}

	return bytesWritten, err
}

func (d *OGGDecoder) SampleRate() int {
	return d.format.SampleRate
}

func (d *OGGDecoder) NumChannels() int {
	return d.format.Channels
}

func (d *OGGDecoder) Close() error {
	return d.file.Close()
}

// AIFFDecoder wraps go-audio/aiff decoder
type AIFFDecoder struct {
	decoder *aiff.Decoder
	file    *os.File
	buffer  *audio.IntBuffer
}

func (d *AIFFDecoder) Read(buf []byte) (int, error) {
	// Read PCM data using IntBuffer
	n, err := d.decoder.PCMBuffer(d.buffer)
	if err != nil && err != io.EOF {
		return 0, err
	}

	if n == 0 {
		return 0, io.EOF
	}

	// Convert int samples to int16 bytes
	bytesWritten := 0
	samples := d.buffer.Data
	for i := 0; i < len(samples) && bytesWritten < len(buf)-1; i++ {
		sample := int16(samples[i])
		buf[bytesWritten] = byte(sample)
		buf[bytesWritten+1] = byte(sample >> 8)
		bytesWritten += 2
	}

	return bytesWritten, err
}

func (d *AIFFDecoder) SampleRate() int {
	return int(d.decoder.SampleRate)
}

func (d *AIFFDecoder) NumChannels() int {
	return int(d.decoder.NumChans)
}

func (d *AIFFDecoder) Close() error {
	return d.file.Close()
}

// OpusDecoder wraps pion/opus decoder
type OpusDecoder struct {
	decoder  opus.Decoder
	file     *os.File
	buffer   []int16
	pos      int
	finished bool
}

func (d *OpusDecoder) Read(buf []byte) (int, error) {
	if d.finished {
		return 0, io.EOF
	}

	bytesWritten := 0

	for bytesWritten < len(buf)-1 {
		// If we need more samples, decode next packet
		if d.pos >= len(d.buffer) {
			// Read opus packet from file (this is simplified - real Opus files need proper packet parsing)
			packet := make([]byte, 1024)
			n, err := d.file.Read(packet)
			if err != nil {
				if err == io.EOF {
					d.finished = true
				}
				return bytesWritten, err
			}

			// Decode Opus packet to PCM
			pcmOut := make([]byte, 4096) // Output buffer for PCM data
			_, _, err = d.decoder.Decode(packet[:n], pcmOut)
			if err != nil {
				return bytesWritten, err
			}

			// Convert bytes to int16 samples
			samples := make([]int16, len(pcmOut)/2)
			for i := 0; i < len(pcmOut)-1; i += 2 {
				samples[i/2] = int16(pcmOut[i]) | int16(pcmOut[i+1])<<8
			}

			d.buffer = samples
			d.pos = 0
		}

		// Convert samples to bytes
		for d.pos < len(d.buffer) && bytesWritten < len(buf)-1 {
			sample := d.buffer[d.pos]
			buf[bytesWritten] = byte(sample)
			buf[bytesWritten+1] = byte(sample >> 8)
			bytesWritten += 2
			d.pos++
		}
	}

	return bytesWritten, nil
}

func (d *OpusDecoder) SampleRate() int {
	return 48000 // Opus native sample rate
}

func (d *OpusDecoder) NumChannels() int {
	return 1 // Simplified to mono for now
}

func (d *OpusDecoder) Close() error {
	return d.file.Close()
}

// NewAudioDecoder creates a new audio decoder based on the file format
func NewAudioDecoder(filename string) (AudioDecoder, error) {
	format := DetectFormat(filename)

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	switch format {
	case FormatMP3:
		decoder, err := mp3.NewDecoder(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return &MP3Decoder{decoder: decoder, file: file}, nil

	case FormatWAV:
		decoder := wav.NewDecoder(file)
		if !decoder.IsValidFile() {
			file.Close()
			return nil, fmt.Errorf("invalid WAV file")
		}
		// Create a buffer for PCM data
		buffer := &audio.IntBuffer{
			Format: &audio.Format{
				NumChannels: int(decoder.NumChans),
				SampleRate:  int(decoder.SampleRate),
			},
			Data: make([]int, 1024), // Initial buffer size
		}
		return &WAVDecoder{decoder: decoder, file: file, buffer: buffer}, nil

	case FormatFLAC:
		stream, err := flac.Parse(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return &FLACDecoder{
			stream:   stream,
			file:     file,
			buffer:   make([]int32, 0),
			pos:      0,
			finished: false,
		}, nil

	case FormatOGG:
		reader, err := oggvorbis.NewReader(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		format, err := oggvorbis.GetFormat(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		return &OGGDecoder{reader: reader, file: file, format: format}, nil

	case FormatAIFF:
		decoder := aiff.NewDecoder(file)
		if !decoder.IsValidFile() {
			file.Close()
			return nil, fmt.Errorf("invalid AIFF file")
		}
		// Create a buffer for PCM data
		buffer := &audio.IntBuffer{
			Format: &audio.Format{
				NumChannels: int(decoder.NumChans),
				SampleRate:  int(decoder.SampleRate),
			},
			Data: make([]int, 1024), // Initial buffer size
		}
		return &AIFFDecoder{decoder: decoder, file: file, buffer: buffer}, nil

	case FormatOpus:
		decoder := opus.NewDecoder()
		return &OpusDecoder{
			decoder:  decoder,
			file:     file,
			buffer:   make([]int16, 0),
			pos:      0,
			finished: false,
		}, nil

	default:
		file.Close()
		return nil, fmt.Errorf("unsupported audio format: %s", format)
	}
}

// readSamplesFromFormat reads audio samples from any supported format
func readSamplesFromFormat(path string) ([]int16, error) {
	decoder, err := NewAudioDecoder(path)
	if err != nil {
		return nil, err
	}
	defer decoder.Close()

	// Estimate capacity based on file size
	fileInfo, _ := os.Stat(path)
	estimatedSamples := int(fileInfo.Size() / 4) // Rough estimate
	pcm := make([]int16, 0, estimatedSamples)

	const bufferSize = 32768
	buf := make([]byte, bufferSize)

	for {
		n, err := decoder.Read(buf)
		if n == 0 || err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}

		// Convert bytes to int16 samples
		samples := make([]int16, n/2)
		for i := 0; i < n-1; i += 2 {
			samples[i/2] = int16(buf[i]) | int16(buf[i+1])<<8
		}
		pcm = append(pcm, samples...)
	}

	return pcm, nil
}
