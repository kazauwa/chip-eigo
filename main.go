package main

import (
	"fmt"
	"os"

	emulator "github.com/kazauwa/chip-eigo/emulator"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("This a chip-8 emulator stub.")

	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)

	chip8 := emulator.Init()
	chip8.ReadProgram("./roms/ibm_logo.ch8")
	for i := 0; i < 5; i++ {
		chip8.Run()
	}
}
