package command

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/debug"
)

type wrappedCommand struct {
	model  *Model
	stdout bytes.Buffer
	stderr bytes.Buffer
	cmd    *exec.Cmd
}

func newWrappedCommand(m *Model) *wrappedCommand {
	c := exec.Command("bash", "-c", sanitizeForExec(m.opts.Command))
	return &wrappedCommand{
		cmd:   c,
		model: m,
	}
}
func (wc *wrappedCommand) Run() error {
	debug.Output("run command: %s", wc.model.opts.Command)
	err := wc.cmd.Run()
	wc.model.SetStdoutResult(wc.stdout.String())
	wc.model.SetStderrResult(wc.stderr.String())
	wc.model.SetMessage(err)
	debug.Output("err: %v\nstdout (%d): %s\nstderr (%d): %s\n", err, len(wc.model.output.Stdout), wc.model.output.Stdout, len(wc.model.output.Stderr), wc.model.output.Stderr)
	return err
}
func (wc *wrappedCommand) SetStdin(r io.Reader) {
	wc.cmd.Stdin = r
}
func (wc *wrappedCommand) SetStdout(w io.Writer) {
	wc.cmd.Stdout = io.MultiWriter(w, &wc.stdout)
}
func (wc *wrappedCommand) SetStderr(w io.Writer) {
	wc.cmd.Stderr = io.MultiWriter(w, &wc.stderr)
}

var (
	cmdReplacer = strings.NewReplacer(
		`\\n`, "",
		"\n", "",
		"\t", " ",
	)
)

func sanitizeForExec(cmd string) string {
	return cmdReplacer.Replace(cmd)
}
