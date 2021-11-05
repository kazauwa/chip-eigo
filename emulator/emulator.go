package emulator

import (
	"math"
	"os"

	"github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var font = [...]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

type Chip8 struct {
	memory         [4096]byte
	stack          [16]uint16
	data_registers [16]byte
	PC             uint16 // program counter (i.e. instruction pointer)
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

func (chip8 *Chip8) Run() {
	rawInstruction := chip8.fetch()
	chip8.decode(rawInstruction)
	chip8.execute()
}

func (chip8 *Chip8) fetch() uint16 {
	var instructionLen uint16 = 2
	highBits, lowBits := chip8.memory[chip8.PC], chip8.memory[chip8.PC+instructionLen]
	nextOp := (uint16(highBits)<<8 | uint16(lowBits))
	chip8.PC += instructionLen
	return nextOp
}

func (chip8 *Chip8) decode(rawInstruction uint16) {
	op := byte(rawInstruction >> 12)
	x := byte(rawInstruction >> 8 & 0xF)
	y := (byte(rawInstruction) & 0xF0)
	n := (byte(rawInstruction) & 0xF)
	nn := byte(rawInstruction)
	nnn := rawInstruction & 0xFFF
	logrus.Debugf("raw: %x, %x", byte(rawInstruction>>8), byte(rawInstruction))
	logrus.Debugf("op: %x", op)
	logrus.Debugf("x: %x", x)
	logrus.Debugf("y: %x", y)
	logrus.Debugf("n: %x", n)
	logrus.Debugf("nn: %x", nn)
	logrus.Debugf("nnn: %x", nnn)
	switch op {
	case 0x0:
		switch nn {
		case 0xE0:
			logrus.Debugf("clear screen")
			for screenRow := range chip8.screen {
				for pixelPosition := range chip8.screen[screenRow] {
					chip8.screen[screenRow][pixelPosition] = false
				}
			}
		case 0xEE:
			logrus.Debugf("return from subroutine")
			if chip8.SP == 0 {
				logrus.Fatal("Stack is empty!")
				panic("Stack is empty!")
			}
			chip8.PC = chip8.stack[chip8.SP]
			chip8.stack[chip8.SP] = 0
			chip8.SP -= 1
		}
	case 0x1:
		logrus.Debugf("Jump to location nnn")
		chip8.PC = nnn
	case 0x2:
		logrus.Debugf("Call subroutine")
		if int(chip8.SP) == len(chip8.stack) {
			logrus.Fatal("Stack is full!")
			panic("Stack is full!")
		}
		chip8.SP += 1
		chip8.stack[chip8.SP] = chip8.PC
		chip8.PC = nnn
	case 0x3:
		logrus.Debugf("SE Vx, byte")
		if chip8.data_registers[x] == nn {
			chip8.PC += 2
		}
	case 0x4:
		logrus.Debugf("SNE Vx, byte")
		if chip8.data_registers[x] != nn {
			chip8.PC += 2
		}
	case 0x5:
		logrus.Debugf("SE Vx, Vy")
		if chip8.data_registers[x] == chip8.data_registers[y] {
			chip8.PC += 2
		}
	case 0x6:
		logrus.Debugf("LD Vx, byte")
		chip8.data_registers[x] = nn
	case 0x7:
		logrus.Debugf("ADD Vx, byte")
		chip8.data_registers[x] += nn
	case 0x8:
		switch n {
		case 0x0:
			logrus.Debugf("LD Vx, Vy")
			chip8.data_registers[x] = chip8.data_registers[y]
		case 0x1:
			logrus.Debugf("OR Vx, Vy")
			chip8.data_registers[x] = chip8.data_registers[x] | chip8.data_registers[y]
		case 0x2:
			logrus.Debugf("AND Vx, Vy")
			chip8.data_registers[x] = chip8.data_registers[x] & chip8.data_registers[y]
		case 0x3:
			logrus.Debugf("XOR Vx, Vy")
			chip8.data_registers[x] = chip8.data_registers[x] ^ chip8.data_registers[y]
		case 0x4:
			logrus.Debugf("ADD Vx, Vy")
			sum := uint16(chip8.data_registers[x]) + uint16(chip8.data_registers[y])
			if sum > math.MaxUint8 {
				chip8.data_registers[0xF] = 1
			} else {
				chip8.data_registers[0xF] = 0
			}
			chip8.data_registers[x] = byte(sum)
		case 0x5:
			logrus.Debugf("SUB Vx, Vy")
			if chip8.data_registers[y] > chip8.data_registers[x] {
				chip8.data_registers[0xF] = 0
			} else {
				chip8.data_registers[0xF] = 1
			}
			chip8.data_registers[x] = chip8.data_registers[x] - chip8.data_registers[y]
		case 0x6:
			logrus.Debugf("SHR Vx{, Vy}")
			// TODO: skip if we use CHIP-48 or SUPER-CHIP implementations
			// if chip8.isChip48 == true || chip8.isSuperChip == true
			chip8.data_registers[x] = chip8.data_registers[y]
			chip8.data_registers[0xF] = chip8.data_registers[x] & 0b0000_0001
			chip8.data_registers[x] >>= 1
		case 0x7:
			logrus.Debugf("SUB Vy, Vx")
			if chip8.data_registers[x] > chip8.data_registers[y] {
				chip8.data_registers[0xF] = 0
			} else {
				chip8.data_registers[0xF] = 1
			}
			chip8.data_registers[x] = chip8.data_registers[y] - chip8.data_registers[x]
		case 0xE:
			logrus.Debugf("SHL Vx{, Vy}")
			// TODO: skip if we use CHIP-48 or SUPER-CHIP implementations
			// if chip8.isChip48 == true || chip8.isSuperChip == true
			chip8.data_registers[x] = chip8.data_registers[y]
			chip8.data_registers[0xF] = chip8.data_registers[x] & 0b1000_0000
			chip8.data_registers[x] <<= 1
		}
	case 0x9:
		logrus.Debugf("SNE Vx, Vy")
		if chip8.data_registers[x] != chip8.data_registers[y] {
			chip8.PC += 2
		}
	default:
	}
}

func (chip8 *Chip8) execute() {}

func (chip8 *Chip8) ReadProgram(filePath string) {
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

func Init() Chip8 {
	chip8 := Chip8{
		PC: 0x200, // Start of most chip-8 programs
	}
	chip8.loadFont()
	return chip8
}
