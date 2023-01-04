package ui

import (
	"github.com/benhsm/goals/internal/ui/common"
	whys "github.com/benhsm/goals/internal/ui/whys"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	goalsView = iota

// todayPage
// reviewsPage
)

// Model is the main UI model
type Model struct {
	common.Common
	pages []tea.Model
	state int
}

func New() Model {
	c := common.NewCommon()
	result := Model{Common: c}
	result.pages = make([]tea.Model, 4)

	result.pages[goalsView] = whys.New(c)
	return result
}

func (m Model) Init() tea.Cmd {
	return m.pages[goalsView].Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.pages[m.state], cmd = m.pages[m.state].Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.pages[m.state].View()
}
