package whys

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/benhsm/goals/internal/data"
	"github.com/benhsm/goals/internal/ui/common"
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

type iostateEnum int

const (
	synced iostateEnum = iota
	unsynced
	syncing
)

type Model struct {
	common.Common
	whys         []data.Why
	focusIndex   int
	input        goalInputModel
	editing      bool
	adding       bool
	iostate      iostateEnum
	errMessage   string
	whysToDelete []data.Why
}

func New(c common.Common) Model {
	return Model{
		Common: c,
	}
}

func (m Model) Init() tea.Cmd {
	return m.ReadWhys(data.Active)
}

func (m Model) View() string {
	var b strings.Builder

	if m.editing {
		return m.input.View()
	} else {
		b.WriteString("Goals\n\n")
		for i, g := range m.whys {
			if i == m.focusIndex {
				b.WriteString(selectedlistItemStyle.Render(render(g, strconv.Itoa(i+1))))
			} else {
				b.WriteString(listItemStyle.Render(render(g, strconv.Itoa(i+1))))
			}
			b.WriteString("\n\n")
		}
		switch m.iostate {
		case synced:
			b.WriteString("changes synced to database\n")
		case unsynced:
			b.WriteString("Unsaved modifications.\n")
		case syncing:
			b.WriteString("syncing with database...\n")
		}
		b.WriteString(m.errMessage)
		return frame.Render(b.String())
	}
}

func render(w data.Why, prefix string) string {
	title := titleStyle(w.Color).Render(fmt.Sprintf("%s) %s", prefix, w.Name))
	desc := descriptionStyle(w.Color).Render(wordwrap.String(w.Description, 80))
	return title + "\n" + desc
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.editing {
		m.input, cmd = m.input.Update(msg)
		if m.input.Done {
			m.editing = false
			if !m.input.Cancelled {
				m.iostate = unsynced
				if m.adding {
					newGoal := data.Why{
						Name:        m.input.TitleInput.Value(),
						Description: m.input.DescInput.Value(),
						Color:       m.input.Color,
					}
					m.whys = append(m.whys, newGoal)
					m.adding = false
				} else {
					m.whys[m.focusIndex].Name = m.input.TitleInput.Value()
					m.whys[m.focusIndex].Description = m.input.DescInput.Value()
					m.whys[m.focusIndex].Color = m.input.Color
				}
			}
			m.input = goalInputModel{}
		}
		return m, cmd
	} else {
		switch msg := msg.(type) {
		case common.ErrMsg:
			if msg.Error != nil {
				m.errMessage = msg.Error.Error()
			} else {
				return m, m.ReadWhys(data.All)
			}
		case common.WhyDataMsg:
			if msg.Error != nil {
				m.errMessage = msg.Error.Error()
			}
			m.whys = msg.Data
			m.iostate = synced
		case tea.WindowSizeMsg:
			m.SetSize(msg.Height, msg.Width)
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
						m.whys[m.focusIndex-1], m.whys[m.focusIndex] = m.whys[m.focusIndex], m.whys[m.focusIndex-1]
						m.focusIndex--
					}
				case "J":
					if m.focusIndex < len(m.whys)-1 {
						m.whys[m.focusIndex+1], m.whys[m.focusIndex] = m.whys[m.focusIndex], m.whys[m.focusIndex+1]
						m.focusIndex++
					}
				case "d":
					m.whysToDelete = append(m.whysToDelete, m.whys[m.focusIndex])
					m.whys = removeItemFromSlice(m.whys, m.focusIndex)
					m.iostate = unsynced
				case "a", "e":
					m.editing = true
					m.input = newGoalInput()
					m.input.SetSize(m.Height, m.Width)
					initCmd := m.input.Init()
					if string(msg.Runes) == "e" {
						m.input.TitleInput.SetValue(m.whys[m.focusIndex].Name)
						m.input.DescInput.SetValue(m.whys[m.focusIndex].Description)
						m.input.Color = m.whys[m.focusIndex].Color
					} else {
						m.adding = true
					}
					return m, initCmd
				case "s":
					if m.iostate == unsynced {
						cmd = m.UpsertWhys(m.whys)
						cmds = append(cmds, cmd)
						cmd = m.DeleteWhys(m.whysToDelete)
						cmds = append(cmds, cmd)
						m.iostate = syncing
					}
				case "r":
					return m, m.ReadWhys(data.All)
				}
			}
		}

		if m.focusIndex > len(m.whys)-1 {
			m.focusIndex = 0
		}
		if m.focusIndex < 0 {
			m.focusIndex = len(m.whys) - 1
		}

		return m, tea.Batch(cmds...)
	}
}

// Remove an item from a slice of items at the given index. This runs in O(n).
func removeItemFromSlice(i []data.Why, index int) []data.Why {
	if index >= len(i) {
		return i // noop
	}
	copy(i[index:], i[index+1:])
	i[len(i)-1] = data.Why{}
	return i[:len(i)-1]
}
