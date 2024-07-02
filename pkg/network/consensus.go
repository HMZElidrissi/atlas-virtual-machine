package network

import (
	"fmt"
	"log"
	"sync"

	pb "github.com/HMZElidrissi/atlas-virtual-machine/proto"
)

type ConsensusState int

const (
	PrePrepare ConsensusState = iota
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

func NewConsensus(node *Node) *Consensus {
	return &Consensus{
		node:           node,
		state:          PrePrepare,
		currentView:    0,
		prepareCount:   make(map[string]int),
		commitCount:    make(map[string]int),
		decisionQuorum: (len(node.Peers) * 2 / 3) + 1,
	}
}

func (c *Consensus) StartConsensus(msg *pb.ConsensusMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != PrePrepare {
		return fmt.Errorf("consensus already in progress")
	}

	c.state = Prepare
	c.currentView++
	c.prepareCount = make(map[string]int)
	c.commitCount = make(map[string]int)

	msg.View = c.currentView
	msg.Type = pb.ConsensusMessage_PRE_PREPARE

	return c.node.Broadcast(msg)
}

func (c *Consensus) HandlePrePrepare(msg *pb.ConsensusMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != PrePrepare {
		return fmt.Errorf("incorrect state for PrePrepare")
	}

	c.state = Prepare
	c.currentView = msg.View

	prepareMsg := &pb.ConsensusMessage{
		Type:  pb.ConsensusMessage_PREPARE,
		View:  c.currentView,
		State: msg.State,
	}

	return c.node.Broadcast(prepareMsg)
}

func (c *Consensus) HandlePrepare(msg *pb.ConsensusMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Prepare {
		return fmt.Errorf("incorrect state for Prepare")
	}

	valueHash := fmt.Sprintf("%v", msg.State)
	c.prepareCount[valueHash]++

	if c.prepareCount[valueHash] >= c.decisionQuorum {
		c.state = Commit
		commitMsg := &pb.ConsensusMessage{
			Type:  pb.ConsensusMessage_COMMIT,
			View:  c.currentView,
			State: msg.State,
		}
		return c.node.Broadcast(commitMsg)
	}

	return nil
}

func (c *Consensus) HandleCommit(msg *pb.ConsensusMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Commit {
		return fmt.Errorf("incorrect state for Commit")
	}

	valueHash := fmt.Sprintf("%v", msg.State)
	c.commitCount[valueHash]++

	if c.commitCount[valueHash] >= c.decisionQuorum {
		c.state = Finalize
		c.decidedValue = msg
		log.Printf("Consensus reached: %v", c.decidedValue)
		// Update the node's VM state
		c.node.VM.UpdateState(c.decidedValue.State)
		return nil
	}

	return nil
}

func (c *Consensus) GetDecidedValue() *pb.ConsensusMessage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.decidedValue
}
