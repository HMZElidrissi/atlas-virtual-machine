package vm

import "fmt"

const (
	ADD   = 0x00
	SUB   = 0x01
	MUL   = 0x02
	DIV   = 0x03
	AND   = 0x04
	OR    = 0x05
	XOR   = 0x06
	LOAD  = 0x07
	STORE = 0x08
	JUMP  = 0x09
	JZ    = 0x0A
	JNZ   = 0x0B
	IN    = 0x0C
	OUT   = 0x0D
	HALT  = 0x0E
)

func (vm *VM) ExecuteInstruction() {
	opcode := vm.Memory[vm.PC]
	// - Each instruction is 1 byte long.
	// - Opcode occupies the first 4 bits. The remaining bits are used for operand addressing (e.g., memory address) depending on the instruction.
	switch opcode {
	case ADD:
		address := vm.Memory[vm.PC+1]
		vm.ACC += vm.Memory[address]
		vm.PC += 2
	case SUB:
		address := vm.Memory[vm.PC+1]
		vm.ACC -= vm.Memory[address]
		vm.PC += 2
	case MUL:
		address := vm.Memory[vm.PC+1]
		vm.ACC *= vm.Memory[address]
		vm.PC += 2
	case DIV:
		address := vm.Memory[vm.PC+1]
		vm.ACC /= vm.Memory[address]
		vm.PC += 2
	case AND:
		address := vm.Memory[vm.PC+1]
		vm.ACC &= vm.Memory[address]
		vm.PC += 2
	case OR:
		address := vm.Memory[vm.PC+1]
		vm.ACC |= vm.Memory[address]
		vm.PC += 2
	case XOR:
		address := vm.Memory[vm.PC+1]
		vm.ACC ^= vm.Memory[address]
		vm.PC += 2
	case LOAD:
		address := vm.Memory[vm.PC+1]
		vm.ACC = vm.Memory[address]
		vm.PC += 2
	case STORE:
		address := vm.Memory[vm.PC+1]
		vm.Memory[address] = vm.ACC
		vm.PC += 2
	case JUMP:
		address := vm.Memory[vm.PC+1]
		vm.PC = address
	case JZ:
		address := vm.Memory[vm.PC+1]
		if vm.ACC == 0 {
			vm.PC = address
		} else {
			vm.PC += 2
		}
	case JNZ:
		address := vm.Memory[vm.PC+1]
		if vm.ACC != 0 {
			vm.PC = address
		} else {
			vm.PC += 2
		}
	case IN:
		var input byte
		fmt.Print("Enter a value: ")
		fmt.Scanf("%d", &input)
		vm.ACC = input
		vm.PC++
	case OUT:
		fmt.Printf("Output: %d\n", vm.ACC)
		vm.PC++
	case HALT:
		return
	default:
		vm.PC++
	}
}
