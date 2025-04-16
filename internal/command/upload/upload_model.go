package upload

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/ui"
	"github.com/squarecloudofc/cli/pkg/squareconfig"
	"github.com/squarecloudofc/cli/pkg/squareignore"
	"github.com/squarecloudofc/cli/pkg/zipper"
	"github.com/squarecloudofc/squarego/squarecloud"
)

type Step int

const (
	_ Step = iota

	LoadFile
	CompressStep
	UploadStep
	CompletedStep
)

var stepMessages = map[Step]string{
	LoadFile:      "File provided, skipping compression.",
	CompressStep:  "Compressing the current directory.",
	UploadStep:    "Uploading the zip file to Square Cloud.",
	CompletedStep: "Upload completed successfully.",
}

type StepCompleted struct {
	Message string
	err     error
}

var (
	hiddenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250")).Render
	textStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("254")).Render

	linkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Underline(true).Render
)

var defaultSpinner = spinner.New(spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("69"))), spinner.WithSpinner(spinner.Jump))

type model struct {
	cli     *cli.SquareCli
	config  *squareconfig.SquareConfig
	zipFile *os.File
	workDir string

	err                 error
	completed           []StepCompleted
	currentStep         Step
	uploadedApplication *squarecloud.ApplicationUploaded

	spinner spinner.Model
	done    bool
}

func NewModel(squarecli *cli.SquareCli, config *squareconfig.SquareConfig) (*model, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	m := &model{
		cli:     squarecli,
		config:  config,
		workDir: workDir,
	}

	return m, err
}

func (m *model) Init() tea.Cmd {
	m.currentStep = CompressStep

	m.spinner = spinner.New(
		spinner.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("69")),
		),
		spinner.WithSpinner(spinner.Jump),
	)

	return tea.Batch(m.spinner.Tick, m.nextStep())
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case StepCompleted:
		if m.currentStep != CompletedStep {
			m.completed = append(m.completed, msg)
		}

		if msg.err != nil {
			m.err = msg.err
			m.done = true
			return m, tea.Quit
		}

		switch m.currentStep {
		case CompressStep:
			m.currentStep = UploadStep
		case UploadStep:
			m.currentStep = CompletedStep
		case CompletedStep:
			m.done = true
		}

		if m.done {
			return m, tea.Quit
		}

		return m, m.nextStep()
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
	default:
		return m, nil
	}
}

func (m *model) View() string {
	var s string

	for _, step := range m.completed {
		emoji := ui.CheckMark
		if step.err != nil {
			emoji = ui.XMark
		}

		s += fmt.Sprintf("%s %s\n", emoji, hiddenStyle(step.Message))
	}

	if m.currentStep != 0 && !m.done {
		s += fmt.Sprintf("%s %s\n", m.spinner.View(), textStyle(stepMessages[m.currentStep]))
	}

	if m.err != nil {
		s += "\n  Error:"
		s += fmt.Sprintf(" %s", m.err.Error())
	}

	if m.done && m.err == nil {
		link := linkStyle(fmt.Sprintf("https://squarecloud.app/dashboard/app/%s", m.uploadedApplication.ID))
		access := fmt.Sprintf("You can access via %s", link)
		s += fmt.Sprintf("\n%s Application uploaded to Square Cloud!\n  %s", ui.CheckMark, access)
	}

	return lipgloss.NewStyle().Padding(1, 1).Render(s)
}

func (m *model) nextStep() tea.Cmd {
	return func() tea.Msg {
		switch m.currentStep {
		case CompressStep:
			ignoreFiles, _ := squareignore.Load()
			file, err := zipper.ZipFolder(m.workDir, ignoreFiles)
			m.zipFile = file

			return StepCompleted{
				Message: stepMessages[m.currentStep],
				err:     err,
			}

		case UploadStep:
			uploaded, err := m.cli.Rest().PostApplications(m.zipFile)
			if err == nil && uploaded.ID != "" {
				m.config.ID = uploaded.ID
				err = m.config.Save()
			}

			m.uploadedApplication = uploaded
			return StepCompleted{
				Message: stepMessages[m.currentStep],
				err:     err,
			}

		default:
			return StepCompleted{}
		}
	}
}
