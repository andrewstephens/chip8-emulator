package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"
)

const (
	StartAddress        = 0x200
	FontsetStartAddress = 0x50
)

// Chip8 emulates a Chip-8 machine.
type Chip8 struct {
	memory [4096]byte

	// The Chip 8 has 16 registers. The first 15 registers are
	// general purpose registers, with the 16th register being used
	// for the "carry" flag.
	V [16]byte

	// There is one Index register (I)
	I uint16

	// The Program Counter (pc) is used to keep track
	pc     uint16
	opcode uint16
	stack  [16]uint16
	sp     byte

	// The Chip 8 graphics system draws sprites to the screen.
	// The screen is 64 x 32
	gfx      [64 * 32]byte
	drawFlag bool

	delayTimer byte
	soundTimer byte
	keypad     [16]byte
}

func main() {

	chip8 := Chip8{}
	chip8.pc = 0x200 // Start of program memory address
	chip8.I = 0
	chip8.sp = 0
	chip8.opcode = 0

	// Reset / Clear Graphics
	chip8.gfx = [64 * 32]byte{}

	// Clear the registers
	chip8.V = [16]byte{}

	// Clear the memory
	chip8.memory = [4096]byte{}

	// Clear the Stack
	chip8.stack = [16]uint16{}

	// Reset timers
	chip8.delayTimer = 0
	chip8.soundTimer = 0

	// Load Font
	chip8.loadFontSet()

	// Load Game
	chip8.loadGame("c8games/PONG")

	for i := 0; i < 100; i++ {
		chip8.emulateCycle()
	}

	// DEBUGGING
	// fmt.Println(chip8.memory)
	// fmt.Printf("Starting Address: %x \n", chip8.pc)
	// fmt.Println("Array Length (bits): ", len(chip8.V))
}

func (ch *Chip8) loadGame(filename string) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Whoops, couln't read rom")
		os.Exit(1)
	}

	for i, b := range bytes {
		ch.memory[StartAddress+i] = b
	}
}

func (ch *Chip8) emulateCycle() {
	// fetch opcode

	opCode := uint16(ch.memory[ch.pc])<<8 | uint16(ch.memory[ch.pc+1])

	ch.processOpcode(opCode)

	if ch.delayTimer > 0 {
		ch.delayTimer--
	}

	if ch.soundTimer > 0 {
		if ch.soundTimer == 1 {
			fmt.Printf("BEEP!\n")
			ch.soundTimer--
		}
	}
}

