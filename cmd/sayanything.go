package cmd

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

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
		Usage:     "play a google tts audio from the input text",
		Version:   version,
		UsageText: "sayanything <TEXT_TO_SAY>\n   echo \"TEXT_TO_SAY\" | sayanything",
		Authors: []*cli.Author{
			{
				Name:  "deadprogram",
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "lang",
				Usage:   "language of the text",
				Value:   "en",
				Aliases: []string{"l"},
			},
			&cli.StringFlag{
				Name:    "voice",
				Usage:   "voice to use to speak",
				Value:   "",
			},
			&cli.StringFlag{
				Name:    "keys",
				Usage:   "Google TTS keyfile",
				Value:   "",
				Aliases: []string{"k"},
			},
			&cli.BoolFlag{
				Name:    "slow",
				Usage:   "play audio slower",
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
			//slow := c.Bool("slow")
			lang := c.String("lang")
			voice := c.String("voice")
			keys := c.String("keys")
			if keys == "" {
				return cli.Exit(errors.New("keyfile required. use -k=/path/to/keys.json"), 1)
			}

			t := tts.NewGoogle(lang, voice)
			if err := t.Connect(keys); err != nil {
				return cli.Exit(err, 1)
			}

			defer t.Close()

			// input piped to stdin
			if isPiped() {
				scanner := bufio.NewScanner(os.Stdin)
				for scanner.Scan() {
					err := t.SayAnything(scanner.Text())
					if err != nil {
						return cli.Exit(err, 1)
					}
				}

				if err := scanner.Err(); err != nil {
					return cli.Exit(err, 1)
				}
				return nil
			}

			return t.SayAnything(text)
		},
	}

	if err := app.Run(os.Args); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func isPiped() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	notPipe := info.Mode()&os.ModeNamedPipe == 0
	return !notPipe
}
