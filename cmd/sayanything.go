package cmd

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/hybridgroup/go-sayanything/pkg/say"
	"github.com/hybridgroup/go-sayanything/pkg/tts"
)

var (
	// Version placeholder for the version number filled by goreleaser
	Version = ""
)

// RunCLI runs the CLI command
func RunCLI(version string) error {
	app := &cli.App{
		Name:      "sayanything",
		Usage:     "play text to speech audio from the input text",
		Version:   version,
		UsageText: "sayanything <TEXT_TO_SAY>\n   echo \"TEXT_TO_SAY\" | sayanything",
		Authors: []*cli.Author{
			{
				Name: "deadprogram",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Usage:   "language of the text",
				Value:   "en-us",
				Aliases: []string{"l"},
			},
			&cli.StringFlag{
				Name:  "voice",
				Usage: "voice to use to speak",
				Value: "",
			},
			&cli.StringFlag{
				Name:    "keys",
				Usage:   "Google TTS keyfile",
				Value:   "",
				Aliases: []string{"k"},
			},
			&cli.StringFlag{
				Name:    "engine",
				Usage:   "TTS engine to use (google, piper)",
				Aliases: []string{"e"},
			},
			&cli.StringFlag{
				Name:  "data",
				Usage: "data directory for the TTS engine",
			},
			&cli.BoolFlag{
				Name:  "gpu",
				Usage: "use GPU for TTS engine",
			},
		},
		Before: func(c *cli.Context) error {
			if c.NArg() == 0 && !isPiped() {
				return cli.Exit("missing text to play", 1)
			}
			return nil
		},
		Action: func(c *cli.Context) error {
			text := strings.Join(c.Args().Slice(), " ")
			lang := c.String("lang")
			voice := c.String("voice")
			keys := c.String("keys")

			var t tts.Speaker
			var format string
			switch c.String("engine") {
			case "piper":
				t = tts.NewPiper(lang, voice)
				if err := t.Connect(c.String("data")); err != nil {
					return cli.Exit(err, 1)
				}
				if c.Bool("gpu") {
					t.(*tts.Piper).UseGPU(true)
				}
				format = "wav"
			case "google":
				if keys == "" {
					return cli.Exit(errors.New("keyfile required. use -k=/path/to/keys.json"), 1)
				}

				t = tts.NewGoogle(lang, voice)
				if err := t.Connect(keys); err != nil {
					return cli.Exit(err, 1)
				}
				format = "mp3"
			default:
				return cli.Exit(errors.New("unsupported engine"), 1)
			}

			defer t.Close()

			p := say.NewPlayer(format)
			defer p.Close()

			// input piped to stdin
			if isPiped() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					err := SayAnything(t, p, scanner.Text())
					if err != nil {
						return cli.Exit(err, 1)
					}
				}

				if err := scanner.Err(); err != nil {
					return cli.Exit(err, 1)
				}
				return nil
			}

			return SayAnything(t, p, text)
		},
	}

	if err := app.Run(os.Args); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func SayAnything(t tts.Speaker, p *say.Player, text string) error {
	if len(text) == 0 {
		return nil
	}

	data, err := t.Speech(text)
	if err != nil {
		return err
	}

	return p.Say(data)
}

func isPiped() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	notPipe := info.Mode()&os.ModeNamedPipe == 0
	return !notPipe
}
