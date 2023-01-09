package ui

import (
	"github.com/benhsm/goals/internal/ui/common"
	"github.com/benhsm/goals/internal/ui/today"
	whys "github.com/benhsm/goals/internal/ui/whys"
	tea "github.com/charmbracelet/bubbletea"
)

type page int

const (
	todayPage page = iota
	whysPage

// reviewsPage
)

// Model is the main UI model
type Model struct {
	common.Common
	pages      []common.Component
	activePage page
}

func New() Model {
	c := common.NewCommon()
	result := Model{Common: c}
	result.pages = make([]common.Component, 2)

	result.pages[whysPage] = whys.New(c)
	result.pages[todayPage] = today.New(c)
	return result
}

func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	cmd = m.pages[whysPage].Init()
	cmds = append(cmds, cmd)

	cmd = m.pages[todayPage].Init()
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
		for _, p := range m.pages {
			if p != nil {
				p.SetSize(msg.Height, msg.Width)
			}
		}
	case common.WhyDataMsg:
		// All pages need to be updated with current whys
		for page := range m.pages {
			p, cmd := m.pages[page].Update(msg)
			m.pages[page] = p.(common.Component)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	case tea.KeyMsg:
		if msg.String() == "f1" {
			m.activePage = whysPage
			cmd := m.pages[m.activePage].Init()
			cmds = append(cmds, cmd)
		}
		if msg.String() == "f2" {
			m.activePage = todayPage
			cmd := m.pages[m.activePage].Init()
			cmds = append(cmds, cmd)
		}
	}
	pageModel, cmd := m.pages[m.activePage].Update(msg)
	m.pages[m.activePage] = pageModel.(common.Component)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.pages[m.activePage].View()
}
