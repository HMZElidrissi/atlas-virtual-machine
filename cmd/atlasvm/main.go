package main

import (
	"fmt"
	"github.com/HMZElidrissi/atlas-virtual-machine/pkg/vm"
)

func main() {
	fmt.Printf("Welcome to Atlas Virtual Machine\n")
	vm := vm.NewVM()
	vm.Run()
}
