package whys

import (
	"strconv"
	"strings"

	"github.com/benhsm/goalie/internal/data"
	"github.com/benhsm/goalie/internal/ui/common"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
	keys         keyMap
	help         help.Model
}

func New(c common.Common) *Model {
	return &Model{
		common: c,
		keys:   keys,
		help:   help.New(),
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
			listItem := m.WhyRender(g, strconv.Itoa(i))
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
		final := lipgloss.JoinVertical(lipgloss.Center, banner, b.String(), m.help.View(m.keys))
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
				} else {
					m.whys[m.focusIndex].Name = m.input.TitleInput.Value()
					m.whys[m.focusIndex].Description = m.input.DescInput.Value()
					m.whys[m.focusIndex].Color = m.input.Color
				}
			}
			m.adding = false
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
			m.whys = msg.Data
			m.iostate = synced
		case tea.WindowSizeMsg:
			m.SetSize(msg.Height, msg.Width)
			//			m.help.Width = msg.Width
		case tea.KeyMsg:
			switch {
			case key.Matches(msg, m.keys.Quit):
				return m, tea.Quit
			case key.Matches(msg, m.keys.Help):
				m.help.ShowAll = !m.help.ShowAll
			case key.Matches(msg, m.keys.Up):
				m.focusIndex--
			case key.Matches(msg, m.keys.Down):
				m.focusIndex++
			case key.Matches(msg, m.keys.ShiftUp):
				if m.focusIndex > 0 {
					m.whys[m.focusIndex-1], m.whys[m.focusIndex] = m.whys[m.focusIndex], m.whys[m.focusIndex-1]
					m.focusIndex--
				}
				m.iostate = unsynced
			case key.Matches(msg, m.keys.ShiftDown):
				if m.focusIndex < len(m.whys)-1 {
					m.whys[m.focusIndex+1], m.whys[m.focusIndex] = m.whys[m.focusIndex], m.whys[m.focusIndex+1]
					m.focusIndex++
				}
				m.iostate = unsynced
			case key.Matches(msg, m.keys.Delete):
				m.whysToDelete = append(m.whysToDelete, m.whys[m.focusIndex])
				m.whys = removeItemFromSlice(m.whys, m.focusIndex)
				m.iostate = unsynced
			case key.Matches(msg, m.keys.Add, m.keys.Edit):
				m.editing = true
				m.input = newGoalInput()
				m.input.SetSize(m.height, m.width)
				initCmd := m.input.Init()
				if key.Matches(msg, m.keys.Edit) {
					m.input.TitleInput.SetValue(m.whys[m.focusIndex].Name)
					m.input.DescInput.SetValue(m.whys[m.focusIndex].Description)
					m.input.Color = m.whys[m.focusIndex].Color
				} else {
					m.adding = true
				}
				return m, initCmd
			case key.Matches(msg, m.keys.Sync):
				if m.iostate == unsynced {
					for i := range m.whys {
						m.whys[i].Number = i
					}
					cmd = m.common.UpsertWhys(m.whys)
					cmds = append(cmds, cmd)
					cmd = m.common.DeleteWhys(m.whysToDelete)
					cmds = append(cmds, cmd)
					m.iostate = syncing
				}
			case key.Matches(msg, m.keys.Reload):
				return m, m.common.ReadWhys(data.All)
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

type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	ShiftUp   key.Binding
	ShiftDown key.Binding
	Help      key.Binding
	Quit      key.Binding
	Delete    key.Binding
	Edit      key.Binding
	Add       key.Binding
	Reload    key.Binding
	Sync      key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	ShiftUp: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "switch up"),
	),
	ShiftDown: key.NewBinding(
		key.WithKeys("J"),
		key.WithHelp("J", "switch down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete item"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add item"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit item"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "discard changes"),
	),
	Sync: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sync changes to database"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ShiftDown, k.ShiftUp}, // first column
		{k.Add, k.Edit},                        // second column
		{k.Reload, k.Sync, k.Help, k.Quit},
	}
}
