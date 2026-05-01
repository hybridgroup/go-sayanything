package tts

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
)

// Piper implements tts.Speaker using the Piper TTS engine. It requires the
// Piper binary to be installed and accessible in the system's PATH, and the
// appropriate ONNX model files to be available in the specified data directory.
type Piper struct {
	Lang string
	Name string

	datadir string
	gpu     bool
}

// NewPiper returns a Piper speaker for the specified language and voice. The
// language should be in the format "en-US" or "fr-FR", and the voice name should
// correspond to the available models for that language (e.g., "vits" or "fastpitch").
// The ONNX model file should be named in the format "<language>-<voice>.onnx".
func NewPiper(l, name string) *Piper {
	ltag, _ := language.Parse(l)
	lang := ltag.String()

	return &Piper{
		Lang: lang,
		Name: name,
	}
}

// Connect sets the data directory where the Piper ONNX model files are located. It does not
// establish any persistent connection, as Piper is invoked as a subprocess for each speech request.
func (p *Piper) Connect(datadir string) error {
	p.datadir = datadir
	return nil
}

// UseGPU enables or disables GPU acceleration for Piper. When enabled, Piper will attempt to use CUDA if available.
func (p *Piper) UseGPU(gpu bool) {
	p.gpu = gpu
}

// Close is a no-op for Piper; it satisfies the Speaker interface.
// Since Piper is invoked as a subprocess for each speech request, there are no persistent resources to clean up.
func (p *Piper) Close() {
}

// Speech converts the input text to speech using the Piper TTS engine. It constructs the appropriate
// command-line arguments based on the configured language, voice, data directory, and GPU settings,
// and captures the synthesized audio output as a byte slice.
func (p *Piper) Speech(text string) ([]byte, error) {
	lang := strings.Replace(p.Lang, "-", "_", -1)
	model := lang + "-" + p.Name + ".onnx"
	modelpath := filepath.Join(p.datadir, model)

	input := bytes.NewBufferString(text)
	var stdout, stderr bytes.Buffer

	cmds := []string{"--model", modelpath}
	if p.gpu {
		cmds = append(cmds, "--use-cuda")
	}
	cmds = append(cmds, []string{"--output-file", "-"}...)

	cmd := exec.Command("piper", cmds...)
	cmd.Stdin = input
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	return stdout.Bytes(), nil
}
