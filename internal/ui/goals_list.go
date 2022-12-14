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
	docStyle        = lipgloss.NewStyle().Margin(1, 2)
	foregroundColor = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(color)
	}
	titleStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).Padding(0, 1)
	}
)

type goalItem struct {
	title, desc string
	color       lipgloss.Color
}

type goalListModel struct {
	goals []goalItem
}

func NewGoalList() goalListModel {

	glm := goalListModel{
		[]goalItem{
			{
				"be the very best",
				"that no one ever was. I need to beat all the gym leaders, and then the elite four. It's going to take a lot of work and I'll really need to train my team.",
				"#FF0000",
			},
			{
				"catch 'em all",
				"Man, there's so many pokemon in the pokedex, there's a lot of them for me to catch. Like, hundreds",
				"#009900",
			},
		},
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
		b.WriteString(titleStyle(g.color).Render(fmt.Sprintf("%d) %s", i+1, g.title)))
		b.WriteString("\n")
		b.WriteString(foregroundColor(g.color).Render(wordwrap.String(g.desc, 80)))
		b.WriteString("\n\n")
	}
	return docStyle.Render(b.String())
}

func (m goalListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	return m, cmd
}
