package ui

import (
	"fmt"
	"strings"

	goalinput "github.com/benhsm/goals/internal/ui/goal_input"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

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
	editing       bool
	adding        bool
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

	if m.editing {
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

	if m.editing {
		m.input, cmd = m.input.Update(msg)
		if m.input.Done {
			m.editing = false
			if !m.input.Cancelled {
				if m.adding {
					newGoal := goalItem{
						title: m.input.TitleInput.Value(),
						desc:  m.input.DescInput.Value(),
						color: m.input.Color,
					}
					m.goals = append(m.goals, newGoal)
					m.adding = false
				} else {
					m.goals[m.focusIndex].title = m.input.TitleInput.Value()
					m.goals[m.focusIndex].desc = m.input.DescInput.Value()
					m.goals[m.focusIndex].color = m.input.Color
				}
			}
			m.input = goalinput.Model{}
		}
		return m, cmd
	} else {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.height = msg.Height
			m.width = msg.Width
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyRunes:
				switch string(msg.Runes) {
				case "k":
					m.focusIndex--
				case "j":
					m.focusIndex++
				case "K":
					if m.focusIndex > 0 {
						m.goals[m.focusIndex-1], m.goals[m.focusIndex] = m.goals[m.focusIndex], m.goals[m.focusIndex-1]
						m.focusIndex--
					}
				case "J":
					if m.focusIndex < len(m.goals)-1 {
						m.goals[m.focusIndex+1], m.goals[m.focusIndex] = m.goals[m.focusIndex], m.goals[m.focusIndex+1]
						m.focusIndex++
					}
				case "d":
					m.goals = removeItemFromSlice(m.goals, m.focusIndex)
				case "a", "e":
					m.editing = true
					m.input = goalinput.New()
					m.input.SetSize(m.height, m.width)
					initCmd := m.input.Init()
					if string(msg.Runes) == "e" {
						m.input.TitleInput.SetValue(m.goals[m.focusIndex].title)
						m.input.DescInput.SetValue(m.goals[m.focusIndex].desc)
						m.input.Color = m.goals[m.focusIndex].color
					} else {
						m.adding = true
					}
					return m, initCmd
				}
			}
		}

		if m.focusIndex > len(m.goals)-1 {
			m.focusIndex = 0
		}
		if m.focusIndex < 0 {
			m.focusIndex = len(m.goals) - 1
		}

		return m, tea.Batch(cmds...)
	}
}

// Remove an item from a slice of items at the given index. This runs in O(n).
func removeItemFromSlice(i []goalItem, index int) []goalItem {
	if index >= len(i) {
		return i // noop
	}
	copy(i[index:], i[index+1:])
	i[len(i)-1] = goalItem{}
	return i[:len(i)-1]
}
