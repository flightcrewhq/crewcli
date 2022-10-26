package view

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	defaultWidth = 80
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
)

type cmdFinishedErr struct {
	err error
}

type runState string

const (
	stateRun     runState = "run"
	statePrompt  runState = "prompt"
	stateSuccess runState = "success"
	stateFailure runState = "failure"
)

type runModel struct {
	params map[string]string

	commands       []*commandState
	currentIndex   int
	state          runState
	cmdOut, cmdErr bytes.Buffer

	userInput bool
	yesButton *Button
	noButton  *Button

	spinner spinner.Model

	// Used when state is Failure
	errMessage string
	cursorMode textinput.CursorMode
}

type runModelParams struct {
}

func NewRunModel(params map[string]string) *runModel {
	const defaultWidth = 20
	spin := spinner.New()
	spin.Style = spinnerStyle
	spin.Spinner = spinner.Line
	spin.Spinner.FPS = time.Second / 5

	yesButton, err := NewButton("yes", 10)
	_ = err
	noButton, err := NewButton("no", 10)
	_ = err

	checkIAMRole := &commandState{
		Command: `gcloud iam roles describe --project="${GOOGLE_PROJECT_ID}" "${IAM_ROLE}" >/dev/null 2>&1`,
	}

	m := &runModel{
		params:  params,
		state:   statePrompt,
		spinner: spin,
		commands: []*commandState{
			{
				SkipIfSucceed: &commandState{
					Command: `gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" > /dev/null 2>&1`,
				},
				Command: `gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
	--project="${GOOGLE_PROJECT_ID}" \
	--display-name="${SERVICE_ACCOUNT}" \
	--description="Runs Flightcrew's Control Tower VM."`,
				Ran:         false,
				Description: `This command will create a service account, and follow-up commands will attach read and/or write permissions.`,
				Link:        "https://cloud.google.com/iam/docs/creating-managing-service-accounts",
			},
			{
				SkipIfSucceed: checkIAMRole,
				Command: `gcloud iam roles create ${IAM_ROLE} \
	--project=${GOOGLE_PROJECT_ID} \
	--file=${IAM_FILE}
`,
				Description: `This command creates an IAM role from ${IAM_FILE} for the Flightcrew VM to access configs and monitoring data.`,
				Link:        "https://cloud.google.com/iam/docs/understanding-custom-roles",
				Ran:         false,
			},
			{
				SkipIfSucceed: checkIAMRole,
				Command: `gcloud projects add-iam-policy-binding "${GOOGLE_PROJECT_ID}" \\
	--member=serviceAccount:"${SERVICE_ACCT_EMAIL}" \
	--role="projects/${GOOGLE_PROJECT_ID}/roles/${IAM_ROLE}" \
	--condition=None`,
				Description: "This command attaches the IAM role to Flightcrew's service account, which will give the IAM permissions to a new VM.",
				Link:        "https://cloud.google.com/iam/docs/granting-changing-revoking-access",
			},
		},
		yesButton: yesButton,
		noButton:  noButton,
	}

	replaceArgs := make([]string, 0, 2*len(params))
	for key, param := range params {
		replaceArgs = append(replaceArgs, key, param)
	}
	replacer := strings.NewReplacer(replaceArgs...)

	for i := range m.commands {
		m.commands[i].Command = replacer.Replace(m.commands[i].Command)
		m.commands[i].Description = replacer.Replace(m.commands[i].Description)
		if m.commands[i].SkipIfSucceed != nil {
			m.commands[i].SkipIfSucceed.Command = replacer.Replace(m.commands[i].SkipIfSucceed.Command)
		}
	}

	// if params.VirtualMachineName == "" {
	// 	params.VirtualMachineName = "flightcrew-control-tower"
	// }

	// if params.Zone == "" {
	// 	params.Zone = "us-central"
	// }

	// if params.TowerVersion == "" {
	// 	params.TowerVersion = "latest"
	// }
	m.nextCommand()

	return m
}