func (ch *Chip8) processOpcode(opCode uint16) {
	fmt.Printf("Started Executing Opcode: %#x\n", opCode)
	switch opCode & 0xF000 {
	case 0x0000:
		switch opCode & 0x000F {
		case 0x0000:
			for i := range ch.gfx {
				ch.gfx[i] = 0
			}
			ch.pc += 2
		case 0x000E:
			ch.sp--
			ch.pc = ch.stack[ch.sp]
			ch.pc += 2
		default:
			fmt.Printf("Invalid opcode %X\n", opCode)
		}

	case 0x1000:
		ch.pc = opCode & 0x0FFF

	case 0x2000:
		ch.stack[ch.sp] = ch.pc
		ch.sp++
		ch.pc = opCode & 0x0FFF

	case 0x3000:
		register := uint16(ch.V[(opCode&0xF00)>>8])
		value := opCode & 0x00FF
		if register == value {
			ch.pc += 4
		} else {
			ch.pc += 2
		}

	case 0x4000:
		register := uint16(ch.V[(opCode&0xF00)>>8])
		value := opCode & 0x00FF
		if register != value {
			ch.pc += 4
		} else {
			ch.pc += 2
		}

	case 0x5000:
		registerX := uint16(ch.V[(opCode&0x0F00)>>8])
		registerY := uint16(ch.V[(opCode&0x00F0)>>8])
		if registerX == registerY {
			ch.pc += 4
		} else {
			ch.pc += 2
		}

	case 0x6000:
		x := (opCode & 0x0F00) >> 8
		nn := byte(opCode & 0x00FF)
		ch.V[x] = nn
		ch.pc += 2

	case 0x7000:
		x := (opCode & 0x0F00) >> 8
		nn := byte(opCode & 0x00FF)
		ch.V[x] += nn
		ch.pc += 2

	case 0x8000:
		x := (opCode & 0x0F00) >> 8
		y := (opCode & 0x00F0) >> 4
		switch opCode & 0x000F {
		case 0x0000:
			ch.V[x] = ch.V[y]
			ch.pc += 2
		case 0x0001:
			ch.V[x] |= ch.V[y]
			ch.pc += 2
		case 0x0002:
			ch.V[x] &= ch.V[y]
			ch.pc += 2
		case 0x0003:
			ch.V[x] ^= ch.V[y]
			ch.pc += 2
		case 0x0004:
			if (ch.V[x] + ch.V[y]) > 0xFF {
				ch.V[0xF] = 1
			} else {
				ch.V[0xF] = 0
			}

			ch.V[x] += ch.V[y]
			ch.pc += 2
		case 0x0005:
			if ch.V[y] > ch.V[x] {
				ch.V[0xF] = 0
			} else {
				ch.V[0xF] = 1
			}

			ch.V[x] -= ch.V[y]
			ch.pc += 2
		case 0x0006:
			ch.V[x] = ch.V[y] >> 1
			ch.V[0xF] = ch.V[y] & 0x01
			ch.pc += 2
		case 0x0007:
			if ch.V[x] > ch.V[y] {
				ch.V[0xF] = 0
			} else {
				ch.V[0xF] = 1
			}

			ch.V[x] = ch.V[y] - ch.V[x]
			ch.pc += 2
		case 0x000E:
			ch.V[0xF] = (ch.V[x] & 0x80) >> 7
			ch.V[x] = ch.V[x] << 1

		}

	case 0x9000:
		x := (opCode & 0x0F00) >> 8
		y := (opCode & 0x00F0) >> 4

		if ch.V[x] != ch.V[y] {
			ch.pc += 4
		} else {
			ch.pc += 2
		}

	case 0xA000:
		ch.I = opCode & 0x0FFF
		ch.pc += 2

	case 0xB000:
		nnn := opCode & 0x0FFF
		ch.pc = uint16(ch.V[0]) + nnn

	case 0xC000:
		x := uint16(opCode&0x0F00) >> 8
		nn := opCode & 0x00FF
		ch.V[x] = byte(randomByte()) & byte(nn)
		ch.pc += 2

	case 0xD000:
		x := uint16(ch.V[(opCode&0x0F00)>>8])
		y := uint16(ch.V[(opCode&0x00F0)>>4])
		height := opCode & 0x000F
		var pixel uint16

		ch.V[0xF] = 0
		for yline := uint16(0); yline < height; yline++ {
			pixel = uint16(ch.memory[ch.I+yline])
			for xline := uint16(0); xline < 8; xline++ {
				index := (x + xline + ((y + yline) * 64))
				if index > uint16(len(ch.gfx)) {
					continue
				}
				if (pixel & (0x80 >> xline)) != 0 {
					if ch.gfx[index] == 1 {
						ch.V[0xF] = 1
					}
					ch.gfx[index] ^= 1
				}
			}
		}

		ch.drawFlag = true
		ch.pc += 2

	case 0xE000:
		x := (opCode & 0xF00) >> 8
		switch opCode & 0x00FF {
		case 0x009E:
			if ch.keypad[ch.V[x]] != 0 {
				ch.pc += 4
				ch.keypad[ch.V[x]] = 0
			} else {
				ch.pc += 2
			}
		case 0x00A1:
			if ch.keypad[ch.V[x]] == 0 {
				ch.pc += 4
			} else {
				ch.keypad[ch.V[x]] = 0
				ch.pc += 2
			}
		}

	case 0xF000:
		x := (opCode & 0x0F00) >> 8
		switch opCode & 0x00FF {
		case 0x0007:
			ch.V[x] = ch.delayTimer
			ch.pc += 2
		case 0x000A:
			for index, k := range ch.keypad {
				if k != 0 {
					ch.V[x] = byte(index)
					ch.pc += 2
					break
				}
			}
			ch.keypad[ch.V[x]] = 0
		case 0x0015:
			ch.delayTimer = ch.V[x]
			ch.pc += 2
		case 0x0018:
			ch.soundTimer = ch.V[x]
			ch.pc += 2
		case 0x001E:
			ch.I += uint16(ch.V[x])
			ch.pc += 2
		case 0x0029:
			ch.I = uint16(ch.V[x]) * 5
			ch.pc += 2
		case 0x0033:
			ch.memory[ch.I] = ch.V[x] / 100
			ch.memory[ch.I+1] = (ch.V[x] / 10) % 10
			ch.memory[ch.I+2] = (ch.V[x] % 100) % 10
			ch.pc += 2
		case 0x0055:
			for i := uint16(0); i <= x; i++ {
				ch.memory[ch.I+i] = ch.V[i]
			}
			ch.pc += 2
		case 0x0065:
			for i := uint16(0); i <= x; i++ {
				ch.V[i] = ch.memory[ch.I+i]
			}
			ch.pc += 2
		default:
			fmt.Printf("Unknown Opcode 0x%X", opCode)
		}
	}
}

func (ch *Chip8) loadFontSet() {
	fontset := [80]byte{
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

	for i := range fontset {
		ch.memory[FontsetStartAddress+i] = fontset[i]
	}
}

func randomByte() int {
	randSource := rand.NewSource(time.Now().UnixNano())
	random := rand.New(randSource)
	return random.Intn(255)
}
