// Package selector provides a minimal interactive arrow-key picker used to
// choose an application or database when no ID argument is given.
package selector

import (
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/squarecloudofc/cli/internal/ui"
)

// ErrCanceled is returned when the user aborts the selection (esc/ctrl+c).
var ErrCanceled = errors.New("selection canceled")

type Item struct {
	ID    string
	Label string
	// Desc is rendered dimmed next to the label, e.g. "(id - cluster)".
	Desc string
}

type model struct {
	title string
	items []Item

	selectedIndex int
	selected      *Item
	done          bool
}

// Run renders the picker and blocks until the user selects an item or
// cancels.
func Run(title string, items []Item) (Item, error) {
	m := &model{title: title, items: items}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		return Item{}, err
	}

	if m.selected == nil {
		return Item{}, ErrCanceled
	}

	return *m.selected, nil
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch msg.String() {
		case tea.KeyUp.String():
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case tea.KeyDown.String():
			if m.selectedIndex+1 < len(m.items) {
				m.selectedIndex++
			}
		case "enter":
			m.selected = &m.items[m.selectedIndex]
			m.done = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m *model) View() string {
	if m.done {
		return ""
	}

	s := fmt.Sprintf("\n %s\n", lipgloss.NewStyle().Bold(true).Render(m.title))

	for i, item := range m.items {
		label := lipgloss.NewStyle()
		if m.selectedIndex == i {
			label = label.Foreground(lipgloss.ANSIColor(33)).SetString("⏵ ")
		} else {
			label = label.Foreground(ui.TextPrimary.GetForeground()).SetString("  ")
		}

		desc := lipgloss.NewStyle().SetString(item.Desc).Foreground(ui.TextTertiary.GetForeground())

		s += lipgloss.NewStyle().PaddingLeft(2).Render(fmt.Sprintf("%s %s", label.Render(item.Label), desc))
		s += "\n"
	}

	return s
}
