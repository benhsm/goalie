package colorpicker

import (
	_ "embed"
	"log"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var (
	//go:embed hex_colors.txt
	color_data string

	docStyle   = lipgloss.NewStyle().Padding(1, 1)
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder(), false, false, true, false)
	colorRegex = regexp.MustCompile("^(#[A-Z0-9]{6})|([0-9]{1,3})$")
)

type Model struct {
	Colors    []lipgloss.Color
	selection textinput.Model
	info      string
	done      bool
	Choice    lipgloss.Color
	width     int
	height    int
}

func New() Model {
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
	return Model{
		Colors:    colorList,
		selection: selection,
		info:      "Input either a hex code or an ANSI color code.",
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		width, height, err := term.GetSize(0)
		if err != nil {
			log.Fatalf("error setting terminal size")
		}
		m.width, m.height = width, height
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyEnter:
			if !validColor(m.selection.Value()) {
				m.info = "Sorry, that's not a valid color value."
			} else {
				m.Choice = lipgloss.Color(m.selection.Value())
			}
		}
	}

	m.selection, cmd = m.selection.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
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
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, b.String())
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
