package auth

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/squarecloudofc/cli/internal/cli"
	"github.com/squarecloudofc/cli/internal/deviceauth"
	"github.com/squarecloudofc/cli/internal/ui"
)

var (
	boxStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("69")).Padding(1, 4)
	codeStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
)

type (
	warnMsg string
	doneMsg struct {
		result *deviceauth.Result
		err    error
	}
	clockMsg time.Time
)

// deviceModel shows the user code and a countdown while Wait polls in the
// background; doneMsg ends it, Ctrl+C cancels it.
type deviceModel struct {
	squareCli cli.SquareCLI
	grant     *deviceauth.Grant
	spinner   spinner.Model
	now       time.Time
	warnings  []string

	finished bool
	canceled bool
	result   *deviceauth.Result
	err      error
}

func newDeviceModel(squareCli cli.SquareCLI, grant *deviceauth.Grant) *deviceModel {
	return &deviceModel{squareCli: squareCli, grant: grant, spinner: ui.Spinner, now: time.Now()}
}

func tickClock() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return clockMsg(t) })
}

func (m *deviceModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tickClock())
}

func (m *deviceModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "esc" {
			m.canceled, m.finished = true, true
			return m, tea.Quit
		}
	case clockMsg:
		m.now = time.Time(msg)
		return m, tickClock()
	case warnMsg:
		m.warnings = append(m.warnings, string(msg))
	case doneMsg:
		m.result, m.err, m.finished = msg.result, msg.err, true
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *deviceModel) View() string {
	box := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
		m.squareCli.I18n().T("commands.auth.login.device.instructions"),
		"",
		codeStyle.Render(spacedCode(m.grant.UserCode)),
		"",
		ui.TextLink.Render(m.grant.VerificationURI),
	))

	var b strings.Builder
	b.WriteString(box + "\n")
	for _, w := range m.warnings {
		b.WriteString(ui.TextYellow.Render("! "+w) + "\n")
	}
	if !m.finished {
		remaining := max(m.grant.ExpiresAt().Sub(m.now), 0).Round(time.Second)
		fmt.Fprintf(&b, "%s %s\n", m.spinner.View(), m.squareCli.I18n().T("commands.auth.login.device.waiting", map[string]any{
			"Remaining": fmt.Sprintf("%d:%02d", int(remaining.Minutes()), int(remaining.Seconds())%60),
		}))
	}
	return b.String()
}

// spacedCode renders "ABCD2345" as "A B C D - 2 3 4 5" so it reads (and
// types) easily.
func spacedCode(code string) string {
	if len(code) == 8 {
		code = code[:4] + "-" + code[4:]
	}
	return strings.Join(strings.Split(code, ""), " ")
}
