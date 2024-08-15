package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/HMZElidrissi/atlas-virtual-machine/pkg/atlaspl"
	"github.com/HMZElidrissi/atlas-virtual-machine/pkg/network"
	"github.com/HMZElidrissi/atlas-virtual-machine/pkg/vm"
	pb "github.com/HMZElidrissi/atlas-virtual-machine/proto"
)

func main() {
	// AtlasPL source code
	source := `
    @ This program checks if a number is even.
    var number: int;
    number = 10; @ Assign a value to number
    if ((number & 1) == 0) { @ Check if last bit is 0 (even)
      return (0);
    } else {
      return (1);
    }
    `

	// Create lexer
	l := atlaspl.NewLexer(strings.NewReader(source))

	// Create parser
	p := atlaspl.NewParser(l)

	// Parse program
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		log.Println("Parser errors:")
		for _, err := range p.Errors() {
			log.Println(err)
		}
		return
	}

	// Create interpreter
	interpreter := atlaspl.NewInterpreter()

	// Run the program
	err := interpreter.Interpret(program)
	if err != nil {
		log.Fatalf("Interpretation failed: %v", err)
	}

	// Initialize VMs
	vm1 := vm.NewVM(os.Stdin, os.Stdout)
	vm2 := vm.NewVM(os.Stdin, os.Stdout)
	vm3 := vm.NewVM(os.Stdin, os.Stdout)

	// Initialize network nodes
	node1 := network.NewNode("node1", "localhost:50051", vm1)
	node2 := network.NewNode("node2", "localhost:50052", vm2)
	node3 := network.NewNode("node3", "localhost:50053", vm3)

	go node1.Start()
	go node2.Start()
	go node3.Start()

	time.Sleep(time.Second) // Wait for nodes to start

	// Connect nodes
	node1.ConnectToPeer("node2", "localhost:50052")
	node1.ConnectToPeer("node3", "localhost:50053")
	node2.ConnectToPeer("node1", "localhost:50051")
	node2.ConnectToPeer("node3", "localhost:50053")
	node3.ConnectToPeer("node1", "localhost:50051")
	node3.ConnectToPeer("node2", "localhost:50052")

	// Initialize consensus
	consensus1 := network.NewConsensus(node1)

	// Prepare VM state for consensus
	vmState := &pb.VMState{
		Memory: vm1.Memory.Data[:],
		Pc:     uint32(vm1.Registers.PC),
		Acc:    int32(vm1.Registers.ACC),
	}
	consensusMsg := &pb.ConsensusMessage{
		Type:  pb.ConsensusMessage_PRE_PREPARE,
		View:  1,
		State: vmState,
	}

	// Start consensus on node1
	err = consensus1.StartConsensus(consensusMsg)
	if err != nil {
		log.Fatalf("Failed to start consensus: %v", err)
	}

	// Wait for consensus to be reached
	time.Sleep(5 * time.Second)

	// Check the consensus result
	result := consensus1.GetDecidedValue()
	log.Printf("Consensus reached. Final VM state: PC=%d, ACC=%d", result.State.Pc, result.State.Acc)

	// Update VM states based on consensus result
	vm1.UpdateState(result.State)
	vm2.UpdateState(result.State)
	vm3.UpdateState(result.State)

	log.Println("All VMs updated to consensus state")
}
