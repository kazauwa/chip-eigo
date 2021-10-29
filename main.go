package main

import (
	"fmt"
	"os"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

type Chip8 struct {
	memory         [4096]byte
	stack          [16]uint16
	data_registers [16]byte
	PC             uint16 // program counter (basically, instruction pointer)
	I              uint16 // index register
	SP             byte   // stack pointer
	DT             byte   // delay timer
	ST             byte   // sound timer
	screen         [64][32]bool
}

func (chip8 *Chip8) loadFont() {
	fontOffset := 0x050
	for offset, element := range font {
		chip8.memory[fontOffset+offset] = element
	}
}

func (chip8 *Chip8) run() {
	rawInstruction := chip8.fetch()
	chip8.decode(rawInstruction)
	chip8.execute()
}

func (chip8 *Chip8) fetch() []byte {
	var instructionLen uint16 = 2
	nextInstruction := chip8.memory[chip8.PC : chip8.PC+instructionLen]
	chip8.PC += instructionLen
	return nextInstruction
}

func (chip8 *Chip8) decode(rawInstruction []byte) {
	highBits, lowBits := rawInstruction[0], rawInstruction[1]
	op := (highBits & 0xF0)
	x := (highBits & 0xF)
	y := (lowBits & 0xF0)
	n := (lowBits & 0xF)
	nn := lowBits
	fmt.Printf("raw: %x\n", rawInstruction[0])
	fmt.Printf("op: %x\n", op)
	fmt.Printf("x: %x\n", x)
	fmt.Printf("y: %x\n", y)
	fmt.Printf("n: %x\n", n)
	fmt.Printf("nn: %x\n", nn)
	switch rawInstruction[0] {
	case 0x00:
		switch rawInstruction[1] {
		case 0xE0:
			fmt.Println("clear")
		case 0xEE:
			fmt.Println("return")
		}
	}
}

func (chip8 *Chip8) execute() {}

func (chip8 *Chip8) readProgram(filePath string) {
	fileStream, err := os.Open(filePath)
	check(err)
	defer fileStream.Close()

	buffer := chip8.memory[chip8.PC:]
	_, err = fileStream.Read(buffer)
	check(err)

	// fmt.Println(chip8.memory[programMemoryOffset : programMemoryOffset+bytesRead])

	// for _, b := range buffer {
	// 	fmt.Println(b)
	// }
}

func main() {
	fmt.Println("This a chip-8 emulator stub.")
	chip8 := Chip8{}
	chip8.loadFont()
	chip8.PC = 0x200 // Start of most chip-8 programs
	chip8.readProgram("./roms/ibm_logo.ch8")
	for i := 0; i < 5; i++ {
		chip8.run()
	}
}
