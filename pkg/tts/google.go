package tts

import (
	"context"
	"errors"

	"golang.org/x/text/language"
	"google.golang.org/api/option"
	gtts "cloud.google.com/go/texttospeech/apiv1"
	pb "cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
)

type Google struct {
	c *gtts.Client

	Lang string
	Name string
	// SsmlGender
	// CustomVoice
}

func NewGoogle(l, name string) *Google {
	ltag, _ := language.Parse(l)
	lang := ltag.String()

	return &Google{
		Lang: lang,
		Name: name,
	}
}

func (g *Google) Connect(keyfile string) error {
	ctx := context.Background()
	
	c, err := gtts.NewClient(ctx, option.WithCredentialsFile(keyfile))
	if err != nil {
		return err
	}

	g.c = c
	return nil
}

func (g *Google) Close() {
	g.c.Close()
}

func (g *Google) Speech(text string) ([]byte, error) {
	if g.c == nil {
		return nil, errors.New("no Google TTS client")
	}

	return g.speech(text)
}

func (g *Google) speech(text string) ([]byte, error) {
	ctx := context.Background()

	req := &pb.SynthesizeSpeechRequest{
		Input: &pb.SynthesisInput{
			InputSource: &pb.SynthesisInput_Text{
				Text: text,
			},
		},
		Voice: &pb.VoiceSelectionParams{
			LanguageCode: g.Lang,
			Name: g.Name,
		},
		AudioConfig: &pb.AudioConfig{
			AudioEncoding: pb.AudioEncoding_MP3,
			SpeakingRate: 1.0,
			Pitch: 0,
			VolumeGainDb: 0,
			// SampleRateHertz
			// EffectsProfileId
		},
	}
	resp, err := g.c.SynthesizeSpeech(ctx, req)
	if err != nil {
		return nil, err
	}
	
	return resp.AudioContent, nil
}
