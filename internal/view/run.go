package view

import (
	"strings"

	"flightcrew.io/cli/internal/debug"
	"flightcrew.io/cli/internal/style"
	"flightcrew.io/cli/internal/view/button"
	"flightcrew.io/cli/internal/view/command"
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

	commands     []*command.Model
	currentIndex int

	userInput bool
	yesButton *button.Button
	noButton  *button.Button

	quitting bool
}

func NewRunModel(params map[string]string) *runModel {
	debug.Output("New run model time!")

	yesButton, _ := button.New("yes", 10)
	noButton, _ := button.New("no", 10)

	checkServiceAccount := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew service account already exists or needs to be created.",
		Command:     `gcloud iam service-accounts describe --project="${GOOGLE_PROJECT_ID}" "${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" > /dev/null 2>&1`,
		Message: map[command.State]string{
			command.PassState: "The service account already exists.",
			command.FailState: "No service account found. Next step is to create one.",
		},
	})
	checkIAMRole := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew IAM Role already exists or needs to be created.",
		Command:     `gcloud iam roles describe --project="${GOOGLE_PROJECT_ID}" "${IAM_ROLE}" >/dev/null 2>&1`,
		Message: map[command.State]string{
			command.PassState: "This Flightcrew IAM role already exists.",
			command.FailState: "No IAM role found. Next step is to create one.",
		},
	})
	checkVMExists := command.NewReadModel(command.Opts{
		Description: "Check if a Flightcrew VM already exists or needs to be created.",
		Command:     `gcloud compute instances list --format="csv(NAME,EXTERNAL_IP,STATUS)" --project=${GOOGLE_PROJECT_ID} --zones=${ZONE} | awk -F "," "/${VIRTUAL_MACHINE}/ {print f(2), f(3)} function f(n){return (\$n==\"\" ? \"null\" : \$n)}"`,
		Message: map[command.State]string{
			command.PassState: "This Flightcrew VM already exists. Nothing to install.",
			command.FailState: "No existing VM found. Next step is to create it.",
		},
	})

	m := &runModel{
		params: params,
		commands: []*command.Model{
			checkServiceAccount,
			command.NewWriteModel(command.Opts{
				SkipIfSucceed: checkServiceAccount,
				Description: `This command will create a service account, and follow-up commands will attach ${PERMISSIONS} permissions.

https://cloud.google.com/iam/docs/creating-managing-service-accounts`,
				Command: `gcloud iam service-accounts create "${SERVICE_ACCOUNT}" \
	--project="${GOOGLE_PROJECT_ID}" \
	--display-name="${SERVICE_ACCOUNT}" \
	--description="Runs Flightcrew's Control Tower VM."`,
			}),
			checkIAMRole,
			command.NewWriteModel(command.Opts{
				SkipIfSucceed: checkIAMRole,
				Description:   "This command creates an IAM role from `${IAM_FILE}` for the Flightcrew VM to access configs and monitoring data.\n\nhttps://cloud.google.com/iam/docs/understanding-custom-roles",
				Command: `gcloud iam roles create ${IAM_ROLE} \
	--project=${GOOGLE_PROJECT_ID} \
	--file=${IAM_FILE}`,
			}),
			command.NewWriteModel(command.Opts{
				SkipIfSucceed: checkIAMRole,
				Description: `This command attaches the IAM role to Flightcrew's service account, which will give the IAM permissions to a new VM.

https://cloud.google.com/iam/docs/granting-changing-revoking-access`,
				Command: `gcloud projects add-iam-policy-binding "${GOOGLE_PROJECT_ID}" \
	--member=serviceAccount:"${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--role="projects/${GOOGLE_PROJECT_ID}/roles/${IAM_ROLE}" \
	--condition=None`,
			}),
			checkVMExists,
			command.NewWriteModel(command.Opts{
				SkipIfSucceed: checkVMExists,
				Description:   "Create a VM instance attached to Flightcrew's service account, and run the Control Tower image.\n\nYou can open `${ENV_FILE}` to edit your desired environment variables before you run this command.",
				// TODO(chris): Switch between prod and dev
				Command: `gcloud compute instances create-with-container ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--container-command="/ko-app/tower" \
	--container-image=${IMAGE_PATH}:${TOWER_VERSION} \
	--container-arg="--debug=true" \
	--container-env-file=${ENV_FILE} \
	--container-env=FC_API_KEY=${API_TOKEN} \
	--container-env=CLOUD_PLATFORM=${PLATFORM} \
	--container-env=FC_PACKAGE_VERSION=${TOWER_VERSION} \${TRAFFIC_ROUTER}
	--container-env=METRIC_PROVIDERS=stackdriver \
	--container-env=FC_RPC_CONNECT_HOST=${RPC_HOST} \
	--container-env=FC_RPC_CONNECT_PORT=443 \
	--container-env=FC_TOWER_PORT=8080 \
	--machine-type=e2-micro \
	--scopes=cloud-platform \
	--service-account="${SERVICE_ACCOUNT}@${GOOGLE_PROJECT_ID}.iam.gserviceaccount.com" \
	--tags=http-server \
	--zone=${ZONE}`,
			}),
			command.NewWriteModel(command.Opts{
				Command: `gcloud compute instances add-metadata ${VIRTUAL_MACHINE} \
	--project=${GOOGLE_PROJECT_ID} \
	--zone=${ZONE}  \
	--metadata=google-logging-enabled=false`,
				Description: `Disable the VM's builtin logger because it has a memory leak and will cause the VM to crash after 1-2 weeks.

https://serverfault.com/questions/980569/disable-fluentd-on-on-container-optimized-os-gce`,
			}),
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
		m.commands[i].Replace(replacer)
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

	cmd := m.commands[m.currentIndex]
	switch cmd.State() {
	case command.PromptState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				if cmd.State() == command.PromptState && !m.userInput {
					// TODO: Confirm quitting
					return m, tea.Quit
				}

				if cmd.State() == command.PromptState && m.userInput {
					return m, tea.Exec(cmd.GetCommandToRun(), func(err error) tea.Msg {
						return cmdFinishedErr{err}
					})
				}

			case "tab":
				if cmd.State() == command.PromptState {
					m.userInput = !m.userInput
				}

			case "left":
				if cmd.State() == command.PromptState && !m.userInput {
					m.userInput = true
				}

			case "right":
				if cmd.State() == command.PromptState && m.userInput {
					m.userInput = false
				}
			}
		}

	case command.RunningState:
		switch msg := msg.(type) {
		case cmdFinishedErr:
			cmd := m.commands[m.currentIndex]
			cmd.Complete(msg.err == nil)
			return m, nil
			// TODO if we're at the end --> go to another screen

		}
		return m, nil

	case command.PassState:
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

	case command.FailState:
		switch msg.(type) {
		case tea.KeyMsg:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *runModel) View() string {
	var b strings.Builder
	cmd := m.commands[m.paginator.Page]
	b.WriteString("State is currently: ")
	b.WriteString(string(cmd.State()))
	b.WriteRune('\n')
	b.WriteRune('\n')

	b.WriteString(lipgloss.NewStyle().PaddingLeft(2).Render(m.paginator.View()))
	b.WriteRune('\n')

	b.WriteString(cmd.View())
	b.WriteString("\n\n")

	if cmd.State() == command.PromptState {
		b.WriteString(style.Action("[ACTION REQUIRED]"))
		b.WriteString(" Run the command? ")
		b.WriteString(m.yesButton.View(m.userInput))
		b.WriteString("  ")
		b.WriteString(m.noButton.View(!m.userInput))
		b.WriteRune('\n')
	}

	if cmd.State() == command.FailState && !cmd.IsRead() {
		b.WriteString(style.Help("(press any key to quit)"))
	} else if cmd.State() == command.PassState {
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
		if current.State() != command.NoneState {
			continue
		}

		if current.ShouldRun() {
			return true
		}
	}

	return false
}
