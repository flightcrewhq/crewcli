package view

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
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

type runModel struct {
	params    map[string]string
	paginator paginator.Model

	commands     []*CommandState
	currentIndex int

	userInput bool
	yesButton *Button
	noButton  *Button

	spinner spinner.Model

	quitting bool

	logStatements []string
}

func NewRunModel(params map[string]string) *runModel {
	debug.Output("New run model time!")
	const defaultWidth = 20
	spin := spinner.New()
	spin.Style = spinnerStyle
	spin.Spinner = spinner.Line
	spin.Spinner.FPS = time.Second / 5

	yesButton, err := NewButton("yes", 10)
	_ = err
	noButton, err := NewButton("no", 10)
	_ = err

	checkServiceAccount := &CommandState{
		Read: &ReadCommandState{
			SuccessMessage: "Found a service account!",
			FailureMessage: "No service account found. Next step is to create one.",
		},
		Command: `gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" > /dev/null 2>&1`,
		State:   NoneState,
	}
	checkIAMRole := &CommandState{
		Command: `gcloud iam roles describe --project="${GOOGLE_PROJECT_ID}" "${IAM_ROLE}" >/dev/null 2>&1`,
		Read: &ReadCommandState{
			SuccessMessage: "Found the Flightcrew role!",
			FailureMessage: "No IAM role found. Next step is to create one.",
		},
		State: NoneState,
	}
	checkVMExists := &CommandState{
		Command: `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/ {print f(2), f(3)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}"`,
		Read: &ReadCommandState{
			SuccessMessage: "Flightcrew VM already exists! Aborting installation.",
			FailureMessage: "No existing VM found. Continuing to installation.",
		},
		State: NoneState,
	}

	m := &runModel{
		params:  params,
		spinner: spin,
		commands: []*CommandState{
			checkServiceAccount,
			{
				Command: `gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
	--project="${GOOGLE_PROJECT_ID}" \
	--display-name="${SERVICE_ACCOUNT}" \
	--description="Runs Flightcrew's Control Tower VM."`,
				State:       NoneState,
				Description: `This command will create a service account, and follow-up commands will attach read and/or write permissions.`,
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkServiceAccount,
					Link:          "https://cloud.google.com/iam/docs/creating-managing-service-accounts",
				},
			},
			checkIAMRole,
			{
				Command: `gcloud iam roles create ${IAM_ROLE} \
	--project=${GOOGLE_PROJECT_ID} \
	--file=${IAM_FILE}
`,
				Description: `This command creates an IAM role from ${IAM_FILE} for the Flightcrew VM to access configs and monitoring data.`,
				State:       NoneState,
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkIAMRole,
					Link:          "https://cloud.google.com/iam/docs/understanding-custom-roles",
				},
			},
			{
				Command: `gcloud projects add-iam-policy-binding "${GOOGLE_PROJECT_ID}" \\
	--member=serviceAccount:"${SERVICE_ACCT_EMAIL}" \
	--role="projects/${GOOGLE_PROJECT_ID}/roles/${IAM_ROLE}" \
	--condition=None`,
				Description: "This command attaches the IAM role to Flightcrew's service account, which will give the IAM permissions to a new VM.",
				State:       NoneState,
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkIAMRole,
					Link:          "https://cloud.google.com/iam/docs/granting-changing-revoking-access",
				},
			},
			checkVMExists,
			{
				Command: `gcloud compute instances create-with-container ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--container-command="/ko-app/tower" \
	--container-image=${FULL_IMAGE_PATH} \
	--container-arg="--debug=true" \
	--container-env-file=${ENV_FILE} \
	--container-env=FC_API_KEY=${FLIGHTCREW_TOKEN} \
	--container-env=FC_PACKAGE_VERSION=${TOWER_IMAGE_VERSION} \
	--machine-type=e2-micro \
	--scopes=cloud-platform \
	--service-account="${SERVICE_ACCT_EMAIL}" \
	--tags=http-server \
	--zone=${ZONE}`,
				Description: "Create a VM instance attached to Flightcrew's service account, and run the Control Tower image.",
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkVMExists,
				},
				State: NoneState,
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
		command := m.commands[i]
		command.State = NoneState
		command.Command = replacer.Replace(command.Command)
		command.Description = replacer.Replace(command.Description)
	}

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 1
	p.UseHLKeys = true
	p.UseJKKeys = false
	p.UseLeftRightKeys = false
	p.UsePgUpPgDownKeys = false
	p.UseUpDownKeys = false
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")
	p.SetTotalPages(len(m.commands))
	m.paginator = p

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

		case "h":
			var cmd tea.Cmd
			m.paginator, cmd = m.paginator.Update(msg)
			return m, cmd

		case "l":
			// Don't allow user to advance to future commands that haven't been run or prompted yet.
			if m.paginator.Page < m.currentIndex {
				var cmd tea.Cmd
				m.paginator, cmd = m.paginator.Update(msg)
				return m, cmd
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	command := m.commands[m.currentIndex]
	switch command.State {
	case PromptState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if command.State == PromptState && !m.userInput {
					// TODO: Confirm quitting
					return m, tea.Quit
				}

				if command.State == PromptState && m.userInput {
					command.State = RunningState
					return m, tea.Exec(NewExecCommand(command), func(err error) tea.Msg {
						return cmdFinishedErr{err}
					})
				}

			case "tab":
				if command.State == PromptState {
					m.userInput = !m.userInput
				}

			case "left":
				if command.State == PromptState && !m.userInput {
					m.userInput = true
				}

			case "right":
				if command.State == PromptState && m.userInput {
					m.userInput = false
				}
			}
		}

	case RunningState:
		switch msg := msg.(type) {
		case cmdFinishedErr:
			command := m.commands[m.currentIndex]

			if msg.err != nil {
				command.State = FailState
				return m, nil
			}

			command.State = PassState
			return m, nil
			// TODO if we're at the end --> go to another screen

		}
		return m, m.spinner.Tick

	case PassState:
		switch msg.(type) {
		case tea.KeyMsg:
			if m.quitting {
				return m, tea.Quit
			}

			if !m.nextCommand() {
				m.quitting = true
				return m, nil
			}

			return m, m.spinner.Tick
		}

	case FailState:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	}

	return m, m.spinner.Tick
}

