package network

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/HMZElidrissi/atlas-virtual-machine/internal/vm"
	pb "github.com/HMZElidrissi/atlas-virtual-machine/proto"
	"google.golang.org/grpc"
)

type NodeService struct {
	pb.UnimplementedNodeServiceServer
	node *Node
}

type Node struct {
	ID        string
	Address   string
	Peers     map[string]*NodeClient
	VM        *vm.VM
	consensus *Consensus
	mu        sync.Mutex
}

type NodeClient struct {
	ID     string
	Client pb.NodeServiceClient
}

func NewNode(id, address string, vm *vm.VM) *Node {
	return &Node{
		ID:      id,
		Address: address,
		Peers:   make(map[string]*NodeClient),
		VM:      vm,
	}
}

// SetConsensus attaches a consensus instance to the node so it can participate in consensus when receiving messages.
func (n *Node) SetConsensus(c *Consensus) {
	n.consensus = c
}

func (n *Node) Start() error {
	lis, err := net.Listen("tcp", n.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodeServiceServer(grpcServer, &NodeService{node: n})

	log.Printf("Node %s starting on %s", n.ID, n.Address)
	return grpcServer.Serve(lis)
}

func (n *Node) ConnectToPeer(id, address string) error {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("failed to connect to peer %s: %v", id, err)
	}

	client := pb.NewNodeServiceClient(conn)
	n.Peers[id] = &NodeClient{ID: id, Client: client}
	return nil
}

func (n *Node) Broadcast(msg *pb.ConsensusMessage) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, peer := range n.Peers {
		_, err := peer.Client.ReceiveMessage(context.Background(), msg)
		if err != nil {
			log.Printf("Failed to send message to peer %s: %v", peer.ID, err)
		}
	}
	return nil
}

func (s *NodeService) ReceiveMessage(ctx context.Context, msg *pb.ConsensusMessage) (*pb.Empty, error) {
	if msg == nil {
		log.Printf("Received nil consensus message (node=%s)", s.node.ID)
		return &pb.Empty{}, nil
	}
	if msg.State != nil {
		log.Printf(
			"Received consensus msg: type=%v view=%d pc=%d acc=%d (node=%s)",
			msg.Type,
			msg.View,
			msg.State.Pc,
			msg.State.Acc,
			s.node.ID,
		)
	} else {
		log.Printf("Received consensus msg: type=%v view=%d (node=%s)", msg.Type, msg.View, s.node.ID)
	}

	if s.node.consensus == nil {
		log.Printf("Node %s has no consensus instance, ignoring message", s.node.ID)
		return &pb.Empty{}, nil
	}

	// Handle consensus asynchronously to avoid deadlock: the sender (e.g. node1) is
	// blocked in Broadcast until this RPC returns. If we call Broadcast from here
	// (e.g. replica broadcasting PREPARE back to node1), node1 cannot accept the
	// new RPC until its Broadcast returns. So process in a goroutine and return immediately.
	go func() {
		switch msg.Type {
		case pb.ConsensusMessage_PRE_PREPARE:
			s.node.consensus.HandlePrePrepare(msg)
		case pb.ConsensusMessage_PREPARE:
			s.node.consensus.HandlePrepare(msg)
		case pb.ConsensusMessage_COMMIT:
			s.node.consensus.HandleCommit(msg)
		}
	}()

	return &pb.Empty{}, nil
}
