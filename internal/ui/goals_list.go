package ui

import (
	"fmt"
	"strings"

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
	goals      []goalItem
	focusIndex int
	input      goalInputModel
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
		NewGoalInput(),
	}

	return glm
}

func (m goalListModel) Init() tea.Cmd {
	return nil
}

func (m goalListModel) View() string {
	var b strings.Builder
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

func (g goalItem) render(index int) string {
	title := titleStyle(g.color).Render(fmt.Sprintf("%d) %s", index+1, g.title))
	desc := descriptionStyle(g.color).Render(wordwrap.String(g.desc, 80))
	return title + "\n" + desc
}

func (m goalListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.String() == "k":
			m.focusIndex--
		case msg.String() == "j":
			m.focusIndex++
		}
	}

	return m, cmd
}
