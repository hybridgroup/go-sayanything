package say

import (
	"github.com/hajimehoshi/oto"
	"github.com/tosone/minimp3"
)

type Player struct {
	player *oto.Player
}

func NewPlayer() *Player {
	return &Player{}
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

	dec, data, err := minimp3.DecodeFull(b)
	if err != nil {
		return err
	}

	if p.player == nil {
		player, _ := oto.NewContext(dec.SampleRate, dec.Channels, 2, 1024)
		p.player = player.NewPlayer()
	}

	_, err = p.player.Write(data)
	return err
}
