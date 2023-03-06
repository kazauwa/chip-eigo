package emulator

import (
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

type Emulator struct {
	cursor   int
	choices  []string
	selected map[int]struct{}
	busy     bool
	c8       Chip8
}

func (e Emulator) Init() tea.Cmd {
	return nil
}

func (e Emulator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return e, tea.Quit
		case "enter", " ":
			if !e.busy {
				var eg
				wg.
				go e.c8.Run()
			}
		}
	}
	return e, nil
}

func (e Emulator) View() string {
	return "Press 'enter' to start Chip-8"
}
