package ui

import (
	"fmt"
	"strings"

	goalinput "github.com/benhsm/goals/internal/ui/goal_input"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

// I think the bubbles list component introduces too much overhead and isn't as customizable as I'd like it to be.

var (
	docStyle         = lipgloss.NewStyle().Margin(1, 2)
	descriptionStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(color)
	}
	titleStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).Padding(0, 1, 0, 1)
	}
	listItemStyle         = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	selectedlistItemStyle = listItemStyle.Copy().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				Padding(0, 0, 0, 0)
)

type goalItem struct {
	title, desc string
	color       lipgloss.Color
}

type goalListModel struct {
	goals         []goalItem
	focusIndex    int
	input         goalinput.Model
	editting      bool
	height, width int
}

func NewGoalList() goalListModel {

	glm := goalListModel{
		[]goalItem{
			{
				"be the very best",
				"that no one ever was. I need to beat all the gym leaders, and then the elite four. It's going to take a lot of work and I'll really need to train my party.",
				"#FF0000",
			},
			{
				"here's a really long one",
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.",
				"#5F909C",
			},
			{
				"catch 'em all",
				"Man, there's so many pokemon in the pokedex, there's a lot of them for me to catch. There's like, hundreds of them.",
				"#009900",
			},
		},
		0,
		goalinput.Model{},
		false,
		0,
		0,
	}

	return glm
}

func (m goalListModel) Init() tea.Cmd {
	return nil
}

func (m goalListModel) View() string {
	var b strings.Builder

	if m.editting {
		return m.input.View()
	} else {
		b.WriteString("Goals\n\n")
		for i, g := range m.goals {
			if i == m.focusIndex {
				b.WriteString(selectedlistItemStyle.Render(g.render(i)))
			} else {
				b.WriteString(listItemStyle.Render(g.render(i)))
			}
			b.WriteString("\n\n")
		}
		return docStyle.Render(b.String())
	}
}

func (g goalItem) render(index int) string {
	title := titleStyle(g.color).Render(fmt.Sprintf("%d) %s", index+1, g.title))
	desc := descriptionStyle(g.color).Render(wordwrap.String(g.desc, 80))
	return title + "\n" + desc
}

func (m goalListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.editting {
		m.input, cmd = m.input.Update(msg)
		if m.input.Done || m.input.Cancelled {
			m.editting = false
			m.input = goalinput.Model{}
		}
		return m, cmd
	} else {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.height = msg.Height
			m.width = msg.Width
		case tea.KeyMsg:
			switch {
			case msg.Type == tea.KeyCtrlC:
				return m, tea.Quit
			case msg.String() == "k":
				m.focusIndex--
			case msg.String() == "j":
				m.focusIndex++
			case msg.String() == "a":
				m.editting = true
				m.input = goalinput.New()
				m.input.SetSize(m.height, m.width)
				initCmd := m.input.Init()
				return m, initCmd
			}
		}
	}

	return m, tea.Batch(cmds...)
}

type goalListKeyMap struct {
	Up   key.Binding
	Down key.Binding
	Edit key.Binding
	Add  key.Binding
	Help key.Binding
	Quit key.Binding
}

var goalListDefaultKeys = goalListKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit goal"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add goal"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
