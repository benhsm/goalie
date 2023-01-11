package today

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/benhsm/goals/internal/data"
	"github.com/benhsm/goals/internal/ui/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	promptStyle = lipgloss.NewStyle().Bold(true)
	badgeStyle  = lipgloss.NewStyle().Margin(1, 0)
	inputStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true)
)

type Model struct {
	common.Common
	whys       []data.Why
	intentions []data.Intention

	date         time.Time
	inputPage    inputModel
	todayPage    todayModel
	outcomesPage tea.Model
	state        activePage

	height int
	width  int
}

type activePage int

const (
	inputActive activePage = iota
	todayActive
	outcomesActive
)

func New(c common.Common) *Model {
	return &Model{
		Common:    c,
		date:      getCurrentDay(),
		inputPage: newInputModel(c),
		todayPage: newTodayModel(c),
		//		outcomesPage: newReflectModel(),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.inputPage.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)

	case common.WhyDataMsg:
		if msg.Data != nil {
			m.whys = msg.Data
			m.inputPage.whys = &m.whys
			m.todayPage.whys = &m.whys
		}
	}

	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case inputActive:
		m.inputPage, cmd = m.inputPage.Update(msg)
		cmds = append(cmds, cmd)
		if m.inputPage.finished {
			input := m.inputPage.textInput.Value()
			parsedIntentions, err := parseIntentions(m.whys, input)
			if err != nil {
				m.inputPage.finished = false
			} else {
				m.todayPage.intentions = parsedIntentions
				m.state = todayActive
			}
		}
	case todayActive:
		m.todayPage, cmd = m.todayPage.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := strings.Builder{}

	year, month, day := m.date.Date()
	weekday := m.date.Weekday().String()
	fmt.Fprintf(&s, "%s %d, %s %d\n", weekday, day, month.String(), year)

	switch m.state {
	case inputActive:
		s.WriteString(m.inputPage.View())
	case todayActive:
		s.WriteString(m.todayPage.View())
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s.String())
}

func (m *Model) SetSize(height, width int) {
	m.height = height
	m.width = width
}

func getCurrentDay() time.Time {
	now := time.Now().Local()

	// For our purposes, the day is considered to begin/end at 4:00AM
	if now.Hour() < 4 {
		now = now.AddDate(0, 0, -1)
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func parseIntentions(whys []data.Why, input string) ([]data.Intention, error) {
	var results []data.Intention

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue // discard blank lines
		}
		intention := data.Intention{}
		intention.Content = line
		prefix, _, found := strings.Cut(line, ")")
		if !found {
			return nil, errors.New("No goal prefix")
		}
		codes := strings.Split(prefix, ",")
		for _, c := range codes {
			whyNum, err := strconv.Atoi(string(c))
			if err != nil || !(whyNum-1 >= 0 && whyNum-1 <= len(whys)-1) {
				return nil, errors.New("Invalid goal code")
			}
			intention.Whys = append(intention.Whys, &whys[whyNum-1])
		}
		results = append(results, intention)
	}
	return results, nil
}

func whyBadges(whys []data.Why) string {
	var lines []string
	var line strings.Builder
	for i, why := range whys {
		prefix := strconv.Itoa(i+1) + " "
		whyTitle := prefix + why.Name
		// need to use lipgloss.Width here to avoid counting the escape sequences
		if lipgloss.Width(line.String()+whyTitle) > 70 {
			lines = append(lines, line.String())
			line.Reset()
		}
		line.WriteString(common.WhyBadgeStyle(why.Color).Render((whyTitle)))
	}
	if line.Len() != 0 {
		lines = append(lines, line.String())
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}