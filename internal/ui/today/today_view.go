package today

import (
	"fmt"
	"strings"
	"time"

	"github.com/benhsm/goalie/internal/data"
	"github.com/benhsm/goalie/internal/ui/common"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	listBoxStyle = lipgloss.NewStyle().
			Height(10).
			Width(50).
			Border(lipgloss.RoundedBorder(), true).
			Margin(1, 0, 0, 0)
	selectedStyle = lipgloss.NewStyle().
			Bold(true)
	checkBox = "  [" + lipgloss.NewStyle().SetString("✓").
			Foreground(lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}).
			String() + "] "
	boldCheck = selectedStyle.Render("• [") + lipgloss.NewStyle().SetString("✓").
			Foreground(lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}).
			Bold(true).
			String() + selectedStyle.Render("] ")
	cancelledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
			Strikethrough(true)
	cancelledBox      = "  " + cancelledStyle.Render("[x] ")
	selectedCancelled = selectedStyle.Render("• ") + cancelledStyle.Bold(true).Render("[x] ")
	listItemStyle     = func(i data.Intention) lipgloss.TerminalColor {
		var color lipgloss.TerminalColor
		color = lipgloss.NoColor{}
		if len(i.Whys) > 0 {
			color = i.Whys[0].Color
		}
		return color
	}
	pomos = func(i data.Intention) string {
		return " " + strings.Repeat("🍅", i.Pomos)
	}
	listItemRender = func(i data.Intention, selected bool) string {
		color := listItemStyle(i)
		var prefix string
		if selected {
			prefix = selectedStyle.Render("• [ ] ")
		} else {
			prefix = "  [ ] "
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, prefix, lipgloss.NewStyle().
			Foreground(color).
			Width(44).
			Bold(selected).
			Render(i.Content+pomos(i)))
	}
	doneItemRender = func(i data.Intention, selected bool) string {
		color := listItemStyle(i)
		var prefix string
		if selected {
			prefix = boldCheck
		} else {
			prefix = checkBox
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, prefix, lipgloss.NewStyle().
			Foreground(color).
			Width(44).
			Bold(selected).
			Strikethrough(true).
			Render(i.Content+pomos(i)))
	}
	cancelledRender = func(i data.Intention, selected bool) string {
		var prefix string
		if selected {
			prefix = selectedCancelled
		} else {
			prefix = cancelledBox
		}
		return lipgloss.JoinHorizontal(lipgloss.Top, prefix, lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#969B86", Dark: "#696969"}).
			//			Background(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}).
			Strikethrough(true).
			Width(50).
			Bold(selected).
			Render(i.Content))
	}
)

type todayModel struct {
	common     common.Common
	whys       *[]data.Why
	intentions []data.Intention
	input      textinput.Model
	date       *time.Time

	focusIndex int
	adding     bool
	finished   bool

	height int
	width  int
}

func newTodayModel(c common.Common) todayModel {
	return todayModel{
		common: c,
		whys:   &[]data.Why{},
	}
}

func (m *todayModel) Init() tea.Cmd {
	return nil
}

func (m todayModel) Update(msg tea.Msg) (todayModel, tea.Cmd) {
	//	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "j":
			m.focusIndex++
		case "k":
			m.focusIndex--
		case "a":
			m.adding = true
		case "ctrl+d":
			m.finished = true
		// keys that modify the intention list
		case "J", "K", "c", "space", "enter", "p", "P":
			switch msg.String() {
			case "J":
				if m.focusIndex < len(m.intentions)-1 {
					m.intentions[m.focusIndex+1], m.intentions[m.focusIndex] =
						m.intentions[m.focusIndex], m.intentions[m.focusIndex+1]
					m.focusIndex++
				}
				for i := range m.intentions {
					m.intentions[i].Position = i
				}
			case "K":
				if m.focusIndex > 0 {
					m.intentions[m.focusIndex-1], m.intentions[m.focusIndex] =
						m.intentions[m.focusIndex], m.intentions[m.focusIndex-1]
					m.focusIndex--
				}
				for i := range m.intentions {
					m.intentions[i].Position = i
				}
			case "c":
				m.intentions[m.focusIndex].Cancelled = !m.intentions[m.focusIndex].Cancelled
			case "space", "enter":
				m.intentions[m.focusIndex].Done = !m.intentions[m.focusIndex].Done
			case "p", "P":
				if msg.String() == "p" {
					m.intentions[m.focusIndex].Pomos++
				}
				if msg.String() == "P" && m.intentions[m.focusIndex].Pomos > 0 {
					m.intentions[m.focusIndex].Pomos--
				}
			}
			cmds = append(cmds, m.common.UpsertIntentions(m.intentions))
			cmds = append(cmds, m.common.GetDaysIntentions(*m.date))
		}
	}

	if m.focusIndex < 0 {
		m.focusIndex = len(m.intentions) - 1
	}
	if m.focusIndex > len(m.intentions)-1 {
		m.focusIndex = 0
	}
	return m, tea.Sequence(cmds...)
}

func (m *todayModel) View() string {
	var s []string
	var totalIntentions int
	var doneIntentions int
	for _, intention := range m.intentions {
		if !intention.Cancelled {
			totalIntentions++
			if intention.Done {
				doneIntentions++
			}
		}
	}
	prompt := promptStyle.Render(fmt.Sprintf("\n%d intentions for today, %d/%d done", totalIntentions, doneIntentions, totalIntentions))
	for i, intention := range m.intentions {
		var renderedIntention string
		selected := false
		if m.focusIndex == i {
			selected = true
		}
		if intention.Cancelled {
			renderedIntention = cancelledRender(intention, selected)
		} else if intention.Done {
			renderedIntention = doneItemRender(intention, selected)
		} else {
			renderedIntention = listItemRender(intention, selected)
		}
		s = append(s, renderedIntention)
	}
	listBox := lipgloss.JoinVertical(lipgloss.Left, s...)
	listBox = listBoxStyle.Render(listBox)
	badges := badgeStyle.Render(whyBadges(*m.whys))
	return lipgloss.JoinVertical(lipgloss.Center, prompt, listBox, badges)
}

func (m *todayModel) SetSize(height, width int) {
	m.height = height
	m.width = width
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
