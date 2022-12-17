package goalinput

import (
	"math/rand"
	"strings"
	"time"

	"github.com/benhsm/goals/internal/ui/colorpicker"
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

// Model type

type Model struct {
	TitleInput    textinput.Model
	DescInput     textarea.Model
	focusIndex    int
	colorpicker   colorpicker.Model
	Done          bool
	Cancelled     bool
	Color         lipgloss.Color
	height        int
	width         int
	choosingColor bool
}

const (
	focusTitle = iota
	focusDesc
	focusColor
	focusDone
	focusCancel
)

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "goal title"
	ti.Focus()
	ta := textarea.New()
	ta.Placeholder = "goal description"

	rand.Seed(time.Now().UnixNano())
	cp := colorpicker.New()
	randomIndex := rand.Intn(len(cp.Colors))
	return Model{
		TitleInput:  ti,
		DescInput:   ta,
		colorpicker: cp,
		Color:       cp.Colors[randomIndex],
	}
}

// Init
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// View
func (m Model) View() string {
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

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		b.String())
}

// Update

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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

func (m Model) goalInputUpdate(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
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
					m.Cancelled = true
				} else if m.focusIndex == focusColor {
					m.choosingColor = true
					m.colorpicker.SetSize(m.height, m.width)
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

func (m *Model) SetSize(height, width int) {
	m.height = height
	m.width = width
}
