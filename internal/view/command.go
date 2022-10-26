package view

import (
	"strings"

	"flightcrew.io/cli/internal/style"
)

var (
	cmdReplacer = strings.NewReplacer(
		`\`, "",
		"\n", "",
		"\t", " ",
	)
)

type commandState struct {
	// If this command succeeds, then we should not run the actual command.
	SkipIfSucceed *commandState
	Command       string
	Description   string
	Link          string
	Ran           bool
	Succeeded     bool
}

func (s *commandState) View() string {
	var b strings.Builder

	b.WriteString(s.Description)
	b.WriteRune('\n')
	b.WriteString(s.Link)
	b.WriteRune('\n')
	b.WriteString("```sh\n")
	b.WriteString(s.Command)
	b.WriteString("\n```\n")

	out, err := style.Glamour.Render(b.String())
	if err != nil {
		return err.Error()
	}

	return out
}

func ViewError(stdout, stderr, errMessage string) string {
	var b strings.Builder
	var addNewline bool

	if len(stdout) > 0 {
		b.WriteString("## stdout\n\n```sh\n")
		b.WriteString(stdout)
		b.WriteString("```\n")
		addNewline = true
	}

	if len(stderr) > 0 {
		if addNewline {
			b.WriteRune('\n')
			addNewline = false
		}
		b.WriteString("## stderr\n\n```sh\n")
		b.WriteString(stderr)
		b.WriteString("```\n")
		addNewline = true
	}

	if len(errMessage) > 0 {
		if addNewline {
			b.WriteRune('\n')
			addNewline = false
		}
		b.WriteString(errMessage)
		b.WriteRune('\n')
	}

	out, err := style.Glamour.Render(b.String())
	if err != nil {
		return err.Error()
	}

	return out
}

func fmtCommandForExec(cmd string) string {
	return cmdReplacer.Replace(cmd)
}
