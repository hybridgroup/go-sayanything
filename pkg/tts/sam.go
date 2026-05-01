package tts

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strings"

	"github.com/deadprogram/sam/config"
	"github.com/deadprogram/sam/reciter"
	"github.com/deadprogram/sam/render"
	"github.com/deadprogram/sam/sammain"
)

// Sam implements tts.Speaker using the SAM (Software Automatic Mouth)
// synthesiser. It requires no external service or data files.
type Sam struct {
	cfg *config.Config
}

// NewSam returns a Sam speaker with default synthesis parameters.
func NewSam() *Sam {
	return &Sam{cfg: config.DefaultConfig()}
}

// Connect is a no-op for SAM; it satisfies the Speaker interface.
func (s *Sam) Connect(_ string) error { return nil }

// Close is a no-op for SAM; it satisfies the Speaker interface.
func (s *Sam) Close() {}

// Speech converts text to 8-bit 22050 Hz mono PCM wrapped in a WAV container.
func (s *Sam) Speech(text string) ([]byte, error) {
	text = strings.ToUpper(strings.TrimSpace(text))
	if len(text) == 0 {
		return nil, nil
	}
	if len(text) > 255 {
		text = text[:255]
	}

	// Build the input buffer; TextToPhonemes expects a '[' terminator.
	var data [256]byte
	copy(data[:], []byte(text))
	data[len(text)] = '['

	rec := reciter.Reciter{}
	if !rec.TextToPhonemes(data[:], s.cfg) {
		return nil, errors.New("sam: failed to convert text to phonemes")
	}

	sam := sammain.Sam{Config: s.cfg}
	sam.SetInput(data)
	if !sam.SAMMain() {
		return nil, errors.New("sam: synthesis failed")
	}

	r := render.Render{Buffer: make([]byte, 22050*10)}
	sam.PrepareOutput(&r)

	return encodeWAV8(r.GetBuffer(), r.GetBufferLength())
}

// encodeWAV8 wraps raw 8-bit unsigned PCM (22050 Hz, mono) in a RIFF/WAV
// container. Only the first `length` bytes of pcm are included.
func encodeWAV8(pcm []byte, length int) ([]byte, error) {
	if length > len(pcm) {
		length = len(pcm)
	}
	data := pcm[:length]
	dataLen := uint32(len(data))

	var buf bytes.Buffer
	w := func(v interface{}) {
		binary.Write(&buf, binary.LittleEndian, v) //nolint:errcheck
	}

	// RIFF chunk
	buf.WriteString("RIFF")
	w(uint32(36 + dataLen)) // total file size minus 8
	buf.WriteString("WAVE")

	// fmt sub-chunk
	buf.WriteString("fmt ")
	w(uint32(16))    // chunk size
	w(uint16(1))     // PCM
	w(uint16(1))     // mono
	w(uint32(22050)) // sample rate
	w(uint32(22050)) // byte rate (sampleRate × channels × bitsPerSample/8)
	w(uint16(1))     // block align
	w(uint16(8))     // bits per sample

	// data sub-chunk
	buf.WriteString("data")
	w(dataLen)
	buf.Write(data)

	return buf.Bytes(), nil
}
