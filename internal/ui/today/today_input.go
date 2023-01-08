package today

import (
	"github.com/benhsm/goals/internal/ui/common"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type inputModel struct {
	common.Common
	textInput textarea.Model
	finished  bool
}

func newInputModel(c common.Common) inputModel {
	ti := textarea.New()
	ti.SetHeight(10)
	ti.SetWidth(50)
	ti.Placeholder = "Write some intentions for today here."
	ti.Focus()

	return inputModel{
		Common:    c,
		textInput: ti,
	}
}

func (m inputModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m inputModel) Update(msg tea.Msg) (inputModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			if m.textInput.Focused() {
				m.textInput.Blur()
			}
			m.finished = true
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m inputModel) View() string {
	s := m.textInput.View()
	return s
}
