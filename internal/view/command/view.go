package command

import (
	"strings"

	"flightcrew.io/cli/internal/style"
)

func (m Model) viewDescription(b *strings.Builder) {
	var out strings.Builder
	out.WriteString(m.opts.Description)
	out.WriteString("\n\n```sh\n")
	out.WriteString(m.opts.Command)
	out.WriteString("\n```\n")

	desc, err := style.Glamour.Render(out.String())
	if err != nil {
		b.WriteString(err.Error())
		return
	}

	b.WriteString(desc)
}

func (m Model) viewTagline(b *strings.Builder) {
	var out strings.Builder
	switch m.state {
	case NoneState:

	case SkipState:
		out.WriteString("⏭ ")
		out.WriteString(style.Bold("[SKIPPED] "))
		out.WriteString(m.output.Message)

	case RunningState:
		out.WriteString("🚧 ")
		out.WriteString("Running...\n")

	case PassState:
		out.WriteString("✅ ")
		out.WriteString(style.Success("[SUCCESS] "))
		if len(m.output.Message) > 0 {
			out.WriteString(m.output.Message)
			out.WriteRune('\n')
		} else {
			out.WriteString("Command completed.\n")
		}

	case FailState:
		if m.IsRead() {
			out.WriteString("💡")
			out.WriteString(style.Bold("[INFO] "))
		} else {
			out.WriteString("⛔️ ")
			out.WriteString(style.Error("[ERROR] "))
		}
		out.WriteString(m.output.Message)
		out.WriteString("\n")
	}
	b.WriteString(leftPadding.Render(out.String()))
}

func (m Model) viewOutput(b *strings.Builder) {
	var addNewline bool
	if len(m.output.Stdout) > 0 {
		b.WriteString(headerStdout)
		b.WriteRune('\n')
		b.WriteString(m.output.Stdout)
		b.WriteRune('\n')
		addNewline = true
	}

	if len(m.output.Stderr) > 0 {
		if addNewline {
			b.WriteRune('\n')
			addNewline = false
		}
		b.WriteString(headerStderr)
		b.WriteRune('\n')
		b.WriteString(m.output.Stderr)
		b.WriteRune('\n')
	}
}

func (m Model) View() string {
	var out strings.Builder
	m.viewDescription(&out)
	m.viewTagline(&out)
	m.viewOutput(&out)
	return out.String()
}
