package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/kazauwa/chip-eigo/emulator"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	rand.Seed(time.Now().UnixNano())

	chip8 := emulator.NewEmulator()
	chip8.ReadProgram("./roms/ibm_logo.ch8")
	for i := 0; i < 5; i++ {
		chip8.Run()
	}
}
