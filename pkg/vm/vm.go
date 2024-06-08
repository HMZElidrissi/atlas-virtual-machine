package vm

import "fmt"

const MemorySize = 1024

type VM struct {
	Memory    [MemorySize]byte
	PC        byte
	ACC       byte
	DataStack *Stack
}

func NewVM() *VM {
	return &VM{
		PC:        0,
		ACC:       0,
		DataStack: NewStack(),
	}
}

func (vm *VM) Run() {
	// TODO: Implement VM execution
	fmt.Println("Running VM...")
	fmt.Println("   ##   ##### #        ##    ####  ")
	fmt.Println("  #  #    #   #       #  #  #      ")
	fmt.Println("  #  #    #   #       #  #  #      ")
	fmt.Println(" #    #   #   #      #    #  ####  ")
	fmt.Println(" ######   #   #      ######      # ")
	fmt.Println(" #    #   #   #      #    # #    # ")
	fmt.Println(" #    #   #   ###### #    #  ####  ")
}
