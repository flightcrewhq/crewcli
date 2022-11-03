package command

import (
	"strings"
)

func (m Model) String() string {
	var out strings.Builder
	out.WriteString(m.opts.Description)
	out.WriteString("\n\n```sh\n")
	out.WriteString(m.opts.Command)
	out.WriteString("\n```\n\n")

	switch m.state {
	case NoneState:

	case SkipState:
		out.WriteString("⏭ ")
		out.WriteString("[SKIPPED] ")
		out.WriteString(m.output.Message)
		out.WriteRune('\n')

	case RunningState:
		out.WriteString("🚧 ")
		out.WriteString("Running...\n")

	case PassState:
		out.WriteString("✅ [SUCCESS]")
		if len(m.output.Message) > 0 {
			out.WriteString(m.output.Message)
			out.WriteRune('\n')
		} else {
			out.WriteString("Command completed.\n")
		}

	case FailState:
		if m.IsRead() {
			out.WriteString("💡 [INFO] ")
		} else {
			out.WriteString("⛔️ [ERROR] ")
		}
		out.WriteString(m.output.Message)
		out.WriteString("\n")
	}

	if len(m.output.Log) > 0 {
		out.WriteString(headerOutput)
		out.WriteRune('\n')
		out.WriteString(m.output.Log)
		out.WriteRune('\n')
	}

	return out.String()
}
