package vm

import (
	"fmt"
	"log"
	pb "github.com/HMZElidrissi/atlas-virtual-machine/proto"
	"io"
)


type VM struct {
	Memory    *Memory
	Registers *Registers
	Stack     *Stack
	running   bool
	input     io.Reader
	output    io.Writer
}

func NewVM(input io.Reader, output io.Writer) *VM {
	memory := NewMemory()
	return &VM{
		Memory:    memory,
		Registers: NewRegisters(),
		Stack:     NewStack(memory),
		input:     input,
		output:    output,
	}
}

func (vm *VM) Run() {
	log.Println("Running VM...")
	vm.running = true
	for vm.running {
		// PC is a relative offset within the code segment.
		// Add DataSegmentSize to get the absolute memory address.
		instruction := DecodeInstruction(vm.Memory.Read(DataSegmentSize + uint16(vm.Registers.PC)))
		vm.Registers.PC++
		vm.executeInstruction(instruction)
	}
}

func (vm *VM) executeInstruction(instruction Instruction) {
	switch instruction.Opcode {
	case ADD:
		vm.Registers.ACC += int8(vm.Memory.Read(uint16(instruction.Operand)))
	case SUB:
		vm.Registers.ACC -= int8(vm.Memory.Read(uint16(instruction.Operand)))
	case MUL:
		vm.Registers.ACC *= int8(vm.Memory.Read(uint16(instruction.Operand)))
	case DIV:
		vm.Registers.ACC /= int8(vm.Memory.Read(uint16(instruction.Operand)))
	case AND:
		vm.Registers.ACC &= int8(vm.Memory.Read(uint16(instruction.Operand)))
	case OR:
		vm.Registers.ACC |= int8(vm.Memory.Read(uint16(instruction.Operand)))
	case XOR:
		vm.Registers.ACC ^= int8(vm.Memory.Read(uint16(instruction.Operand)))
	case LOAD:
		vm.Registers.ACC = int8(vm.Memory.Read(uint16(instruction.Operand)))
	case STORE:
		vm.Memory.Write(uint16(instruction.Operand), byte(vm.Registers.ACC))
	case JUMP:
		vm.Registers.PC = instruction.Operand
	case JZ:
		if vm.Registers.ACC == 0 {
			vm.Registers.PC = instruction.Operand
		}
	case JNZ:
		if vm.Registers.ACC != 0 {
			vm.Registers.PC = instruction.Operand
		}
	case IN:
		var input byte
		fmt.Fscan(vm.input, &input)
		vm.Registers.ACC = int8(input)
	case OUT:
		fmt.Fprintf(vm.output, "%d\n", vm.Registers.ACC)
	case HALT:
		vm.running = false
	default:
		panic(fmt.Sprintf("Unknown opcode: %d", instruction.Opcode))
	}
}

func (vm *VM) UpdateState(state *pb.VMState) {
	// Copy the state memory into VM memory
	copy(vm.Memory.Data[:], state.Memory)
	vm.Registers.PC = uint8(state.Pc)
	vm.Registers.ACC = int8(state.Acc)
}

// LoadProgram copies bytecode into the code segment and resets the CPU state.
// PC is stored as a zero-based offset within the code segment (0 = address 512).
func (vm *VM) LoadProgram(program []byte) error {
	if len(program) > CodeSegmentSize {
		return fmt.Errorf("program size (%d bytes) exceeds code segment size (%d bytes)", len(program), CodeSegmentSize)
	}

	// Clear the existing code segment
	for i := DataSegmentSize; i < MemorySize; i++ {
		vm.Memory.Data[i] = 0
	}

	// Load the program into the code segment
	copy(vm.Memory.Data[DataSegmentSize:], program)

	// PC = 0 means the first instruction at absolute address DataSegmentSize.
	vm.Registers.PC = 0
	vm.Registers.ACC = 0

	return nil
}

// LoadData pre-populates data-segment addresses with initial values.
// Used by the compiler to seed constant pools and pre-initialized variables.
func (vm *VM) LoadData(data map[uint8]byte) {
	for addr, val := range data {
		vm.Memory.Data[addr] = val
	}
}

// Running reports whether the VM is currently executing.
func (vm *VM) Running() bool {
	return vm.running
}
