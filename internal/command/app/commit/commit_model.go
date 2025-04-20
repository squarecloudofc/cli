package commit

import (
	"crypto/rand"
	"math/big"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
)

type Task struct {
	id      string
	message string

	err       error
	completed bool
}

type model struct {
	cli     cli.SquareCLI
	options *CommitOptions

	zipFile *os.File
	workDir string

	tasks []*Task

	err error

	spinner spinner.Model
	done    bool
}

func NewModel(squarecli cli.SquareCLI, options *CommitOptions) (*model, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	m := &model{
		cli:     squarecli,
		options: options,
		workDir: workDir,
	}

	return m, err
}

func (m *model) Init() tea.Cmd {
	m.spinner = ui.Spinner

	return tea.Batch(m.spinner.Tick, m.runTask("ZIP_WORKDIR"))
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *Task:
		for i, task := range m.tasks {
			if task.id == msg.id {
				m.tasks[i] = msg
			}
		}

		if m.done || msg.err != nil {
			return m, tea.Quit
		}

		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		default:
			return m, nil
		}
	}

	return m, nil
}

func (m *model) View() string {
	var s string

	for _, t := range m.tasks {
		if t.completed {
			s += ui.CheckMark.Render() + " " + ui.TextSecondary.Render(t.message) + "\n"
		}

		if !t.completed && t.err == nil {
			s += m.spinner.View() + " " + ui.TextPrimary.Render(t.message) + "\n" + "\n"
		}

		if t.err != nil {
			test := lipgloss.NewStyle().
				PaddingLeft(1).
				Render(t.message + "\n" + "Error: " + t.err.Error())

			s += lipgloss.JoinHorizontal(lipgloss.Top, ui.XMark.Render(), test)

		}
	}

	if m.done && m.err == nil {
		test := lipgloss.NewStyle().
			PaddingLeft(1).
			Render(m.cli.I18n().T("commands.app.commit.success"))

		s += lipgloss.NewStyle().PaddingTop(1).Render(
			lipgloss.JoinHorizontal(0.2, ui.CheckMark.Render(), test),
		)
	}

	return lipgloss.NewStyle().Padding(1, 1).Render(s)
}

func (m *model) runTask(name string) tea.Cmd {
	return func() tea.Msg {
		taskId := generateRandomString(6)
		switch name {
		case "ZIP_WORKDIR":
			task := &Task{
				id:      taskId,
				message: m.cli.I18n().T("commands.app.commit.states.compressing"),
			}

			m.tasks = append(m.tasks, task)
			m.Update(nil)

			if m.options.FileName != "" {
				m.options.File, _ = handleCommitFile(m.cli, m.options)
			}

			if m.options.File == nil {
				m.options.File, task.err = handleCommitWorkingDirectory()
				if task.err != nil {
					return task
				}
			}

			task.completed = true

			return m.runTask("UPLOAD")()
		case "UPLOAD":
			task := &Task{
				id:      taskId,
				message: m.cli.I18n().T("commands.app.commit.states.uploading") + m.options.ApplicationID,
			}

			m.tasks = append(m.tasks, task)
			m.Update(nil)

			if task.err = m.cli.Rest().PostApplicationCommit(m.options.ApplicationID, m.options.File); task.err != nil {
				return task
			}

			task.completed = true
			m.done = true

			return task

		}

		return nil
	}
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		randomInt, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[randomInt.Int64()]
	}
	return string(result)
}
