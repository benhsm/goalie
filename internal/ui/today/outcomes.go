package today

import (
	"fmt"
	"strconv"
	"time"

	"github.com/benhsm/goalie/internal/data"
	"github.com/benhsm/goalie/internal/ui/common"
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

		switch msg.Type {
		case tea.KeyCtrlD:
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
		case tea.KeyTab, tea.KeyShiftTab:
			if msg.Type == tea.KeyTab {
				m.focusIndex++
			}
			if msg.Type == tea.KeyShiftTab {
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
		case tea.KeyEsc:
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
				switch msg.Type {
				case tea.KeyEnter, tea.KeySpace:
					if m.sections[m.sectionIndex].intentions != nil {
						m.sections[m.sectionIndex].intentions[m.outcomeIndex].Done =
							!m.sections[m.sectionIndex].intentions[m.outcomeIndex].Done
					}
				case tea.KeyBackspace:
					if m.sections[m.sectionIndex].intentions != nil {
						m.sections[m.sectionIndex].intentions[m.outcomeIndex].Cancelled =
							!m.sections[m.sectionIndex].intentions[m.outcomeIndex].Cancelled
					}
				case tea.KeyRunes:
					switch msg.String() {
					case "j":
						m.outcomeIndex++
					case "k":
						m.outcomeIndex--
					case "l":
						m.sectionIndex++
					case "h":
						m.sectionIndex--
					case "y":
						m.sections[m.sectionIndex].enough = true
					case "n":
						m.sections[m.sectionIndex].enough = false
					case "a":
						cmd := m.sections[m.sectionIndex].addInput.Focus()
						cmds = append(cmds, cmd)
						return m, tea.Batch(cmds...)
					}
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
			enoughLine, " ["+checkBox+"]")
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
		section.addInput.Prompt = fmt.Sprintf("[+] %d) ", why.Number)
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
	miscSection.addInput.Prompt = "[+] &) "
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
