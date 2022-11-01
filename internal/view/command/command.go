package command

import (
	"bytes"
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerOutput string
	leftPadding  = lipgloss.NewStyle().PaddingLeft(2)
)

func init() {
	var err error
	if len(headerOutput) == 0 {
		headerOutput, err = style.Glamour.Render("## output\n")
		if err != nil {
			panic(err)
		}
	}
}

type State string

const (
	NoneState    State = "none"
	PromptState  State = "prompt"
	RunningState State = "running"
	SkipState    State = "skip"
	PassState    State = "pass"
	FailState    State = "fail"
)

type Type string

const (
	ReadType  Type = "read"
	WriteType Type = "write"
)

type Output struct {
	Log     string
	Message string
}

type Opts struct {
	Command     string
	Description string
	Message     map[State]string
	// If this read-only command succeeds, then we should not run the actual command.
	SkipIfSucceed *Model
}

type Model struct {
	state       State
	commandType Type
	opts        Opts
	output      Output
}

func NewReadModel(opts Opts) *Model {
	return &Model{
		opts:        opts,
		state:       NoneState,
		commandType: ReadType,
	}
}

func NewWriteModel(opts Opts) *Model {
	return &Model{
		opts:        opts,
		state:       NoneState,
		commandType: WriteType,
	}
}

func (m *Model) SetOutputLog(log string) {
	m.output.Log = log
}

func (m *Model) SetMessage(err error) {
	if err != nil {
		m.output.Message = err.Error()
		m.state = FailState
	} else {
		m.state = PassState
	}
}

func (m *Model) Complete(pass bool) {
	if pass {
		m.state = PassState
	} else {
		m.state = FailState
	}
	if msg := m.opts.Message[m.state]; len(msg) > 0 {
		m.output.Message = msg
	}
}

func (m *Model) Replace(replacer *strings.Replacer) {
	m.opts.Command = replacer.Replace(m.opts.Command)
	m.opts.Description = replacer.Replace(m.opts.Description)
}

func (m Model) State() State {
	return m.state
}

func (m Model) IsRead() bool {
	return m.commandType == ReadType
}

func (m *Model) GetCommandToRun() *wrappedCommand {
	m.state = RunningState
	return newWrappedCommand(m)
}

func (m *Model) Prompt() {
	m.state = PromptState
}

func (m *Model) ShouldPrompt() bool {
	if m.IsRead() {
		bashCommand := sanitizeForExec(m.opts.Command)
		c := exec.Command("bash", "-c", bashCommand) //nolint:gosec
		var b bytes.Buffer
		c.Stdout = &b
		c.Stderr = &b
		debug.Output("run `%s`", bashCommand)
		debug.Output("output: %s", b.String())
		err := c.Run()
		m.Complete(err == nil)
		debug.Output("error: %v", err)
		return false
	}

	if m.opts.SkipIfSucceed == nil {
		m.state = PromptState
		return true
	}

	state := m.opts.SkipIfSucceed.State()
	if state == PassState {
		m.output.Message = m.opts.SkipIfSucceed.opts.Message[PassState]
		m.state = SkipState
		return false
	}

	if state == NoneState {
		panic("prereq commands should go before the dependent commands")
	}

	m.state = PromptState
	return true
}
