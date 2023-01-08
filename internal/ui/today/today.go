package today

import (
	"fmt"
	"strings"
	"time"

	"github.com/benhsm/goals/internal/data"
	"github.com/benhsm/goals/internal/ui/common"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	common.Common
	whys       []data.Why
	intentions []data.Intention

	date         time.Time
	inputPage    inputModel
	todayPage    tea.Model
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
		date:      time.Now(),
		inputPage: newInputModel(c),
		//		todayPage:    newTodaymodel(),
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
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := strings.Builder{}

	year, month, day := m.date.Date()
	weekday := m.date.Weekday().String()
	fmt.Fprintf(&s, "%s %d, %s %d\n\n", weekday, day, month.String(), year)

	s.WriteString(m.inputPage.View())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s.String())
}

func (m *Model) SetSize(height, width int) {
	m.height = height
	m.width = width
}
