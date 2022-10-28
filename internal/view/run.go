package view

import (
	"os/exec"
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

	quitting bool
}

func NewRunModel(params map[string]string) *runModel {
	debug.Output("New run model time!")

	yesButton, err := NewButton("yes", 10)
	_ = err
	noButton, err := NewButton("no", 10)
	_ = err

	checkServiceAccount := &CommandState{
		Read: &ReadCommandState{
			SuccessMessage: "This service account already exists.",
			FailureMessage: "No service account found. Next step is to create one.",
		},
		Command:     `gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" > /dev/null 2>&1`,
		Description: "Check if a Flightcrew service account already exists or needs to be created.",
	}
	checkIAMRole := &CommandState{
		Command: `gcloud iam roles describe --project="${GOOGLE_PROJECT_ID}" "${IAM_ROLE}" >/dev/null 2>&1`,
		Read: &ReadCommandState{
			SuccessMessage: "This Flightcrew IAM role already exists.",
			FailureMessage: "No IAM role found. Next step is to create one.",
		},
		Description: "Check if a Flightcrew IAM Role already exists or needs to be created.",
	}
	checkVMExists := &CommandState{
		Command: `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/ {print f(2), f(3)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}"`,
		Read: &ReadCommandState{
			SuccessMessage: "This Flightcrew VM already exists. Nothing to install.",
			FailureMessage: "No existing VM found. Next step is to create it.",
		},
		Description: "Check if a Flightcrew VM already exists or needs to be created.",
	}

	m := &runModel{
		params: params,
		commands: []*CommandState{
			checkServiceAccount,
			{
				Command: `gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
	--project="${GOOGLE_PROJECT_ID}" \
	--display-name="${SERVICE_ACCOUNT}" \
	--description="Runs Flightcrew's Control Tower VM."`,
				Description: `This command will create a service account, and follow-up commands will attach ${PERMISSIONS} permissions.`,
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
				Description: "This command creates an IAM role from `${IAM_FILE}` for the Flightcrew VM to access configs and monitoring data.",
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkIAMRole,
					Link:          "https://cloud.google.com/iam/docs/understanding-custom-roles",
				},
			},
			{
				Command: `gcloud projects add-iam-policy-binding "${GOOGLE_PROJECT_ID}" \
	--member=serviceAccount:"${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--role="projects/${GOOGLE_PROJECT_ID}/roles/${IAM_ROLE}" \
	--condition=None`,
				Description: "This command attaches the IAM role to Flightcrew's service account, which will give the IAM permissions to a new VM.",
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkIAMRole,
					Link:          "https://cloud.google.com/iam/docs/granting-changing-revoking-access",
				},
			},
			checkVMExists,
			{
				// TODO(chris): Switch between prod and dev
				Command: `gcloud compute instances create-with-container ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--container-command="/ko-app/tower" \
	--container-image=${IMAGE_PATH}:${TOWER_VERSION} \
	--container-arg="--debug=true" \
	--container-env-file=${ENV_FILE} \
	--container-env=FC_API_KEY=${API_TOKEN} \
	--container-env=FC_PACKAGE_VERSION=${TOWER_VERSION} \
	--container-env=METRIC_PROVIDERS=stackdriver \
	--container-env=FC_RPC_CONNECT_HOST=${RPC_HOST} \
	--container-env=FC_RPC_CONNECT_PORT=443 \
	--container-env=FC_TOWER_PORT=8080 \
	--machine-type=e2-micro \
	--scopes=cloud-platform \
	--service-account="${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--tags=http-server \
	--zone=${ZONE}`,
				Description: "Create a VM instance attached to Flightcrew's service account, and run the Control Tower image.\n\nYou can open `${ENV_FILE}` to edit your desired environment variables before you run this command.",
				Mutate: &MutateCommandState{
					SkipIfSucceed: checkVMExists,
				},
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

	m.nextCommand()

	return m
}

func (m *runModel) Init() tea.Cmd {
	return nil
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
		return m, nil

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

			return m, nil
		}

	case FailState:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *runModel) View() string {
	var b strings.Builder
	command := m.commands[m.paginator.Page]
	b.WriteString("State is currently: ")
	b.WriteString(string(command.State))
	b.WriteRune('\n')
	b.WriteRune('\n')

	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(m.paginator.View()))
	b.WriteRune('\n')

	b.WriteString(command.View())
	b.WriteString("\n\n")

	if command.State == PromptState {
		b.WriteString(style.Action("[ACTION REQUIRED]"))
		b.WriteString(" Run the command? ")
		b.WriteString(m.yesButton.View(m.userInput))
		b.WriteString("  ")
		b.WriteString(m.noButton.View(!m.userInput))
		b.WriteRune('\n')
	}

	if command.State == FailState && command.Mutate != nil {
		b.WriteString(style.Help("(press any key to quit)"))
	} else if command.State == PassState {
		b.WriteString(style.Help("(press any key to continue)"))
	} else {
		b.WriteString(style.Help("ctrl+c/esc: quit • h/l: page • ←/→/enter: run command"))
	}

	b.WriteRune('\n')
	return b.String()
}

func (m *runModel) nextCommand() bool {
	for ; m.currentIndex < len(m.commands); m.currentIndex++ {
		m.paginator.Page = m.currentIndex
		current := m.commands[m.currentIndex]
		if current.State != NoneState {
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
			return true
		}

		if prereq.State == PassState {
			current.State = SkipState
			continue
		}

		if prereq.State == NoneState {
			c := exec.Command("bash", "-c", fmtCommandForExec(prereq.Command)) //nolint:gosec
			if err := c.Run(); err != nil {
				prereq.State = FailState
				current.State = PromptState
				return true
			}

			prereq.State = PassState
		}
	}

	return false
}
