package view

import (
	"bytes"
	"io"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/debug"
)

type ExecCommand struct {
	state  *CommandState
	stdout bytes.Buffer
	stderr bytes.Buffer
	cmd    *exec.Cmd
}

func NewExecCommand(s *CommandState) *ExecCommand {
	c := exec.Command("bash", "-c", fmtCommandForExec(s.Command))
	return &ExecCommand{
		cmd:   c,
		state: s,
	}
}
func (c *ExecCommand) Run() error {
	debug.Output("run command: %s", c.state.Command)
	err := c.cmd.Run()
	c.state.Stdout = c.stdout.String()
	c.state.Stderr = c.stderr.String()
	if err != nil {
		c.state.ErrorMessage = err.Error()
	}
	debug.Output("err: %v\nstdout (%d): %s\nstderr (%d): %s\n", err, len(c.state.Stdout), c.state.Stdout, len(c.state.Stderr), c.state.Stderr)
	return err
}
func (c *ExecCommand) SetStdin(r io.Reader) {
	c.cmd.Stdin = r
}
func (c *ExecCommand) SetStdout(w io.Writer) {
	c.cmd.Stdout = io.MultiWriter(w, &c.stdout)
}
func (c *ExecCommand) SetStderr(w io.Writer) {
	c.cmd.Stderr = io.MultiWriter(w, &c.stderr)
}

var (
	cmdReplacer = strings.NewReplacer(
		`\`, "",
		"\n", "",
		"\t", " ",
	)
)

func fmtCommandForExec(cmd string) string {
	return cmdReplacer.Replace(cmd)
}
