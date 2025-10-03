package main

import (
	"bufio"
	"errors"
	"os"
	"regexp"
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
			&cli.StringFlag{
				Name:    "strip",
				Usage:   "characters to strip before speaking",
				Aliases: []string{"s"},
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
			strip := c.String("strip")

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
					say := scanner.Text()
					if strip != "" {
						strips := strings.Split(strip, ",")
						for _, s := range strips {
							say = strings.ReplaceAll(say, s, "")
						}
					}

					say = removeEmoji(say)
					err := SayAnything(t, p, say)
					if err != nil {
						return cli.Exit(err, 1)
					}
				}

				if err := scanner.Err(); err != nil {
					return cli.Exit(err, 1)
				}
				return nil
			}

			say := text
			if strip != "" {
				strips := strings.Split(strip, ",")
				for _, s := range strips {
					say = strings.ReplaceAll(say, s, "")
				}
			}

			say = removeEmoji(say)
			return SayAnything(t, p, say)
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

func removeEmoji(str string) string {
	// Regex pattern to match most emoji characters
	emojiPattern := "[\U0001F600-\U0001F64F\U0001F300-\U0001F5FF\U0001F680-\U0001F6FF\U0001F700-\U0001F77F\U0001F780-\U0001F7FF\U0001F800-\U0001F8FF\U0001F900-\U0001F9FF\U00002702-\U000027B0\U000024C2-\U0001F251]+"
	re := regexp.MustCompile(emojiPattern)
	// Replace matched emoji with an empty string to remove it
	return re.ReplaceAllString(str, "")
}
