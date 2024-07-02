package assembler

import (
	"fmt"
	"strconv"
)

type Assembler struct {
	symbolTable map[string]byte
}

func NewAssembler() *Assembler {
	return &Assembler{
		symbolTable: make(map[string]byte),
	}
}

func (a *Assembler) Assemble(instructions []Instruction) ([]byte, error) {
	var machineCode []byte

	// First pass: build symbol table
	for address, instruction := range instructions {
		if instruction.Opcode == "LABEL" {
			a.symbolTable[instruction.Operand] = byte(address)
		}
	}

	// Second pass: generate machine code
	for _, instruction := range instructions {
		if instruction.Opcode == "LABEL" {
			continue
		}

		opcode, err := a.encodeOpcode(instruction.Opcode)
		if err != nil {
			return nil, err
		}

		operand, err := a.encodeOperand(instruction.Operand)
		if err != nil {
			return nil, err
		}

		machineCode = append(machineCode, opcode|operand)
	}

	return machineCode, nil
}

func (a *Assembler) encodeOpcode(opcode string) (byte, error) {
	switch opcode {
	case "ADD":
		return 0x00, nil
	case "SUB":
		return 0x10, nil
	case "MUL":
		return 0x20, nil
	case "DIV":
		return 0x30, nil
	case "AND":
		return 0x40, nil
	case "OR":
		return 0x50, nil
	case "XOR":
		return 0x60, nil
	case "LOAD":
		return 0x70, nil
	case "STORE":
		return 0x80, nil
	case "JUMP":
		return 0x90, nil
	case "JZ":
		return 0xA0, nil
	case "JNZ":
		return 0xB0, nil
	case "IN":
		return 0xC0, nil
	case "OUT":
		return 0xD0, nil
	case "HALT":
		return 0xE0, nil
	default:
		return 0, fmt.Errorf("unknown opcode: %s", opcode)
	}
}

func (a *Assembler) encodeOperand(operand string) (byte, error) {
	if operand == "" {
		return 0, nil
	}

	// Check if it's a label
	if value, ok := a.symbolTable[operand]; ok {
		return value, nil
	}

	// Try to parse as a number
	value, err := strconv.ParseUint(operand, 0, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid operand: %s", operand)
	}

	return byte(value), nil
}
