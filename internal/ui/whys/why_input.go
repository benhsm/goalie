// Package goalinput provides an application specific Bubble Tea UI component
// for adding or editing goals
package whys

import (
	"math/rand"
	"strings"
	"time"

	"github.com/benhsm/goalie/internal/ui/common"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleInputStyle = lipgloss.NewStyle().
			Padding(0, 0, 1, 2)
	descInputStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder(), true, true, true, true)
	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#888B7E")).
			BorderTop(true).
			BorderLeft(true).
			BorderBottom(true).Padding(0, 1).Margin(1, 2, 0, 2)

	focusedButton = buttonStyle.Copy().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#8F26D9")).
			Underline(true)
)

// Model represents a goal input UI
type goalInputModel struct {
	common.Common
	TitleInput    textinput.Model
	DescInput     textarea.Model
	focusIndex    int
	colorpicker   colorPickerModel
	Done          bool
	Cancelled     bool
	Color         lipgloss.Color
	choosingColor bool
}

const (
	focusTitle = iota
	focusDesc
	focusColor
	focusDone
	focusCancel
)

// New returns a New goalinput model
func newGoalInput() goalInputModel {
	ti := textinput.New()
	ti.Placeholder = "goal title"
	ti.CharLimit = 50
	ti.Focus()
	ta := textarea.New()
	ta.Placeholder = "goal description"

	rand.Seed(time.Now().UnixNano())
	cp := newColorPicker()
	randomIndex := rand.Intn(len(cp.Colors))
	return goalInputModel{
		TitleInput:  ti,
		DescInput:   ta,
		colorpicker: cp,
		Color:       cp.Colors[randomIndex],
	}
}

// Init returns a Bubble Tea command that initializes the goal model
func (m goalInputModel) Init() tea.Cmd {
	return textinput.Blink
}

// View returns a string that represents the goal input UI
func (m goalInputModel) View() string {
	var b strings.Builder

	if m.choosingColor {
		return m.colorpicker.View()
	}

	titleInput := titleInputStyle.Render(m.TitleInput.View())

	descInput := descInputStyle.Render(m.DescInput.View())

	inputFields := lipgloss.JoinVertical(lipgloss.Left, titleInput, descInput)

	var colorButton, colorDisplay string
	colorDisplay = lipgloss.NewStyle().Background(m.Color).Foreground(lipgloss.Color("#FFFFFF")).Render(string(m.Color))
	if m.focusIndex == focusColor {
		colorButton = focusedButton.Render("change color")
	} else {
		colorButton = buttonStyle.Render("change color")
	}

	colorField := lipgloss.JoinHorizontal(lipgloss.Center, colorButton, colorDisplay)

	var doneButton, cancelButton string
	if m.focusIndex == focusDone {
		doneButton = focusedButton.Render("done")
	} else {
		doneButton = buttonStyle.Render("done")
	}

	if m.focusIndex == focusCancel {
		cancelButton = focusedButton.Render("cancel")
	} else {
		cancelButton = buttonStyle.Render("cancel")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, doneButton, cancelButton)

	b.WriteString(lipgloss.JoinVertical(lipgloss.Center, inputFields, colorField, buttons))

	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center,
		b.String())
}

// Update is the Bubble Tea update loop for the goalInput component
func (m goalInputModel) Update(msg tea.Msg) (goalInputModel, tea.Cmd) {
	var cmd tea.Cmd
	if m.choosingColor {
		m.colorpicker, cmd = m.colorpicker.Update(msg)
		if m.colorpicker.Choice != "" {
			m.choosingColor = false
			m.Color = m.colorpicker.Choice
			m.colorpicker.Choice = ""
		}
	} else {
		m, cmd = m.goalInputUpdate(msg)
	}
	return m, cmd
}

func (m goalInputModel) goalInputUpdate(msg tea.Msg) (goalInputModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
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
					m.Done = true
				} else if m.focusIndex == focusCancel {
					m.Done = true
					m.Cancelled = true
				} else if m.focusIndex == focusColor {
					m.choosingColor = true
					m.colorpicker.SetSize(m.Height, m.Width)
				} else {
					m.focusIndex++
				}
			}

			if m.focusIndex > focusCancel {
				m.focusIndex = focusTitle
			} else if m.focusIndex < focusTitle {
				m.focusIndex = focusCancel
			}

			if m.focusIndex == focusTitle {
				cmds = append(cmds, m.TitleInput.Focus())
			} else {
				m.TitleInput.Blur()
			}

			if m.focusIndex == focusDesc {
				cmds = append(cmds, m.DescInput.Focus())
			} else {
				m.DescInput.Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}

	m.TitleInput, cmd = m.TitleInput.Update(msg)
	cmds = append(cmds, cmd)

	m.DescInput, cmd = m.DescInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
