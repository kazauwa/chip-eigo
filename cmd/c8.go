package main

import (
	"math/rand"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sirupsen/logrus"

	"github.com/kazauwa/chip-eigo/emulator"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
	// zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	rand.Seed(time.Now().UnixNano())

	chip8 := emulator.NewEmulator()
	if err := chip8.ReadProgram("./roms/ibm_logo.ch8"); err != nil {
		logrus.Fatalf("Error reading from file: %v", err)
	}

	program := tea.NewProgram(chip8)
	if _, err := program.Run(); err != nil {
		logrus.Fatalf("Error running chip-8: %v", err)
	}

	// for i := 0; i < 50; i++ {
	// 	chip8.Run()
	// }
}
