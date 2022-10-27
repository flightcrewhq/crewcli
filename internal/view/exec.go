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
	stdout := c.stdout.String()
	stderr := c.stderr.String()
	c.state.Stdout = stdout
	c.state.Stderr = stderr
	if err != nil {
		c.state.ErrorMessage = err.Error()
	}
	debug.Output("err: %v\nstdout (%d): %s\nstderr (%d): %s\n", err, c.stdout.Len(), stdout, c.stderr.Len(), stderr)
	return err
}
func (c *ExecCommand) SetStdin(r io.Reader) {
	c.cmd.Stdin = r
}
func (c *ExecCommand) SetStdout(w io.Writer) {
	debug.Output("set stdout")
	c.cmd.Stdout = &wrapOutput{
		real: w,
		buf:  &c.stdout,
	}
	debug.Output("stdout is %p", c.cmd.Stdout)
}
func (c *ExecCommand) SetStderr(w io.Writer) {
	debug.Output("set stderr")
	c.cmd.Stderr = &wrapOutput{
		real: w,
		buf:  &c.stderr,
	}
	debug.Output("stderr is %p", c.cmd.Stderr)
}

// Captures output to a buffer but also writes to the real io writer.
type wrapOutput struct {
	buf  *bytes.Buffer
	real io.Writer
}

func (wo *wrapOutput) Write(p []byte) (int, error) {
	debug.Output("Write %s for %p", string(p), wo)
	n, err := wo.buf.Write(p)
	debug.Output("Write to buf (%d, %v)", n, err)
	r, err := wo.real.Write(p)
	debug.Output("Write to real (%d, %v)", r, err)

	return n, err
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
