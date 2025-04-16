package textinput

import (
	"bytes"
	"text/template"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	*TextInput

	Err error

	input textinput.Model

	tmpl       *template.Template
	resultTmpl *template.Template

	quitting bool
}

func NewModel(text *TextInput) *model {
	return &model{TextInput: text}
}

func (m *model) Init() tea.Cmd {
	m.tmpl, m.Err = m.initTemplate()
	if m.Err != nil {
		return tea.Quit
	}

	m.input = m.initInput()

	return textinput.Blink
}

func (m *model) initTemplate() (*template.Template, error) {
	tmpl := template.New("view")
	// tmpl.Funcs(termenv.TemplateFuncs(m.ColorProfile))
	// tmpl.Funcs(promptkit.UtilFuncMap())
	// tmpl.Funcs(m.ExtendedTemplateFuncs)
	// tmpl.Funcs(template.FuncMap{
	// 	"Mask": m.mask,
	// 	"AutoCompleteSuggestions": func() []string {
	// 		return m.AutoComplete(m.input.Value())
	// 	},
	// })

	return tmpl.Parse(m.Template)
}

func (m *model) initInput() textinput.Model {
	input := textinput.New()
	input.Prompt = ""

	if m.Hidden {
		input.EchoMode = textinput.EchoPassword
		input.EchoCharacter = m.HideMask
	}

	input.Focus()
	return input
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}

		// We handle errors just like any other message
		// case errMsg:
		// 	m.err = msg
		// 	return m, nil
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	if m.quitting {
		return ""
	}

	buf := &bytes.Buffer{}

	m.tmpl.Execute(buf, map[string]any{
		"Prompt": m.Prompt,
		"Input":  m.input.View(),
	})

	return buf.String()
}

func (m *model) Value() (string, error) {
	return m.input.Value(), nil
}
