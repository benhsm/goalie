package goalinput

import (
	"strings"

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
			BorderBottom(true).Padding(0, 1).Margin(1, 2)

	focusedButton = buttonStyle.Copy().
			Foreground(lipgloss.Color("#FFF7DB")).
			Background(lipgloss.Color("#8F26D9")).
			Underline(true)
)

// Model type

type Model struct {
	titleInput  textinput.Model
	descInput   textarea.Model
	focusIndex  int
	colorpicker colorpicker.Model
	Done        bool
	Cancelled   bool
	height      int
	width       int
}

const (
	focusTitle = iota
	focusDesc
	focusDone
	focusCancel
)

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "goal title"
	ti.Focus()
	ta := textarea.New()
	ta.Placeholder = "goal description"
	return Model{
		titleInput:  ti,
		descInput:   ta,
		colorpicker: colorpicker.New(),
	}
}

// Init
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// View
func (m Model) View() string {
	var b strings.Builder

	titleInput := titleInputStyle.Render(m.titleInput.View())

	descInput := descInputStyle.Render(m.descInput.View())

	inputFields := lipgloss.JoinVertical(lipgloss.Left, titleInput, descInput)

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

	b.WriteString(lipgloss.JoinVertical(lipgloss.Center, inputFields, buttons))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		b.String())
}

// Update
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
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

func (m *Model) SetSize(height, width int) {
	m.height = height
	m.width = width
}
