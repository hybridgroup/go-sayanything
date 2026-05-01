package say

import (
	"bytes"
	"errors"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/hajimehoshi/oto"
	"github.com/tosone/minimp3"
)

type Player struct {
	player *oto.Player
	format string
}

func NewPlayer(format string) *Player {
	return &Player{
		format: format,
	}
}

func (p *Player) Close() {
	if p.player != nil {
		p.player.Close()
		p.player = nil
	}
}

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

func (p *Player) sayMP3(b []byte) error {
	dec, data, err := minimp3.DecodeFull(b)
	if err != nil {
		return err
	}

	if p.player == nil {
		player, _ := oto.NewContext(dec.SampleRate, dec.Channels, 2, 8192)
		p.player = player.NewPlayer()
	}

	_, err = p.player.Write(data)
	return err
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

	bytesPerSample := int(dec.BitDepth) / 8
	if bytesPerSample < 1 {
		bytesPerSample = 1
	}

	if p.player == nil {
		ctx, err := oto.NewContext(int(dec.SampleRate), int(dec.NumChans), bytesPerSample, 8192)
		if err != nil {
			return err
		}
		p.player = ctx.NewPlayer()
	}

	pcmBuf := &audio.IntBuffer{
		Format: &audio.Format{
			SampleRate:  int(dec.SampleRate),
			NumChannels: int(dec.NumChans),
		},
		Data: make([]int, 4096),
	}

	raw := make([]byte, 4096*bytesPerSample)
	for {
		n, err := dec.PCMBuffer(pcmBuf)
		if n == 0 {
			break
		}

		out := raw[:n*bytesPerSample]
		for i := 0; i < n; i++ {
			v := pcmBuf.Data[i]
			switch bytesPerSample {
			case 1:
				out[i] = byte(v)
			case 2:
				out[i*2] = byte(v)
				out[i*2+1] = byte(v >> 8)
			}
		}

		if _, werr := p.player.Write(out); werr != nil {
			return werr
		}

		if err != nil { // io.EOF after last chunk
			break
		}
	}

	return nil
}
