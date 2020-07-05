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

type Chip8 struct {
	registers [16]byte
	memory    [4096]byte
	index     uint16
	pc        uint16
	opcode    uint16
	stack     [16]uint16
	sp        byte
	gfx       [2048]byte
	dt        uint8
	st        uint8
	keypad    [16]byte
}

func main() {

	chip8 := Chip8{}
	chip8.pc = StartAddress
	chip8.loadGame("c8games/PONG")
	chip8.loadFontSet()

	for i := 0; i < 100; i++ {
		chip8.emulateCycle()
	}

	// DEBUGGING
	// fmt.Println(chip8.memory)
	// fmt.Printf("Starting Address: %x \n", chip8.pc)
	// fmt.Println("Array Length (bits): ", len(chip8.registers))
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

	ch.pc += 2
}

func (ch *Chip8) processOpcode(opCode uint16) {
	fmt.Printf("Opcode: %#x \n", opCode)

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
		register := uint16(ch.registers[(opCode&0xF00)>>8])
		value := opCode & 0x00FF
		if register == value {
			fmt.Println("skip instruction")
			ch.pc += 4
		} else {
			ch.pc += 2
		}
	case 0x4000:
		register := uint16(ch.registers[(opCode&0xF00)>>8])
		value := opCode & 0x00FF
		if register != value {
			ch.pc += 4
		} else {
			ch.pc += 2
		}
	case 0x5000:
		registerX := uint16(ch.registers[(opCode&0x0F00)>>8])
		registerY := uint16(ch.registers[(opCode&0x00F0)>>8])
		if registerX == registerY {
			ch.pc += 4
		} else {
			ch.pc += 2
		}
	case 0x6000:
		register := (opCode & 0x0F00) >> 8
		value := byte(opCode & 0x00FF)
		ch.registers[register] = value
		ch.pc += 2
	case 0x7000:
		ch.registers[(opCode&0x0F00)>>8] = ch.registers[(opCode&0x0F00)>>8] + byte(opCode&0x00FF)
		ch.pc += 2
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
