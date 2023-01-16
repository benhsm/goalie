package common

import (
	_ "embed"
	"sort"
	"time"

	"github.com/benhsm/goals/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mbndr/figlet4go"
)

//go:embed future.tlf
var fontFuture []byte

var (
	WhyBadgeStyle = func(color lipgloss.Color) lipgloss.Style {
		return lipgloss.NewStyle().Background(color).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1, 0, 1).
			Margin(0, 1, 0, 0).Bold(true)
	}
)

type Component interface {
	tea.Model
	SetSize(height, width int)
}

// Common is a struct all components should embed
type Common struct {
	Width      int
	Height     int
	Zone       *zone.Manager
	Store      data.Store
	Figlet     *figlet4go.AsciiRender
	FigletOpts *figlet4go.RenderOptions
}

func NewCommon() Common {
	figlet := figlet4go.NewAsciiRender()
	figletOpts := figlet4go.NewRenderOptions()
	figlet.LoadBindataFont(fontFuture, "future")
	figletOpts.FontName = "future"
	return Common{
		Store:      data.NewStore(),
		Figlet:     figlet,
		FigletOpts: figletOpts,
	}
}

// SetSize sets the width and height of the common struct
func (c *Common) SetSize(height, width int) {
	c.Width = width
	c.Height = height
}

// Commands providing an interface between the tui and the data layer

type ErrMsg struct{ Error error }

type WhyDataMsg struct {
	Data  []data.Why
	Error error
}

func (c *Common) ReadWhys(filter data.WhyStatusEnum) tea.Cmd {
	return func() tea.Msg {
		res, err := c.Store.GetWhys(filter)
		sort.Slice(res, func(i, j int) bool {
			return res[i].Number < res[j].Number
		})
		return WhyDataMsg{
			Data:  res,
			Error: err,
		}
	}
}

func (c *Common) UpsertWhys(whys []data.Why) tea.Cmd {
	return func() tea.Msg {
		err := c.Store.UpsertWhys(whys)
		return ErrMsg{err}
	}
}

func (c *Common) DeleteWhys(whys []data.Why) tea.Cmd {
	return func() tea.Msg {
		err := c.Store.DeleteWhys(whys)
		return ErrMsg{err}
	}
}

type IntentionMsg struct {
	Yesterday []data.Intention
	Today     []data.Intention
	Tomorrow  []data.Intention
	Error     error
}

func (c *Common) UpsertIntentions(intentions []data.Intention) tea.Cmd {
	return func() tea.Msg {
		err := c.Store.UpsertIntentions(intentions)
		return ErrMsg{err}
	}
}

func (c *Common) GetDaysIntentions(day time.Time) tea.Cmd {
	var days [3]time.Time
	days[0] = day.AddDate(0, 0, -1)
	days[1] = day
	days[2] = day.AddDate(0, 0, 1)
	var res [3][]data.Intention
	return func() tea.Msg {
		for i := range days {
			data, err := c.Store.GetDaysIntentions(days[i])
			if err != nil {
				return IntentionMsg{Error: err}
			}
			sort.Slice(data, func(i, j int) bool {
				return data[i].Position < data[j].Position
			})
			res[i] = data
		}
		return IntentionMsg{
			Yesterday: res[0],
			Today:     res[1],
			Tomorrow:  res[2],
			Error:     nil,
		}
	}
}

func (c *Common) UpsertDayReview(days []data.Day) tea.Cmd {
	return func() tea.Msg {
		err := c.Store.UpsertDayReview(days)
		return ErrMsg{err}
	}
}
