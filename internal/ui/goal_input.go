package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			BorderTop(true).
			BorderLeft(true).
			BorderBottom(true).Padding(0, 1)

	focusedButton = buttonStyle.Copy().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#8F26D9")).
			Underline(true)
)

// Model type

type goalInputModel struct {
	goals      []string
	titleInput textinput.Model
	descInput  textarea.Model
	focusIndex int
}

const (
	focusTitle = iota
	focusDesc
	focusDone
	focusCancel
)

func NewGoalInput() goalInputModel {
	ti := textinput.New()
	ti.Placeholder = "goal title"
	ti.Focus()
	ta := textarea.New()
	ta.Placeholder = "goal description"
	return goalInputModel{
		titleInput: ti,
		descInput:  ta,
	}
}

// Init
func (m goalInputModel) Init() tea.Cmd {
	return textinput.Blink
}

// View
func (m goalInputModel) View() string {
	var b strings.Builder

	for i, s := range m.goals {
		b.WriteString(fmt.Sprintf("%d) %s\n", i, s))
	}

	b.WriteString("   ")
	b.WriteString(m.titleInput.View())
	b.WriteString("\n\n")
	b.WriteString(m.descInput.View())
	b.WriteString("\n\n")

	if m.focusIndex == focusDone {
		b.WriteString(focusedButton.Render("done"))
	} else {
		b.WriteString(buttonStyle.Render("done"))
	}

	b.WriteString("  ")
	if m.focusIndex == focusCancel {
		b.WriteString(focusedButton.Render("cancel"))
	} else {
		b.WriteString(buttonStyle.Render("cancel"))
	}
	b.WriteString("\n")

	return b.String()
}

// Update
func (m goalInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyEnter:
			if msg.Type == tea.KeyTab {
				m.focusIndex++
			} else if msg.Type == tea.KeyShiftTab {
				m.focusIndex--
			}

			if msg.Type == tea.KeyEnter {
				if m.focusIndex == focusDone {
					m.goals = append(m.goals, m.titleInput.Value())
					m.titleInput.SetValue("")
				} else if m.focusIndex == focusCancel {
					return m, tea.Quit
				}
			}

			if m.focusIndex > focusCancel {
				m.focusIndex = focusTitle
			} else if m.focusIndex < focusTitle {
				m.focusIndex = focusCancel
			}

			if m.focusIndex == focusTitle {
				cmds = append(cmds, m.titleInput.Focus())
			} else {
				m.titleInput.Blur()
			}

			if m.focusIndex == focusDesc {
				cmds = append(cmds, m.descInput.Focus())
			} else {
				m.descInput.Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	m.titleInput, cmd = m.titleInput.Update(msg)
	cmds = append(cmds, cmd)

	m.descInput, cmd = m.descInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
