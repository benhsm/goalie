package today

import (
	"strconv"
	"strings"
	"unicode"

	"github.com/benhsm/goals/internal/data"
	"github.com/benhsm/goals/internal/ui/common"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inputModel struct {
	common.Common
	textInput textarea.Model
	finished  bool
	whys      []data.Why
}

func newInputModel(c common.Common) inputModel {
	ti := textarea.New()
	ti.SetHeight(10)
	ti.SetWidth(50)
	ti.Placeholder = "Write some intentions for today here."
	ti.Focus()

	return inputModel{
		Common:    c,
		textInput: ti,
		whys:      []data.Why{},
	}
}

func (m inputModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m inputModel) Update(msg tea.Msg) (inputModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+d":
			if m.textInput.Focused() {
				m.textInput.Blur()
			}
			m.finished = true
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m inputModel) View() string {
	var b strings.Builder
	for i, why := range m.whys {
		prefix := strconv.Itoa(i+1) + " "
		whyTitle := prefix + why.Name
		// need to use lipgloss.Width here to avoid counting the escape sequences
		if lipgloss.Width(b.String()+whyTitle) > 70 {
			b.WriteString("\n")
		}
		b.WriteString(common.WhyBadgeStyle(why.Color).Render((whyTitle)))
	}
	b.WriteString("\n")
	textBox := m.textInput.View()
	prompt := "What are you doing towards what's important today?\n"
	return lipgloss.JoinVertical(lipgloss.Left, b.String(), prompt, textBox)
}

func parseIntentions(input string) []data.Intention {
	var result []data.Intention

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		intention := data.Intention{}
		intention.Content = line
		intention.Content = line
		for _, c := range line {
			if c == ')' {
				break
			}
			if unicode.IsDigit(c) {
				//digit, _ := strconv.Atoi(string(c))
				// TODO: add code here for adding relevant goals to
				// intention.Whys
				panic("unimplemented")
			}
		}
		result = append(result, intention)
	}
	return result
}
