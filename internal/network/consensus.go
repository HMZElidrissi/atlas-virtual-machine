package network

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	pb "github.com/HMZElidrissi/atlas-virtual-machine/proto"
)

type ConsensusState int

const (
	INITIAL ConsensusState = iota
	PrePrepare
	Prepare
	Commit
	Finalize
)

type Consensus struct {
	node           *Node
	state          ConsensusState
	currentView    int64
	prepareCount   map[string]int
	commitCount    map[string]int
	mu             sync.Mutex
	decidedValue   *pb.ConsensusMessage
	decisionQuorum int
}

func stateHash(state *pb.VMState) string {
	// Create a compact deterministic key for the VM state.
	// This avoids using fmt on the full `memory` byte slice in quorum maps.
	if state == nil {
		return "nil"
	}

	h := sha256.New()
	_, _ = h.Write(state.Memory)

	var b [8]byte
	binary.LittleEndian.PutUint32(b[0:4], state.Pc)
	h.Write(b[0:4])

	binary.LittleEndian.PutUint32(b[0:4], uint32(state.Acc))
	h.Write(b[0:4])

	return hex.EncodeToString(h.Sum(nil))
}

func NewConsensus(node *Node) *Consensus {
	if node == nil {
		log.Println("Error: Cannot create Consensus with nil Node")
		return nil
	}

	return &Consensus{
		node:           node,
		state:          INITIAL,
		currentView:    0,
		prepareCount:   make(map[string]int),
		commitCount:    make(map[string]int),
		decisionQuorum: (len(node.Peers) * 2 / 3) + 1,
	}
}

func (c *Consensus) StartConsensus(msg *pb.ConsensusMessage) error {
	if c == nil {
		return fmt.Errorf("consensus object is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != INITIAL {
		return fmt.Errorf("consensus already in progress, current state: %v", c.state)
	}

	log.Printf("Starting consensus for view %d", c.currentView+1)
	c.state = PrePrepare
	c.currentView++
	c.prepareCount = make(map[string]int)
	c.commitCount = make(map[string]int)

	msg.View = c.currentView
	msg.Type = pb.ConsensusMessage_PRE_PREPARE

	return c.node.Broadcast(msg)
}

func (c *Consensus) HandlePrePrepare(msg *pb.ConsensusMessage) (*pb.Empty, error) {
	if c == nil {
		log.Println("Error: Consensus object is nil in HandlePrePrepare")
		return nil, fmt.Errorf("consensus object is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Replicas receive PRE_PREPARE when idle (INITIAL); leader is in PrePrepare when it sends.
	if c.state != INITIAL && c.state != PrePrepare {
		return nil, fmt.Errorf("incorrect state for PrePrepare: current state %v", c.state)
	}

	log.Printf("Handling PrePrepare for view %d", msg.View)
	c.state = Prepare
	c.currentView = msg.View

	prepareMsg := &pb.ConsensusMessage{
		Type:  pb.ConsensusMessage_PREPARE,
		View:  c.currentView,
		State: msg.State,
	}

	// Count our own PREPARE so replicas reach quorum (each has 2 peers; we need 2 PREPARE;
	// we send one and receive one from the other replica).
	valueHash := stateHash(msg.State)
	c.prepareCount[valueHash] = 1

	err := c.node.Broadcast(prepareMsg)
	if err != nil {
		return &pb.Empty{}, err
	}
	return &pb.Empty{}, nil
}

func (c *Consensus) HandlePrepare(msg *pb.ConsensusMessage) (*pb.Empty, error) {
	if c == nil {
		log.Println("Error: Consensus object is nil in HandlePrepare")
		return nil, fmt.Errorf("consensus object is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Leader stays in PrePrepare when it receives PREPARE; replicas are in Prepare.
	if c.state != Prepare && c.state != PrePrepare {
		return nil, fmt.Errorf("incorrect state for Prepare: current state %v", c.state)
	}

	log.Printf("Handling Prepare for view %d", msg.View)
	valueHash := stateHash(msg.State)
	c.prepareCount[valueHash]++

	if c.prepareCount[valueHash] >= c.decisionQuorum {
		c.state = Commit
		commitMsg := &pb.ConsensusMessage{
			Type:  pb.ConsensusMessage_COMMIT,
			View:  c.currentView,
			State: msg.State,
		}
		err := c.node.Broadcast(commitMsg)
		return &pb.Empty{}, err
	}

	return &pb.Empty{}, nil
}

func (c *Consensus) HandleCommit(msg *pb.ConsensusMessage) (*pb.Empty, error) {
	if c == nil {
		log.Println("Error: Consensus object is nil in HandleCommit")
		return nil, fmt.Errorf("consensus object is nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Commit {
		return nil, fmt.Errorf("incorrect state for Commit: current state %v", c.state)
	}

	log.Printf("Handling Commit for view %d", msg.View)
	valueHash := stateHash(msg.State)
	c.commitCount[valueHash]++

	if c.commitCount[valueHash] >= c.decisionQuorum {
		c.state = Finalize
		c.decidedValue = msg
		if msg.State != nil {
			log.Printf(
				"Consensus reached: view=%d pc=%d acc=%d",
				msg.View,
				msg.State.Pc,
				msg.State.Acc,
			)
		} else {
			log.Printf("Consensus reached: view=%d", msg.View)
		}
		// Update the node's VM state
		c.node.VM.UpdateState(c.decidedValue.State)
	}

	return &pb.Empty{}, nil
}

func (c *Consensus) GetDecidedValue() *pb.ConsensusMessage {
	if c == nil {
		log.Println("Error: Consensus object is nil in GetDecidedValue")
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.decidedValue
}