func (m *runModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *runModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Allow user to quit at any time.
		case "ctrl+c", "esc":
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	switch m.state {
	case statePrompt:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if m.state == statePrompt && !m.userInput {
					// TODO: Confirm quitting
					return m, tea.Quit
				}

				if m.state == statePrompt && m.userInput {
					m.state = stateRun

					c := exec.Command("bash", "-c", fmtCommandForExec(m.commands[m.currentIndex].Command)) //nolint:gosec
					c.Stdout = &m.cmdOut
					c.Stderr = &m.cmdErr
					return m, tea.ExecProcess(c, func(err error) tea.Msg {
						return cmdFinishedErr{err}
					})
				}

			case "tab":
				if m.state == statePrompt {
					m.userInput = !m.userInput
				}

			case "left":
				if m.state == statePrompt && !m.userInput {
					m.userInput = true
				}

			case "right":
				if m.state == statePrompt && m.userInput {
					m.userInput = false
				}
			}
		}

	case stateRun:
		switch msg := msg.(type) {
		case cmdFinishedErr:
			m.commands[m.currentIndex].Ran = true

			if msg.err != nil {
				m.errMessage = msg.err.Error()
				m.state = stateFailure
				return m, nil
			}

			m.state = stateSuccess
			return m, nil
			// TODO if we're at the end --> go to another screen

		}
		return m, m.spinner.Tick

	case stateSuccess:
		switch msg.(type) {
		case tea.KeyMsg:
			_ = m.nextCommand()
			// TODO: Proceed to the next part of the screen.

			return m, nil
		}

	case stateFailure:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	}

	return m, m.spinner.Tick
}

func (m *runModel) View() string {
	var b strings.Builder
	b.WriteString("State is currently: ")
	b.WriteString(string(m.state))
	b.WriteRune('\n')
	b.WriteRune('\n')

	command := m.commands[m.currentIndex]
	b.WriteString(command.View())
	b.WriteString("\n\n")
	switch m.state {
	case stateRun:
		b.WriteString("Running... ")
		b.WriteString(m.spinner.View())
		b.WriteRune('\n')

		b.WriteString(ViewError(m.cmdOut.String(), m.cmdErr.String(), m.errMessage))

	case statePrompt:
		b.WriteString("Run the command? ")
		b.WriteString(m.yesButton.View(m.userInput))
		b.WriteString("  ")
		b.WriteString(m.noButton.View(!m.userInput))
		b.WriteRune('\n')

	case stateSuccess:
		b.WriteString(style.Success("[SUCCESS]"))
		b.WriteString(" Command completed.\n")

		b.WriteString(ViewError(m.cmdOut.String(), m.cmdErr.String(), m.errMessage))

	case stateFailure:
		b.WriteString(style.Error("[ERROR]"))
		b.WriteString(" Command failed:\n")

		b.WriteString(ViewError(m.cmdOut.String(), m.cmdErr.String(), m.errMessage))
	}

	b.WriteRune('\n')
	switch m.state {
	case stateFailure:
		b.WriteString(helpStyle.Render("(press any key to quit)"))
	case stateSuccess:
		b.WriteString(helpStyle.Render("(press any key to continue)"))
	default:
		b.WriteString(helpStyle.Render("(ctrl+c or esc to quit)"))
	}

	return b.String()
}

func (m *runModel) nextCommand() bool {
	for ; m.currentIndex < len(m.commands); m.currentIndex++ {
		current := m.commands[m.currentIndex]
		if current.Ran {
			continue
		}

		prereq := current.SkipIfSucceed
		if prereq == nil {
			m.cmdErr.Reset()
			m.cmdOut.Reset()
			m.errMessage = ""
			m.state = statePrompt
			return true
		}

		if prereq.Ran && prereq.Succeeded {
			continue
		}

		if !prereq.Ran {
			prereq.Ran = true

			c := exec.Command("bash", "-c", fmtCommandForExec(prereq.Command)) //nolint:gosec
			if err := c.Run(); err != nil {
				m.cmdErr.Reset()
				m.cmdOut.Reset()
				m.errMessage = ""
				m.state = statePrompt
				prereq.Succeeded = false
				return true
			}

			prereq.Succeeded = true
		}
	}

	return false
}
