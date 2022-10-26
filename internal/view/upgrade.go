package view

import (
	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/lipgloss"
)

/*
IP=$(gcloud compute instances list --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} | awk '/'$VIRTUAL_MACHINE'/ {print $5}')
if nc -w 1 -z $IP 22 ; then
  print_ask_exec \
    "gcloud compute instances update-container ${VIRTUAL_MACHINE} \\
    --project=${GOOGLE_PROJECT_ID} \\
    --zone=${ZONE} \\
    --container-image=\"${TOWER_IMAGE_PATH}:${TOWER_IMAGE_VERSION}\"" \
    "Update the image on the existing Flightcrew virtual machine named ${VIRTUAL_MACHINE}."
*/

var (
	requiredStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("199"))
	cursorStyle         = style.Focused.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	submitText     = "[  Submit  ]"
	confirmYesText = "[ Continue ]"
	confirmNoText  = "[   Edit   ]"
)

/*
type upgradeModel struct {
	vmName       string
	projectID    string
	zone         string
	towerVersion string

	imagePath string

	inputs     []wrappedInput
	focusIndex int

	commands []cmdState

	running bool

	confirming bool
	quitting   bool

	cursorMode textinput.CursorMode
}

type UpgradeModelParams struct {
}

func InitialUpgradeModel() upgradeModel {

	const defaultWidth = 20

	m := upgradeModel{
		projectID:    "",
		vmName:       "flightcrew-control-tower",
		zone:         "us-central",
		imagePath:    "us-west1-docker.pkg.dev/flightcrew-artifacts/client/tower",
		towerVersion: "latest",
		commands: []cmdState{
			{
				command: `gcloud compute instances update-container "${VIRTUAL_MACHINE}" \
  --project="${GOOGLE_PROJECT_ID}" \
  --zone="${ZONE}" \
  --container-image="${TOWER_IMAGE_PATH}:${TOWER_IMAGE_VERSION}"`,
				ran: false,
			},
		},
		inputs: make([]wrappedInput, 4),
	}

	for i := range m.inputs {
		input := wrappedInput{
			Input: textinput.New(),
		}
		input.Input.CursorStyle = cursorStyle
		input.Input.CharLimit = 32

		switch i {
		case 0:
			input.Input.Placeholder = "project-id-1234"
			input.Input.Focus()
			input.Input.PromptStyle = style.Focused
			input.Input.TextStyle = style.Focused
			input.Title = "Project ID"
			input.Required = true
		case 1:
			input.Input.Placeholder = "flightcrew-control-tower"
			input.Input.CharLimit = 64
			input.Title = "   VM Name"
			input.Default = "flightcrew-control-tower"
			input.Required = true
		case 2:
			input.Input.Placeholder = "us-central"
			input.Input.CharLimit = 32
			input.Title = "      Zone"
			input.Default = "us-central"
		case 3:
			input.Input.Placeholder = "latest"
			input.Title = "   Version"
			input.Default = "latest"
		}

		m.inputs[i] = input
	}

	return m
}

func (m upgradeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m upgradeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		// Change cursor mode
		case "ctrl+r":
			m.cursorMode = textinput.CursorBlink
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Input.SetCursorMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		// Set focus to next input
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// Did the user press enter while the submit button was focused?
			if s == "enter" && m.focusIndex == len(m.inputs) {
				if !m.confirming {
					m.confirming = true
					return m, nil
				}

				for i, input := range m.inputs {
					fmt.Printf("input %d: %s", i, input.Input.Value())
				}
				return m, tea.Quit
			}

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				if !m.confirming || m.focusIndex > len(m.inputs)+1 {
					m.focusIndex = 0
				}
			} else if m.focusIndex < 0 {
				if m.confirming {
					m.focusIndex = len(m.inputs) + 1

				} else {
					m.focusIndex = len(m.inputs)
				}
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					// Set focused state
					cmds[i] = m.inputs[i].Input.Focus()
					m.inputs[i].Input.PromptStyle = style.Focused
					m.inputs[i].Input.TextStyle = style.Focused
					continue
				}
				// Remove focused state
				m.inputs[i].Input.Blur()
				m.inputs[i].Input.PromptStyle = style.None
				m.inputs[i].Input.TextStyle = style.None
			}

			return m, tea.Batch(cmds...)
		}
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *upgradeModel) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i].Input, cmds[i] = m.inputs[i].Input.Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m upgradeModel) View() string {
	var b strings.Builder

	for i := range m.inputs {
		b.WriteString(m.inputs[i].Title)
		if m.inputs[i].Required {
			b.WriteString(requiredStyle.Render("*"))
		} else {
			b.WriteRune(' ')
		}
		b.WriteString(": ")
		b.WriteString(m.inputs[i].Input.View())
		if len(m.inputs[i].HelpText) > 0 {
			b.WriteRune('\n')
			b.WriteString(helpStyle.Render(m.inputs[i].HelpText))
		}
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	b.WriteString("\n\n")

	if m.confirming {
		if m.focusIndex == len(m.inputs) {
			b.WriteString(style.Focused.Render(confirmYesText))
		} else {
			b.WriteString(style.Blurred.Render(confirmYesText))
		}

		b.WriteRune(' ')

		if m.focusIndex == len(m.inputs)+1 {
			b.WriteString(style.Focused.Render(confirmNoText))
		} else {
			b.WriteString(style.Blurred.Render(confirmNoText))
		}
	} else {
		if m.focusIndex == len(m.inputs) {
			b.WriteString(style.Focused.Render(submitText))
		} else {
			b.WriteString(style.Blurred.Render(submitText))
		}
	}

	b.WriteString("\n\n")

	b.WriteString(requiredStyle.Render("*"))
	b.WriteString(" - required\n\n")
	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}

*/
