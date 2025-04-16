package textinput

import (
	"fmt"
	"io"
	"os"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit"
	"github.com/muesli/termenv"
)

const DefaultTemplate = `
{{- .Prompt }}
> {{ .Input -}}
`
const DefaultMask = '●'

type TextInput struct {
	Prompt string

	Placeholder  string
	InitialValue string

	Validate     func(string) error
	AutoComplete func(string) []string

	Hidden                bool
	HideMask              rune
	CharLimit             int
	InputWidth            int
	Template              string
	ResultTemplate        string
	ExtendedTemplateFuncs template.FuncMap

	// Styles of the actual input field. These will be applied as inline styles.
	//
	// For an introduction to styling with Lip Gloss see:
	// https://github.com/charmbracelet/lipgloss
	InputTextStyle        lipgloss.Style
	InputBackgroundStyle  lipgloss.Style // Deprecated: This property is not used anymore.
	InputPlaceholderStyle lipgloss.Style
	InputCursorStyle      lipgloss.Style

	WrapMode     promptkit.WrapMode
	Output       io.Writer
	Input        io.Reader
	ColorProfile termenv.Profile
}

func New(prompt string) *TextInput {
	return &TextInput{
		Prompt:   prompt,
		Template: DefaultTemplate,
		// ResultTemplate:        DefaultResultTemplate,
		// KeyMap:                NewDefaultKeyMap(),
		InputPlaceholderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		// Validate:              ValidateNotEmpty,
		HideMask:              DefaultMask,
		ExtendedTemplateFuncs: template.FuncMap{},
		WrapMode:              promptkit.Truncate,
		Output:                os.Stdout,
		Input:                 os.Stdin,
	}
}

func (t *TextInput) RunPrompt() (string, error) {
	// err := validateKeyMap(t.KeyMap)
	// if err != nil {
	// 	return "", fmt.Errorf("insufficient key map: %w", err)
	// }

	m := NewModel(t)
	p := tea.NewProgram(m, tea.WithOutput(t.Output), tea.WithInput(t.Input))

	_, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("running prompt: %w", err)
	}

	return m.Value()
}
