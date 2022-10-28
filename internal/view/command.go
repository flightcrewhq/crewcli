package view

import (
	"strings"

	"flightcrew.io/cli/internal/style"
)

type commandState string

var (
	headerStdout string
	headerStderr string
)

func init() {
	var err error
	if len(headerStdout) == 0 {
		headerStdout, err = style.Glamour.Render("## stdout\n")
		if err != nil {
			panic(err)
		}
	}
	if len(headerStderr) == 0 {
		headerStderr, err = style.Glamour.Render("## stderr\n")
		if err != nil {
			panic(err)
		}
	}
}

const (
	NoneState    commandState = "none"
	PromptState  commandState = "prompt"
	RunningState commandState = "running"
	SkipState    commandState = "skip"
	PassState    commandState = "pass"
	FailState    commandState = "fail"
)

type CommandState struct {
	Read         *ReadCommandState
	Mutate       *MutateCommandState
	State        commandState
	Command      string
	Description  string
	Stdout       string
	Stderr       string
	ErrorMessage string
}

type ReadCommandState struct {
	SuccessMessage, FailureMessage string
}

type MutateCommandState struct {
	// If this read-only command succeeds, then we should not run the actual command.
	SkipIfSucceed *CommandState
	Link          string
}

func (s *CommandState) View() string {
	var descB strings.Builder

	descB.WriteString(s.Description)
	descB.WriteRune('\n')
	if s.Mutate != nil && len(s.Mutate.Link) > 0 {
		descB.WriteString(s.Mutate.Link)
		descB.WriteRune('\n')
	}
	descB.WriteString("```sh\n")
	descB.WriteString(s.Command)
	descB.WriteString("\n```\n")

	desc, err := style.Glamour.Render(descB.String())
	if err != nil {
		return err.Error()
	}

	var outB strings.Builder
	outB.WriteString(desc)

	switch s.State {
	case NoneState:

	case SkipState:
		outB.WriteString("⏭ ")
		outB.WriteString(style.Bold("[SKIPPED] "))
		outB.WriteString(s.Mutate.SkipIfSucceed.Read.SuccessMessage)

	case RunningState:
		outB.WriteString("🚧 ")
		outB.WriteString("Running...\n")

	case PassState:
		outB.WriteString("✅ ")
		outB.WriteString(style.Success("[SUCCESS] "))
		if s.Read != nil && len(s.Read.SuccessMessage) > 0 {
			outB.WriteString(s.Read.SuccessMessage)
			outB.WriteRune('\n')
		} else {
			outB.WriteString("Command completed.\n")
		}

	case FailState:
		if s.Read != nil &&
			len(s.Read.FailureMessage) > 0 {
			outB.WriteString(s.Read.FailureMessage)
			outB.WriteRune('\n')
		} else {
			outB.WriteString("⛔️ ")
			outB.WriteString(style.Error("[ERROR] "))
			outB.WriteString("Command failed.\n\n")
		}
	}

	var addNewline bool
	stdout := s.Stdout
	stderr := s.Stderr

	if len(stdout) > 0 {
		outB.WriteString(headerStdout)
		outB.WriteRune('\n')
		outB.WriteString(stdout)
		outB.WriteRune('\n')
		addNewline = true
	}

	if len(stderr) > 0 {
		if addNewline {
			outB.WriteRune('\n')
			addNewline = false
		}
		outB.WriteString(headerStderr)
		outB.WriteRune('\n')
		outB.WriteString(stderr)
		outB.WriteRune('\n')
		addNewline = true
	}

	if len(s.ErrorMessage) > 0 {
		if addNewline {
			outB.WriteRune('\n')
			addNewline = false
		}
		outB.WriteString(s.ErrorMessage)
		outB.WriteRune('\n')
	}

	return outB.String()
}
