package today

import (
	"fmt"
	"strconv"

	"github.com/benhsm/goals/internal/data"
	"github.com/benhsm/goals/internal/ui/common"
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
		case outcomesFocus:
			if m.sections[m.sectionIndex].addInput.Focused() {
				switch msg.Type {
				case tea.KeyEnter:
					why := m.sections[m.sectionIndex].why
					content := fmt.Sprintf("%d) %s", why.Number, m.sections[m.sectionIndex].addInput.Value())
					newIntention := data.Intention{
						Whys:       []*data.Why{why},
						Content:    content,
						Unintended: true,
						Done:       true,
					}
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
	why := m.whys[m.sectionIndex]
	number := strconv.Itoa(why.Number)
	title := titleStyle(why.Color).Render(number + " " + why.Name)
	var s []string
	for i, intention := range m.sections[m.sectionIndex].intentions {
		var renderedIntention string
		if intention.Cancelled {
			renderedIntention = cancelledRender(intention)
		} else if intention.Done {
			renderedIntention = doneItemRender(intention)
		} else {
			renderedIntention = listItemRender(intention)
		}
		if m.outcomeIndex == i && m.focusIndex == outcomesFocus {
			s = append(s, selectedStyle.Render(renderedIntention))
		} else {
			s = append(s, renderedIntention)
		}
	}
	s = append(s, m.sections[m.sectionIndex].addInput.View())
	//	m.sections[m.sectionIndex].input.PromptStyle.Background(why.Color)
	//	m.sections[m.sectionIndex].input.TextStyle.Background(why.Color).Border(lipgloss.NormalBorder(), true)
	//	m.sections[m.sectionIndex].input.BackgroundStyle.Background(why.Color)
	//	m.sections[m.sectionIndex].input.PlaceholderStyle.Background(why.Color)
	inputBox := lipgloss.NewStyle().
		BorderForeground(why.Color).
		Border(lipgloss.RoundedBorder(), true).
		Width(50).
		Padding(0, 0, 0, 1).
		Render(m.sections[m.sectionIndex].input.View())

	enoughLine := lipgloss.JoinHorizontal(lipgloss.Center,
		"Is this enough? ", "<==")
	enoughLine = lipgloss.NewStyle().Foreground(why.Color).Render(enoughLine)
	if m.sections[m.sectionIndex].enough {
		enoughLine = lipgloss.JoinHorizontal(lipgloss.Center,
			enoughLine, " ["+checkmark+"]")
	} else {
		enoughLine = lipgloss.JoinHorizontal(lipgloss.Center,
			enoughLine, " [X]")
	}

	outcomeBox := lipgloss.JoinVertical(lipgloss.Left, s...)
	outcomeBox = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true).
		BorderForeground(why.Color).
		Width(50).
		Render(outcomeBox)
	rightBox := lipgloss.JoinVertical(lipgloss.Left, title, outcomeBox, enoughLine, inputBox)
	goalCount := fmt.Sprintf("Goal %d/%d to review", m.sectionIndex+1, len(m.whys))
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
	if len(result) == 0 {
		panic("empty")
	}
	return result
}
