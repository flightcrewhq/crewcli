package command

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/debug"
)

type wrappedCommand struct {
	model          *Model
	combinedOutput bytes.Buffer
	cmd            *exec.Cmd
}

func newWrappedCommand(m *Model) *wrappedCommand {
	bashCommand := sanitizeForExec(m.opts.Command)
	c := exec.Command("bash", "-c", bashCommand)
	debug.Output("new wrapped command:\n  %s", bashCommand)
	return &wrappedCommand{
		cmd:   c,
		model: m,
	}
}
func (wc *wrappedCommand) Run() error {
	debug.Output("run command: %s", wc.model.opts.Command)
	err := wc.cmd.Run()
	wc.model.SetOutputLog(wc.combinedOutput.String())
	wc.model.SetMessage(err)
	debug.Output("err: %v\ncombined (%d): %s\n", err, len(wc.model.output.Log), wc.model.output.Log)
	return err
}
func (wc *wrappedCommand) SetStdin(r io.Reader) {
	wc.cmd.Stdin = r
}
func (wc *wrappedCommand) SetStdout(w io.Writer) {
	wc.cmd.Stdout = io.MultiWriter(w, &wc.combinedOutput)
}
func (wc *wrappedCommand) SetStderr(w io.Writer) {
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
