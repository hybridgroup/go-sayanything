package say

import (
	"bytes"
	"errors"

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
	buf := bytes.NewReader(b)
	wv := wav.NewDecoder(buf)
	if wv == nil {
		return errors.New("failed to create wav decoder")
	}

	wv.ReadInfo()
	if wv.SampleRate == 0 {
		return errors.New("failed to read wav info")
	}

	if p.player == nil {
		player, _ := oto.NewContext(int(wv.SampleRate), int(wv.NumChans), 2, 8192)
		p.player = player.NewPlayer()
	}

	_, e := p.player.Write(b)
	return e
}
