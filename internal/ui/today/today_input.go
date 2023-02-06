package today

import (
	"github.com/benhsm/goalie/internal/data"
	"github.com/benhsm/goalie/internal/ui/common"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputModel struct {
	common.Common
	textInput textarea.Model
	finished  bool
	whys      *[]data.Why

	help help.Model
	keys inputKeyMap
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
		whys:      &[]data.Why{},
		help:      help.New(),
		keys:      inputKeys,
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
		switch {
		case key.Matches(msg, m.keys.Done):
			m.finished = true
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m inputModel) View() string {
	badges := badgeStyle.Render(whyBadges(*m.whys))
	textBox := inputStyle.Render(m.textInput.View())
	prompt := "What are you doing towards your goals today?"
	prompt = promptStyle.Render(prompt)
	return lipgloss.JoinVertical(lipgloss.Center, badges, prompt, textBox, m.help.View(inputKeys))
}

type inputKeyMap struct {
	Done        key.Binding
	Quit        key.Binding
	ChangeFocus key.Binding
}

var inputKeys = inputKeyMap{
	Done: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "submit"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	ChangeFocus: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "change focus"),
	),
}

// Shorthelp is part of the key.Map interface
func (k inputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Done, k.Quit}
}

// FullHelp is part of the key.Map interface
func (k inputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}
