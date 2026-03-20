package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/compiler"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/lexer"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/atlaspl/parser"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/network"
	"github.com/HMZElidrissi/atlas-virtual-machine/internal/vm"
	pb "github.com/HMZElidrissi/atlas-virtual-machine/proto"
)

const helpText = `AtlasVM — a distributed virtual machine for AtlasPL programs.

Usage:
  atlasvm [flags] <program.atlas>
  atlasvm [flags]              (reads from stdin)

Example:
  atlasvm examples/even_odd.atlas
  atlasvm --local examples/sum.atlas

Flags:
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, helpText)
		flag.PrintDefaults()
	}
	localOnly := flag.Bool("local", false, "skip distributed consensus and just print the VM output")
	flag.Parse()

	// ─── 1. Read source from file or stdin ────────────────────────────────────
	var src []byte
	var err error

	switch flag.NArg() {
	case 0:
		log.Println("No file given — reading from stdin (Ctrl-D when done)...")
		src, err = os.ReadFile("/dev/stdin")
	case 1:
		src, err = os.ReadFile(flag.Arg(0))
		if err == nil {
			log.Printf("Running %s", flag.Arg(0))
		}
	default:
		flag.Usage()
		os.Exit(1)
	}
	if err != nil {
		log.Fatalf("Could not read source: %v", err)
	}

	// ─── 2. Lex + Parse ───────────────────────────────────────────────────────
	l := lexer.NewLexer(bytes.NewReader(src))
	p := parser.NewParser(l)
	program := p.ParseProgram()

	if errs := p.Errors(); len(errs) != 0 {
		fmt.Fprintln(os.Stderr, "Parse errors:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e)
		}
		os.Exit(1)
	}

	// ─── 3. Compile AST → bytecode ────────────────────────────────────────────
	c := compiler.NewCompiler()
	compiled, err := c.Compile(program)
	if err != nil {
		log.Fatalf("Compilation failed: %v", err)
	}
	log.Printf("Compiled %d instruction bytes", len(compiled.Bytecode))
	for i, b := range compiled.Bytecode {
		log.Printf("  [%02d] 0x%02X", i, b)
	}

	// ─── 4. Load + run on VM 1 ────────────────────────────────────────────────
	vm1 := vm.NewVM(os.Stdin, os.Stdout)
	if err := vm1.LoadProgram(compiled.Bytecode); err != nil {
		log.Fatalf("LoadProgram: %v", err)
	}
	vm1.LoadData(compiled.InitialData)

	log.Println("Running VM...")
	vm1.Run()
	log.Printf("VM finished: PC=%d ACC=%d", vm1.Registers.PC, vm1.Registers.ACC)

	if *localOnly {
		return
	}

	// ─── 5. Distribute across 3 nodes with PBFT consensus ────────────────────
	vm2 := vm.NewVM(os.Stdin, os.Stdout)
	vm3 := vm.NewVM(os.Stdin, os.Stdout)

	node1 := network.NewNode("node1", "localhost:50051", vm1)
	node2 := network.NewNode("node2", "localhost:50052", vm2)
	node3 := network.NewNode("node3", "localhost:50053", vm3)

	go func() { _ = node1.Start() }()
	go func() { _ = node2.Start() }()
	go func() { _ = node3.Start() }()
	time.Sleep(time.Second)

	mustConnect(node1, "node2", "localhost:50052")
	mustConnect(node1, "node3", "localhost:50053")
	mustConnect(node2, "node1", "localhost:50051")
	mustConnect(node2, "node3", "localhost:50053")
	mustConnect(node3, "node1", "localhost:50051")
	mustConnect(node3, "node2", "localhost:50052")

	c1 := network.NewConsensus(node1)
	c2 := network.NewConsensus(node2)
	c3 := network.NewConsensus(node3)
	node1.SetConsensus(c1)
	node2.SetConsensus(c2)
	node3.SetConsensus(c3)

	vmState := &pb.VMState{
		Memory: vm1.Memory.Data[:],
		Pc:     uint32(vm1.Registers.PC),
		Acc:    int32(vm1.Registers.ACC),
	}
	if err := c1.StartConsensus(&pb.ConsensusMessage{
		Type:  pb.ConsensusMessage_PRE_PREPARE,
		View:  1,
		State: vmState,
	}); err != nil {
		log.Fatalf("Consensus start failed: %v", err)
	}

	deadline := time.Now().Add(20 * time.Second)
	var result *pb.ConsensusMessage
	for time.Now().Before(deadline) {
		if result = c1.GetDecidedValue(); result != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if result == nil {
		log.Fatalf("Consensus timed out after 20s")
	}

	log.Printf("Consensus reached. Final VM state: PC=%d, ACC=%d",
		result.State.Pc, result.State.Acc)

	vm1.UpdateState(result.State)
	vm2.UpdateState(result.State)
	vm3.UpdateState(result.State)
	log.Println("All VMs updated to consensus state.")
}

func mustConnect(n *network.Node, id, addr string) {
	if err := n.ConnectToPeer(id, addr); err != nil {
		log.Fatalf("%s → %s: %v", n.ID, id, err)
	}
}
