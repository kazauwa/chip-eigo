package emulator

import (
	"math"
	"math/rand"
	"os"

	"github.com/sirupsen/logrus"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

const (
	fontOffset             = 0x050
	defaultFontSpriteWidth = 5
)

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

type chip8Display struct {
	enableDrawWrap bool
	screen         [][]byte
}

func (display *chip8Display) width() int {
	return len(display.screen)
}

func (display *chip8Display) height() int {
	return len(display.screen[0])
}

type Chip8 struct {
	memory        [4096]byte
	stack         [16]uint16
	dataRegisters [16]byte
	PC            uint16 // program counter (i.e. instruction pointer)
	I             uint16 // index register
	SP            byte   // stack pointer
	DT            byte   // delay timer
	ST            byte   // sound timer
	display       *chip8Display
}

func (chip8 *Chip8) loadFont() {
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
			for screenRow := range chip8.display.screen {
				for pixelPosition := range chip8.display.screen[screenRow] {
					chip8.display.screen[screenRow][pixelPosition] = 0
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
		if chip8.dataRegisters[x] == nn {
			chip8.PC += 2
		}
	case 0x4:
		logrus.Debugf("SNE Vx, byte")
		if chip8.dataRegisters[x] != nn {
			chip8.PC += 2
		}
	case 0x5:
		logrus.Debugf("SE Vx, Vy")
		if chip8.dataRegisters[x] == chip8.dataRegisters[y] {
			chip8.PC += 2
		}
	case 0x6:
		logrus.Debugf("LD Vx, byte")
		chip8.dataRegisters[x] = nn
	case 0x7:
		logrus.Debugf("ADD Vx, byte")
		chip8.dataRegisters[x] += nn
	case 0x8:
		switch n {
		case 0x0:
			logrus.Debugf("LD Vx, Vy")
			chip8.dataRegisters[x] = chip8.dataRegisters[y]
		case 0x1:
			logrus.Debugf("OR Vx, Vy")
			chip8.dataRegisters[x] = chip8.dataRegisters[x] | chip8.dataRegisters[y]
		case 0x2:
			logrus.Debugf("AND Vx, Vy")
			chip8.dataRegisters[x] = chip8.dataRegisters[x] & chip8.dataRegisters[y]
		case 0x3:
			logrus.Debugf("XOR Vx, Vy")
			chip8.dataRegisters[x] = chip8.dataRegisters[x] ^ chip8.dataRegisters[y]
		case 0x4:
			logrus.Debugf("ADD Vx, Vy")
			sum := uint16(chip8.dataRegisters[x]) + uint16(chip8.dataRegisters[y])
			chip8.dataRegisters[0xF] = 0
			if sum > math.MaxUint8 {
				chip8.dataRegisters[0xF] = 1
			}
			chip8.dataRegisters[x] = byte(sum)
		case 0x5:
			logrus.Debugf("SUB Vx, Vy")
			chip8.dataRegisters[0xF] = 1
			if chip8.dataRegisters[y] > chip8.dataRegisters[x] {
				chip8.dataRegisters[0xF] = 0
			}
			chip8.dataRegisters[x] = chip8.dataRegisters[x] - chip8.dataRegisters[y]
		case 0x6:
			logrus.Debugf("SHR Vx{, Vy}")
			// TODO: skip assigment if we use CHIP-48 or SUPER-CHIP implementations
			// if chip8.isChip48 == true || chip8.isSuperChip == true
			chip8.dataRegisters[x] = chip8.dataRegisters[y]

			chip8.dataRegisters[0xF] = chip8.dataRegisters[x] & 0b0000_0001
			chip8.dataRegisters[x] >>= 1
		case 0x7:
			logrus.Debugf("SUB Vy, Vx")
			chip8.dataRegisters[0xF] = 1
			if chip8.dataRegisters[x] > chip8.dataRegisters[y] {
				chip8.dataRegisters[0xF] = 0
			}
			chip8.dataRegisters[x] = chip8.dataRegisters[y] - chip8.dataRegisters[x]
		case 0xE:
			logrus.Debugf("SHL Vx{, Vy}")
			// TODO: skip assigment if we use CHIP-48 or SUPER-CHIP implementations
			// if chip8.isChip48 == true || chip8.isSuperChip == true
			chip8.dataRegisters[x] = chip8.dataRegisters[y]

			chip8.dataRegisters[0xF] = chip8.dataRegisters[x] & 0b1000_0000
			chip8.dataRegisters[x] <<= 1
		}
	case 0x9:
		logrus.Debugf("SNE Vx, Vy")
		if chip8.dataRegisters[x] != chip8.dataRegisters[y] {
			chip8.PC += 2
		}
	case 0xA:
		logrus.Debugf("Set I to nnn")
		chip8.I = nnn
	case 0xB:
		logrus.Debugf("Jump with offset")
		// TODO: toggle behaviour for CHIP-48 and SUPER-CHIP
		// chip8.PC = nnn + uint16(chip8.dataRegisters[x])
		chip8.PC = nnn + uint16(chip8.dataRegisters[0x0]) // original behaviour
	case 0xC:
		logrus.Debugf("Generate random number")
		chip8.dataRegisters[x] = byte(rand.Intn(256)) & nn
	case 0xD:
		logrus.Debugf("Display")
		xCoord := chip8.dataRegisters[x] % byte(chip8.display.width())
		yCoord := chip8.dataRegisters[y] % byte(chip8.display.height())
		chip8.dataRegisters[0xF] = 0

		for spriteRowOffset := 0; byte(spriteRowOffset) < n; spriteRowOffset++ {
			spriteRow := chip8.memory[chip8.I+uint16(spriteRowOffset)]
			yCoord = (yCoord + byte(spriteRowOffset)) % byte(chip8.display.height())
			if (int(yCoord)+spriteRowOffset) > chip8.display.height() && !chip8.display.enableDrawWrap {
				break
			}
			for pixel := 0; pixel < math.MaxUint8; pixel++ {
				xCoord = (xCoord + byte(pixel)) % byte(chip8.display.width())
				if (int(xCoord)+pixel) > chip8.display.width() && !chip8.display.enableDrawWrap {
					break
				}
				nBitMask := math.Pow(2, float64(pixel))
				spriteRowNBit := spriteRow & byte(nBitMask)
				if chip8.dataRegisters[0xF] == 0 && chip8.display.screen[yCoord][xCoord+byte(pixel)] == 1 && spriteRowNBit == 1 {
					chip8.dataRegisters[0xF] = 1
				}
				chip8.display.screen[yCoord][xCoord+byte(pixel)] ^= spriteRowNBit
			}
		}
	case 0xE:
		switch nn {
		case 0x9E:
			logrus.Debugf("SKP Vx")
			// TODO: Skip next op if key in Vx is pressed down
			// if isKeyDown(chip8.dataRegisters[x]) {
			// 	chip8.PC += 2
			// }
		case 0xA1:
			logrus.Debugf("SKNP Vx")
			// TODO: Skip next op if key in Vx is NOT pressed down
			// if !isKeyDown(chip8.dataRegisters[x]) {
			// 	chip8.PC += 2
			// }
		}
	case 0xF:
		switch nn {
		case 0x07:
			logrus.Debugf("LD Vx, DT")
			chip8.dataRegisters[x] = chip8.DT
		case 0x0A:
			logrus.Debugf("LD Vx, K")
			// TODO: block until key is pressed
			// On the original COSMAC VIP, the key was only registered when it was pressed and then released.
			// keyDown := getKeyDown()
			// if keyDown == nil {
			// 	chip8.PC -= 2
			// }
			// chip8.dataRegisters[x] = keyDown
		case 0x15:
			logrus.Debugf("LD DT, Vx")
			chip8.DT = chip8.dataRegisters[x]
		case 0x18:
			logrus.Debugf("LD ST, Vx")
			chip8.ST = chip8.dataRegisters[x]
		case 0x1E:
			logrus.Debugf("ADD I, Vx")
			chip8.I += uint16(chip8.dataRegisters[x])
			// set overflow if I value exceeds normal adressing range (0xFFF)
			if chip8.I > 4096 { // TODO: replace with named constant and add toggle to switch this behaviour
				chip8.dataRegisters[0xF] = 1
			}
		case 0x29:
			logrus.Debugf("LD F, Vx")
			chip8.I = fontOffset + uint16(chip8.dataRegisters[x])*defaultFontSpriteWidth
		case 0x33:
			logrus.Debugf("LD B, Vx")
			chip8.memory[chip8.I] = chip8.dataRegisters[x] / 100
			chip8.memory[chip8.I+1] = chip8.dataRegisters[x] % 100 / 10
			chip8.memory[chip8.I+2] = chip8.dataRegisters[x] % 100 % 10
		case 0x55:
			logrus.Debugf("LD [I], Vx")
			for register := 0; register < int(x); register++ {
				chip8.memory[chip8.I+uint16(register)] = chip8.dataRegisters[register]
			}
			// TODO: CHIP-48 and SUPER-CHIP skip setting the I register
			chip8.I += uint16(x + 1)
		case 0x65:
			logrus.Debugf("LD Vx, [I]")
			for register := 0; register < int(x); register++ {
				chip8.dataRegisters[register] = chip8.memory[chip8.I+uint16(register)]
			}
			// TODO: CHIP-48 and SUPER-CHIP skip setting the I register
			chip8.I += uint16(x + 1)
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
}

func newDisplay(height int, width int) *chip8Display {
	displayScreen := make([][]byte, height)
	for row := range displayScreen {
		displayScreen[row] = make([]byte, width)
	}
	return &chip8Display{
		enableDrawWrap: false,
		screen:         displayScreen,
	}
}

func NewEmulator() *Chip8 {
	chip8 := Chip8{
		PC:      0x200, // Start of most chip-8 programs
		display: newDisplay(32, 64),
	}
	chip8.loadFont()
	return &chip8
}