func (m *runModel) View() string {
	var b strings.Builder
	command := m.commands[m.paginator.Page]
	b.WriteString("State is currently: ")
	b.WriteString(string(command.State))
	b.WriteRune('\n')
	b.WriteRune('\n')

	b.WriteString(m.paginator.View())
	b.WriteRune('\n')

	b.WriteString(command.View(m.spinner))
	b.WriteString("\n\n")

	if command.State == PromptState {
		b.WriteString("Run the command? ")
		b.WriteString(m.yesButton.View(m.userInput))
		b.WriteString("  ")
		b.WriteString(m.noButton.View(!m.userInput))
		b.WriteRune('\n')
	}

	switch command.State {
	case FailState:
		b.WriteString(style.Help("(press any key to quit)"))
	case PassState:
		b.WriteString(style.Help("(press any key to continue)"))
	default:
		b.WriteString(style.Help("(ctrl+c or esc to quit | h/l page)"))
	}

	b.WriteRune('\n')
	for _, str := range m.logStatements {
		b.WriteString(str)
		b.WriteRune('\n')
	}

	return b.String()
}

func (m *runModel) nextCommand() bool {
	for ; m.currentIndex < len(m.commands); m.currentIndex++ {
		m.paginator.Page = m.currentIndex
		current := m.commands[m.currentIndex]
		if current.State != NoneState {
			m.logStatements = append(m.logStatements, fmt.Sprintf("command %d is not in state none: %s", m.currentIndex, current.State))
			continue
		}

		if current.Read != nil {
			c := exec.Command("bash", "-c", fmtCommandForExec(current.Command)) //nolint:gosec
			if err := c.Run(); err != nil {
				current.State = FailState
				continue
			}

			current.State = PassState
			continue
		}

		if current.Mutate == nil {
			current.State = PromptState
			return true
		}

		prereq := current.Mutate.SkipIfSucceed
		if prereq == nil || prereq.State == FailState {
			current.State = PromptState
			m.logStatements = append(m.logStatements, fmt.Sprintf("Returning true for command %d: failed or no prereq", m.currentIndex))
			return true
		}

		if prereq.State == PassState {
			current.State = SkipState
			m.logStatements = append(m.logStatements, fmt.Sprintf("command %d's prereq is in state passed", m.currentIndex))
			continue
		}

		if prereq.State == NoneState {
			c := exec.Command("bash", "-c", fmtCommandForExec(prereq.Command)) //nolint:gosec
			if err := c.Run(); err != nil {
				prereq.State = FailState
				current.State = PromptState
				m.logStatements = append(m.logStatements, fmt.Sprintf("Returning true for command %d: just failed prereq", m.currentIndex))
				return true
			}

			m.logStatements = append(m.logStatements, fmt.Sprintf("Readonly command %d just passed", m.currentIndex))
			prereq.State = PassState
		}
	}

	return false
}
