package application_selector

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/sdk-api-go/squarecloud"
)

type model struct {
	cli  cli.SquareCLI
	apps []squarecloud.UserApplication

	err error

	selectedIndex       int
	SelectedApplication squarecloud.UserApplication

	done bool
}

func RunSelector(squareCli cli.SquareCLI) (squarecloud.UserApplication, error) {
	applications, err := squareCli.Rest().GetApplications()
	if err != nil {
		return squarecloud.UserApplication{}, nil
	}

	m, err := NewModel(squareCli, applications)
	if err != nil {
		return squarecloud.UserApplication{}, err
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return squarecloud.UserApplication{}, err
	}

	return m.SelectedApplication, nil
}

func NewModel(squarecli cli.SquareCLI, applications []squarecloud.UserApplication) (*model, error) {
	m := &model{
		cli:  squarecli,
		apps: applications,
	}

	return m, nil
}

func (m *model) Init() tea.Cmd {
	return nil
	// m.spinner = ui.Spinner
	//
	// return tea.Batch(m.spinner.Tick, m.runTask("LOAD_FILE"))
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// case spinner.TickMsg:
	// 	var cmd tea.Cmd
	// 	// m.spinner, cmd = m.spinner.Update(msg)
	// 	return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case tea.KeyUp.String():
			if m.selectedIndex != 0 {
				m.selectedIndex -= 1
			}

			return m, nil
		case tea.KeyDown.String():
			if m.selectedIndex+1 != len(m.apps) {
				m.selectedIndex += 1
			}

			return m, nil
		case "enter":
			m.SelectedApplication = m.apps[m.selectedIndex]
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m *model) View() string {
	if m.done {
		return ""
	}

	var s string
	s += fmt.Sprintf("\n %s\n", lipgloss.NewStyle().Bold(true).Render(m.cli.I18n().T("ui.select.application")))

	for i, app := range m.apps {
		applicationName := lipgloss.NewStyle()
		if m.selectedIndex == i {
			applicationName = applicationName.Foreground(lipgloss.ANSIColor(33)).SetString("⏵ ")
		} else {
			applicationName = applicationName.Foreground(ui.TextPrimary.GetForeground()).SetString("  ")
		}

		applicationData := lipgloss.NewStyle().SetString(fmt.Sprintf("(%s - %s)", app.ID, app.Cluster)).Foreground(ui.TextTertiary.GetForeground())

		var appInfo string
		appInfo += fmt.Sprintf("%s %s", applicationName.Render(app.Name), applicationData)

		s += lipgloss.NewStyle().PaddingLeft(2).Render(appInfo)
		s += "\n"
	}

	return s
}
