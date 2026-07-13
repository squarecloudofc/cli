// Package textinput is a thin wrapper around bubbles/textinput used for the
// hidden token prompt.
package textinput

import (
	"errors"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// ErrCanceled is returned when the user aborts the prompt (esc/ctrl+c).
var ErrCanceled = errors.New("input canceled")

type TextInput struct {
	Prompt      string
	Placeholder string
	// Hidden masks the typed value (password style).
	Hidden bool
}

func New(prompt string) *TextInput {
	return &TextInput{Prompt: prompt}
}

// RunPrompt renders the prompt and blocks until the user submits or cancels.
func (t *TextInput) RunPrompt() (string, error) {
	input := textinput.New()
	input.Placeholder = t.Placeholder
	if t.Hidden {
		input.EchoMode = textinput.EchoPassword
		input.EchoCharacter = '*'
	}
	input.Focus()

	m := &model{prompt: t.Prompt, input: input}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return "", err
	}

	if m.canceled {
		return "", ErrCanceled
	}

	return m.input.Value(), nil
}

type model struct {
	prompt   string
	input    textinput.Model
	done     bool
	canceled bool
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			m.canceled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.done {
		return ""
	}

	return m.prompt + "\n" + m.input.View() + "\n"
}
