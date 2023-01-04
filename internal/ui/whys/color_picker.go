package whys

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/benhsm/goals/internal/ui/common"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	//go:embed hex_colors.txt
	color_data string

	frame      = lipgloss.NewStyle().Padding(1, 1)
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, false, true, false)
	colorRegex = regexp.MustCompile("^(#[A-Z0-9]{6})|([0-9]{1,3})$")
)

// colorPickerModel represents a UI for choosing a hex color
type colorPickerModel struct {
	// Colors is a slice a list of preset colors
	Colors []lipgloss.Color

	// Choice is a color chosen by the user
	Choice    lipgloss.Color
	selection textinput.Model
	info      string
	done      bool
	common.Common
}

// New returns a new color picker model
func newColorPicker() colorPickerModel {
	colorStrings := strings.Split(color_data, "\n")
	var colorList []lipgloss.Color
	for _, color := range colorStrings {
		if color == "" {
			break
		}
		colorList = append(colorList, lipgloss.Color(color))
	}
	selection := textinput.New()
	selection.Placeholder = "choose a color"
	selection.CharLimit = 7
	selection.Width = 13
	selection.Focus()
	return colorPickerModel{
		Colors:    colorList,
		selection: selection,
		info:      "Input either a hex code or an ANSI color code.",
	}
}

func (m colorPickerModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update is a Bubble Tea
func (m colorPickerModel) Update(msg tea.Msg) (colorPickerModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyEnter:
			if !validColor(m.selection.Value()) {
				m.info = "Sorry, that's not a valid color value."
			} else {
				m.Choice = lipgloss.Color(m.selection.Value())
				m.selection.SetValue("")
			}
		}
	}

	m.selection, cmd = m.selection.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m colorPickerModel) View() string {
	var b strings.Builder
	b.WriteString(inputStyle.Render(m.selection.View()))
	b.WriteString("\n")
	b.WriteString(m.info)
	for i, color := range m.Colors {
		if i%7 == 0 {
			b.WriteString("\n\n")
		}
		b.WriteString(string(color))
		b.WriteString(" ")
		b.WriteString(renderColor(color))
		b.WriteString(" ")
	}
	return lipgloss.Place(m.Width, m.Height, lipgloss.Center, lipgloss.Center, b.String())
}

func renderColor(color lipgloss.Color) string {
	return lipgloss.NewStyle().Background(color).Render("  ")
}

func validColor(s string) bool {
	if colorRegex.MatchString(s) && colorRegex.FindString(s) == s {
		return true
	}
	return false
}
