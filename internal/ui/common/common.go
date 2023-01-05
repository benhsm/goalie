package common

import (
	_ "embed"

	"github.com/benhsm/goals/internal/data"
	tea "github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/mbndr/figlet4go"
)

//go:embed future.tlf
var fontFuture []byte

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
