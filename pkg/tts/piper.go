package tts

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/text/language"
)

type Piper struct {
	Lang string
	Name string

	datadir string
	gpu     bool
}

func NewPiper(l, name string) *Piper {
	ltag, _ := language.Parse(l)
	lang := ltag.String()

	return &Piper{
		Lang: lang,
		Name: name,
	}
}

func (p *Piper) Connect(datadir string) error {
	p.datadir = datadir
	return nil
}

func (p *Piper) UseGPU(gpu bool) {
	p.gpu = gpu
}

func (p *Piper) Close() {
}

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
