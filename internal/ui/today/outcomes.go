package today

import (
	"fmt"
	"strconv"
	"time"

	"github.com/benhsm/goalie/internal/data"
	"github.com/benhsm/goalie/internal/ui/common"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().
			Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).Padding(0, 1, 0, 1)
	}
	sectionStyle = lipgloss.NewStyle().Margin(1, 0, 0, 0)
)

type outcomeModel struct {
	common.Common
	whys         []data.Why
	intentions   []data.Intention
	outcomeIndex int

	date       *time.Time
	focusIndex int

	sections     []outcomeSection
	sectionIndex int
	finished     bool

	help help.Model
	keys outcomesKeyMap
}

type outcomeSection struct {
	why        *data.Why
	intentions []data.Intention
	enough     bool
	reflection string
	addInput   textinput.Model
	input      textinput.Model
}

const (
	outcomesFocus = iota
	reflectFocus
)

func newOutcomeModel(c common.Common, whys []data.Why, intentions []data.Intention) outcomeModel {
	return outcomeModel{
		Common:     c,
		whys:       whys,
		intentions: intentions,
		sections:   makeOutcomeSections(whys, intentions),
		help:       help.New(),
		keys:       outcomeKeys,
	}
}

func (m outcomeModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m outcomeModel) Update(msg tea.Msg) (outcomeModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
	case tea.KeyMsg:

		switch {
		case key.Matches(msg, m.keys.SubmitOutcomes):
			var outcomes []data.Intention
			var days []data.Day
			for i := range m.sections {
				outcomes = append(outcomes, m.sections[i].intentions...)
				var day data.Day
				day.Date = *m.date
				if m.sections[i].why != nil {
					day.Why = *m.sections[i].why
				}
				day.Enough = m.sections[i].enough
				day.Reflection = m.sections[i].input.Value()
				days = append(days, day)
			}
			for i := range outcomes {
				outcomes[i].Outcome = true
			}
			cmd = m.UpsertIntentions(outcomes)
			cmds = append(cmds, cmd)
			cmd = m.UpsertDayReview(days)
			cmds = append(cmds, cmd)
			cmds = append(cmds, m.GetDaysIntentions(*m.date))
			return m, tea.Sequence(cmds...)
		case key.Matches(msg, m.keys.ChangeFocus, m.keys.ChangeFocusBack):
			if key.Matches(msg, m.keys.ChangeFocus) {
				m.focusIndex++
			}
			if key.Matches(msg, m.keys.ChangeFocusBack) {
				m.focusIndex--
			}

			if m.focusIndex < outcomesFocus {
				m.focusIndex = reflectFocus
			}
			if m.focusIndex > reflectFocus {
				m.focusIndex = outcomesFocus
			}

			if m.sections[m.sectionIndex].addInput.Focused() {
				m.sections[m.sectionIndex].addInput.Blur()
			}

			if m.focusIndex == reflectFocus {
				cmd = m.sections[m.sectionIndex].input.Focus()
				cmds = append(cmds, cmd)
			} else {
				m.sections[m.sectionIndex].input.Blur()
			}

			return m, tea.Batch(cmds...)
		case key.Matches(msg, m.keys.Escape):
			if m.sections[m.sectionIndex].addInput.Focused() {
				m.sections[m.sectionIndex].addInput.Blur()
			}
		}

		// component specific keys
		switch m.focusIndex {
		case reflectFocus:
			switch msg.Type {
			case tea.KeyEnter:
				m.sections[m.sectionIndex].input.Blur()
				m.sectionIndex++
				m.focusIndex = outcomesFocus
			}
		case outcomesFocus:
			if m.sections[m.sectionIndex].addInput.Focused() {
				switch msg.Type {
				case tea.KeyEnter:
					var why *data.Why
					var content string
					var newIntention data.Intention

					if m.sections[m.sectionIndex].why != nil {
						why = m.sections[m.sectionIndex].why
						content = fmt.Sprintf("%d) %s", why.Number, m.sections[m.sectionIndex].addInput.Value())
						newIntention = data.Intention{
							Whys: []*data.Why{why},
						}
					} else {
						content = fmt.Sprintf("&) %s", m.sections[m.sectionIndex].addInput.Value())
						newIntention = data.Intention{}
					}
					newIntention.Content = content
					newIntention.Unintended = true
					newIntention.Done = true
					newIntention.Date = *m.date

					m.sections[m.sectionIndex].addInput.Reset()
					m.sections[m.sectionIndex].intentions =
						append(m.sections[m.sectionIndex].intentions, newIntention)
					m.sections[m.sectionIndex].addInput.Blur()
				}
			} else {
				switch {
				case key.Matches(msg, m.keys.MarkDone):
					if m.sections[m.sectionIndex].intentions != nil {
						m.sections[m.sectionIndex].intentions[m.outcomeIndex].Done =
							!m.sections[m.sectionIndex].intentions[m.outcomeIndex].Done
					}
				case key.Matches(msg, m.keys.Cancel):
					if m.sections[m.sectionIndex].intentions != nil {
						m.sections[m.sectionIndex].intentions[m.outcomeIndex].Cancelled =
							!m.sections[m.sectionIndex].intentions[m.outcomeIndex].Cancelled
					}
				case key.Matches(msg, m.keys.Down):
					m.outcomeIndex++
				case key.Matches(msg, m.keys.Up):
					m.outcomeIndex--
				case key.Matches(msg, m.keys.Right):
					m.sectionIndex++
				case key.Matches(msg, m.keys.Left):
					m.sectionIndex--
				case key.Matches(msg, m.keys.Yes):
					m.sections[m.sectionIndex].enough = true
				case key.Matches(msg, m.keys.Yes):
					m.sections[m.sectionIndex].enough = false
				case key.Matches(msg, m.keys.Add):
					cmd := m.sections[m.sectionIndex].addInput.Focus()
					cmds = append(cmds, cmd)
					return m, tea.Batch(cmds...)
				case key.Matches(msg, m.keys.Quit):
					return m, tea.Quit
				case key.Matches(msg, m.keys.Help):
					m.help.ShowAll = !m.help.ShowAll
				}
			}
		}
	}

	if m.sectionIndex < 0 {
		m.sectionIndex = len(m.sections) - 1
	}
	if m.sectionIndex > len(m.sections)-1 {
		m.sectionIndex = 0
	}

	if len(m.sections[m.sectionIndex].intentions) > 0 {
		if m.outcomeIndex < 0 {
			m.outcomeIndex = len(m.sections[m.sectionIndex].intentions) - 1
		}
		if m.outcomeIndex > len(m.sections[m.sectionIndex].intentions)-1 {
			m.outcomeIndex = 0
		}
	}

	for i := range m.sections {
		// each input model will only respond if focused,
		// so we can update all of them
		m.sections[i].input, cmd = m.sections[i].input.Update(msg)
		cmds = append(cmds, cmd)

		m.sections[i].addInput, cmd = m.sections[i].addInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m outcomeModel) View() string {
	prompt := promptStyle.Render("Reflect on what you did towards your goals today.")

	var why *data.Why
	var prefix string
	var color lipgloss.Color
	var name string
	if m.sections[m.sectionIndex].why != nil {
		why = m.sections[m.sectionIndex].why
		prefix = strconv.Itoa(why.Number)
		color = why.Color
		name = why.Name
	} else {
		prefix = "&"
		color = lipgloss.Color("#808080")
		name = "MISC"
	}

	title := titleStyle(color).Render(prefix + " " + name)
	var s []string
	for i, intention := range m.sections[m.sectionIndex].intentions {
		var renderedIntention string
		var selected bool
		if m.outcomeIndex == i && m.focusIndex == outcomesFocus {
			selected = true
		} else {
			selected = false
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
	s = append(s, m.sections[m.sectionIndex].addInput.View())
	inputBox := lipgloss.NewStyle().
		BorderForeground(color).
		Border(lipgloss.RoundedBorder(), true).
		Width(50).
		Padding(0, 0, 0, 1).
		Render(m.sections[m.sectionIndex].input.View())

	enoughLine := lipgloss.JoinHorizontal(lipgloss.Center,
		"Is this enough? ", "<==")
	enoughLine = lipgloss.NewStyle().Foreground(color).Render(enoughLine)
	if m.sections[m.sectionIndex].enough {
		enoughLine = lipgloss.JoinHorizontal(lipgloss.Center,
			enoughLine, checkBox)
	} else {
		enoughLine = lipgloss.JoinHorizontal(lipgloss.Center,
			enoughLine, " [X]")
	}

	outcomeBox := lipgloss.JoinVertical(lipgloss.Left, s...)
	outcomeBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).
		BorderForeground(color).
		Width(50).
		Render(outcomeBox)
	rightBox := lipgloss.JoinVertical(lipgloss.Left, title, outcomeBox, enoughLine, inputBox)
	goalCount := fmt.Sprintf("Page %d/%d to review", m.sectionIndex+1, len(m.sections))
	rightBox = lipgloss.JoinVertical(lipgloss.Right, prompt, "", rightBox, goalCount)
	rightBox = lipgloss.JoinVertical(lipgloss.Center, rightBox, "", m.help.View(outcomeKeys))

	return sectionStyle.Render(rightBox)
}

func makeOutcomeSections(whys []data.Why, intentions []data.Intention) []outcomeSection {
	result := []outcomeSection{}
	for i, why := range whys {
		section := outcomeSection{}
		section.why = &whys[i]
		for _, intention := range intentions {
			for _, assocWhy := range intention.Whys {
				if assocWhy.ID == why.ID {
					section.intentions = append(section.intentions, intention)
					break
				}
			}
		}
		section.addInput = textinput.New()
		section.addInput.Width = 42
		section.addInput.Prompt = fmt.Sprintf("  [+] %d) ", why.Number)
		section.addInput.Placeholder = ""

		section.input = textinput.New()
		section.input.Width = 48
		section.input.Prompt = ""
		section.input.Placeholder = "say more..."
		result = append(result, section)
	}

	miscSection := outcomeSection{}
	for _, intention := range intentions {
		if len(intention.Whys) == 0 {
			miscSection.intentions = append(miscSection.intentions, intention)
		}
	}
	miscSection.addInput = textinput.New()
	miscSection.addInput.Width = 42
	miscSection.addInput.Prompt = "  [+] &) "
	miscSection.addInput.Placeholder = ""

	miscSection.input = textinput.New()
	miscSection.input.Width = 48
	miscSection.input.Prompt = ""
	miscSection.input.Placeholder = "overall remarks for today"
	result = append(result, miscSection)

	if len(result) == 0 {
		panic("empty")
	}
	return result
}

type outcomesKeyMap struct {
	Up              key.Binding
	Down            key.Binding
	Left            key.Binding
	Right           key.Binding
	Yes             key.Binding
	No              key.Binding
	ShiftUp         key.Binding
	ShiftDown       key.Binding
	Help            key.Binding
	Quit            key.Binding
	Cancel          key.Binding
	Add             key.Binding
	MarkDone        key.Binding
	AssignPomo      key.Binding
	UnassignPomo    key.Binding
	SubmitOutcomes  key.Binding
	ChangeFocus     key.Binding
	ChangeFocusBack key.Binding
	Escape          key.Binding
}

var outcomeKeys = outcomesKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev goal"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next goal"),
	),
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "+enough"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "-enough"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "cancel"),
	),
	Add: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add"),
	),
	MarkDone: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "check"),
	),
	AssignPomo: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "+pomo"),
	),
	UnassignPomo: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "-pomo"),
	),
	SubmitOutcomes: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "done"),
	),
	ChangeFocus: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "+focus"),
	),
	ChangeFocusBack: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "-focus"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "unfocus"),
	),
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k outcomesKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k outcomesKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},                   // first column
		{k.Add, k.MarkDone, k.AssignPomo, k.UnassignPomo}, // second column
		{k.Yes, k.No, k.SubmitOutcomes, k.Help, k.Quit},
		{k.ChangeFocus, k.ChangeFocusBack, k.Escape},
	}
}
