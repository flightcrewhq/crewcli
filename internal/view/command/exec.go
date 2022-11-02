package command

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/debug"
)

type WrappedCommand struct {
	model          *Model
	cmd            *exec.Cmd
	combinedOutput bytes.Buffer
}

func newWrappedCommand(m *Model) *WrappedCommand {
	bashCommand := sanitizeForExec(m.opts.Command)
	c := exec.Command("bash", "-c", bashCommand)
	debug.Output("new wrapped command:\n  %s", bashCommand)
	return &WrappedCommand{
		cmd:   c,
		model: m,
	}
}
func (wc *WrappedCommand) Run() error {
	debug.Output("run command: %s", wc.model.opts.Command)
	err := wc.cmd.Run()
	wc.model.SetOutputLog(wc.combinedOutput.String())
	wc.model.SetMessage(err)
	debug.Output("err: %v\ncombined (%d): %s\n", err, len(wc.model.output.Log), wc.model.output.Log)
	return err
}
func (wc *WrappedCommand) SetStdin(r io.Reader) {
	wc.cmd.Stdin = r
}
func (wc *WrappedCommand) SetStdout(w io.Writer) {
	wc.cmd.Stdout = io.MultiWriter(w, &wc.combinedOutput)
}
func (wc *WrappedCommand) SetStderr(w io.Writer) {
	wc.cmd.Stderr = io.MultiWriter(w, &wc.combinedOutput)
}

var (
	cmdReplacer = strings.NewReplacer(
		`\\n`, "",
		"\t", " ",
	)
)

func sanitizeForExec(cmd string) string {
	return cmdReplacer.Replace(cmd)
}
