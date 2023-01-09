package whys

import (
	"sort"
	"strconv"
	"strings"

	"github.com/benhsm/goals/internal/data"
	"github.com/benhsm/goals/internal/ui/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const banner = `░█▀█░█▀▀░▀█▀░▀█▀░█░█░█▀▀░░░█▀▀░█▀█░█▀█░█░░░█▀▀
░█▀█░█░░░░█░░░█░░▀▄▀░█▀▀░░░█░█░█░█░█▀█░█░░░▀▀█
░▀░▀░▀▀▀░░▀░░▀▀▀░░▀░░▀▀▀░░░▀▀▀░▀▀▀░▀░▀░▀▀▀░▀▀▀
`

var (
	docStyle         = lipgloss.NewStyle().Margin(1, 2)
	descriptionStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Width(80).
			Height(2).
			Margin(0, 0, 0, 1).
			Padding(0, 0, 0, 1)
	}
	titleStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).Padding(0, 0, 0, 1).
			Margin(0, 0, 0, 1).
			Width(80)
	}
	listItemStyle         = lipgloss.NewStyle().Padding(0, 0, 0, 1)
	selectedlistItemStyle = listItemStyle.Copy().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				Padding(0, 0, 0, 0)
	prefixStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(color))
	}
)

type iostateEnum int

const (
	synced iostateEnum = iota
	unsynced
	syncing
)

type Model struct {
	common       common.Common
	whys         []data.Why
	focusIndex   int
	input        goalInputModel
	editing      bool
	adding       bool
	iostate      iostateEnum
	errMessage   string
	whysToDelete []data.Why
	height       int
	width        int
}

func New(c common.Common) *Model {
	return &Model{
		common: c,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.common.ReadWhys(data.Active)
}

func (m *Model) View() string {
	var b strings.Builder

	if m.editing {
		return m.input.View()
	} else {
		for i, g := range m.whys {
			listItem := m.WhyRender(g, strconv.Itoa(i+1))
			if i == m.focusIndex {
				b.WriteString(selectedlistItemStyle.
					Render(listItem))
			} else {
				b.WriteString(listItemStyle.
					Render(listItem))
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
		final := lipgloss.JoinVertical(lipgloss.Center, banner, b.String())
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, docStyle.Render(final))
	}
}

func (m *Model) WhyRender(w data.Why, prefix string) string {
	m.common.FigletOpts.FontName = "future"
	// TODO: This won't work with non-ascii prefixes
	bigPrefix, _ := m.common.Figlet.RenderOpts(prefix, m.common.FigletOpts)
	bigPrefix = strings.TrimRight(bigPrefix, "\n")
	title := titleStyle(w.Color).Render(w.Name)
	desc := descriptionStyle(w.Color).Render(w.Description)
	contents := lipgloss.JoinVertical(lipgloss.Left, title, desc)
	result := lipgloss.JoinHorizontal(lipgloss.Center, prefixStyle(w.Color).Render(bigPrefix), contents)
	return result
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				return m, m.common.ReadWhys(data.All)
			}
		case common.WhyDataMsg:
			if msg.Error != nil {
				m.errMessage = msg.Error.Error()
			}
			sort.Slice(msg.Data, func(i, j int) bool {
				return msg.Data[i].Number < msg.Data[j].Number
			})
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
					m.iostate = unsynced
				case "J":
					if m.focusIndex < len(m.whys)-1 {
						m.whys[m.focusIndex+1], m.whys[m.focusIndex] = m.whys[m.focusIndex], m.whys[m.focusIndex+1]
						m.focusIndex++
					}
					m.iostate = unsynced
				case "d":
					m.whysToDelete = append(m.whysToDelete, m.whys[m.focusIndex])
					m.whys = removeItemFromSlice(m.whys, m.focusIndex)
					m.iostate = unsynced
				case "a", "e":
					m.editing = true
					m.input = newGoalInput()
					m.input.SetSize(m.height, m.width)
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
						for i := range m.whys {
							m.whys[i].Number = i + 1
						}
						cmd = m.common.UpsertWhys(m.whys)
						cmds = append(cmds, cmd)
						cmd = m.common.DeleteWhys(m.whysToDelete)
						cmds = append(cmds, cmd)
						m.iostate = syncing
					}
				case "r":
					return m, m.common.ReadWhys(data.All)
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

func (m *Model) SetSize(height, width int) {
	m.height = height
	m.width = width
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
