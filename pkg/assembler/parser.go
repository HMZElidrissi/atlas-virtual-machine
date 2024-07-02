package assembler

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Instruction struct {
	Opcode  string
	Operand string
}

type Parser struct {
	scanner *bufio.Scanner
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{
		scanner: bufio.NewScanner(reader),
	}
}

func (p *Parser) Parse() ([]Instruction, error) {
	var instructions []Instruction

	for p.scanner.Scan() {
		line := strings.TrimSpace(p.scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "@") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 1 || len(parts) > 2 {
			return nil, fmt.Errorf("invalid instruction format: %s", line)
		}

		instruction := Instruction{
			Opcode: strings.ToUpper(parts[0]),
		}

		if len(parts) == 2 {
			instruction.Operand = parts[1]
		}

		instructions = append(instructions, instruction)
	}

	if err := p.scanner.Err(); err != nil {
		return nil, err
	}

	return instructions, nil
}
