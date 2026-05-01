package say

import (
	"bytes"
	"errors"
	"io"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/tosone/minimp3"
)

type Player struct {
	ctx          *oto.Context
	sampleRate   int
	channelCount int
	audioFormat  oto.Format
	format       string
}

func NewPlayer(format string) *Player {
	return &Player{format: format}
}

func (p *Player) Close() {}

func (p *Player) Say(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	switch p.format {
	case "mp3":
		return p.sayMP3(b)
	case "wav":
		return p.sayWAV(b)
	default:
		return errors.New("unsupported format")
	}
}

// ensureContext creates or reuses an oto.Context for the given audio parameters.
func (p *Player) ensureContext(sampleRate, channelCount int, format oto.Format) error {
	if p.ctx != nil && p.sampleRate == sampleRate && p.channelCount == channelCount && p.audioFormat == format {
		return nil
	}
	op := &oto.NewContextOptions{
		SampleRate:   sampleRate,
		ChannelCount: channelCount,
		Format:       format,
	}
	ctx, ready, err := oto.NewContext(op)
	if err != nil {
		return err
	}
	<-ready
	p.ctx = ctx
	p.sampleRate = sampleRate
	p.channelCount = channelCount
	p.audioFormat = format
	return nil
}

// playPCM sends raw PCM bytes to the audio device and blocks until done.
//
// IsPlaying() returns false when oto's internal software buffer is empty, but
// the ALSA ring buffer and PulseAudio still hold ~400 ms of audio at that
// point.  We pad the stream with 500 ms of silence so that IsPlaying() only
// becomes false after all real audio has cleared every hardware stage.
func (p *Player) playPCM(pcm []byte) {
	// bytesPerSample for the current format (always int16 after our conversion)
	bytesPerSample := 2                                              // FormatSignedInt16LE
	silenceLen := p.sampleRate * p.channelCount * bytesPerSample / 2 // 500 ms
	silence := make([]byte, silenceLen)
	src := io.MultiReader(bytes.NewReader(pcm), bytes.NewReader(silence))

	player := p.ctx.NewPlayer(src)
	defer player.Close()
	player.Play()
	for player.IsPlaying() {
		time.Sleep(time.Millisecond)
	}
}

func (p *Player) sayMP3(b []byte) error {
	dec, data, err := minimp3.DecodeFull(b)
	if err != nil {
		return err
	}
	if err := p.ensureContext(dec.SampleRate, dec.Channels, oto.FormatSignedInt16LE); err != nil {
		return err
	}
	p.playPCM(data)
	return nil
}

func (p *Player) sayWAV(b []byte) error {
	dec := wav.NewDecoder(bytes.NewReader(b))
	if dec == nil {
		return errors.New("failed to create wav decoder")
	}
	dec.ReadInfo()
	if dec.SampleRate == 0 {
		return errors.New("failed to read wav info")
	}

	// oto v3's FormatUnsignedInt8 has a Go byte-arithmetic bug (v8-(1<<7)
	// wraps unsigned, producing inverted output). Always decode to signed
	// 16-bit LE, which oto handles correctly.
	if err := p.ensureContext(int(dec.SampleRate), int(dec.NumChans), oto.FormatSignedInt16LE); err != nil {
		return err
	}

	srcBytesPerSample := int(dec.BitDepth) / 8
	if srcBytesPerSample < 1 {
		srcBytesPerSample = 1
	}
	pcmBuf := &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  int(dec.SampleRate),
			NumChannels: int(dec.NumChans),
		},
		Data: make([]int, 4096),
	}

	var raw []byte
	for {
		n, err := dec.PCMBuffer(pcmBuf)
		if n == 0 {
			break
		}
		out := make([]byte, n*2) // always output signed 16-bit LE
		for i := 0; i < n; i++ {
			v := pcmBuf.Data[i]
			var s16 int16
			switch dec.BitDepth {
			case 8:
				// go-audio/wav returns unsigned 0-255; centre at 128, scale to 16-bit
				s16 = int16((v - 128) * 256)
			default:
				// 16-bit (and others): already signed, keep as-is
				s16 = int16(v)
			}
			out[i*2] = byte(s16)
			out[i*2+1] = byte(s16 >> 8)
		}
		raw = append(raw, out...)
		if err != nil {
			break
		}
	}

	p.playPCM(raw)
	return nil
}
