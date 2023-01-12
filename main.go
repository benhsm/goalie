package main

import (
	"fmt"
	"log"
	"os"

	"github.com/benhsm/goals/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Getenv("TEA_DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer f.Close()
	}
	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}
