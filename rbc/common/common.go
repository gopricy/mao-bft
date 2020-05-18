package common

import (
	"context"
	"github.com/gopricy/mao-bft/merkle"
	"github.com/gopricy/mao-bft/pb"
	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
	"sync"
)

type EchoReceived struct{
	// TODO: improve the efficiency with better locking
	echos map[merkle.RootString]map[string]*pb.Payload
	mu sync.Mutex
}

func (er *EchoReceived) Add(ip string, pl *pb.Payload) (int, error){
	er.mu.Lock()
	defer er.mu.Unlock()
	root := merkle.MerkleRootToString(pl.MerkleProof.Root)
	if _, ok := er.echos[root]; !ok{
		// if this message hasn't been seen
		er.echos[root] = make(map[string]*pb.Payload)
	}
	if _, ok := er.echos[root][ip]; ok{
		return len(er.echos[root]), errors.New("Duplicate ECHO from same IP carrying same message")
	}
	er.echos[root][ip] = pl
	return len(er.echos[root]), nil
}

type ReadyReceived struct{
	readies map[merkle.RootString]map[string]struct{}
	mu sync.Mutex
}

func (rr *ReadyReceived) Add(ip string, pl *pb.ReadyRequest) (int, error){
	rr.mu.Lock()
	defer rr.mu.Unlock()
	root := merkle.MerkleRootToString(pl.MerkleRoot)
	if _, ok := rr.readies[root]; !ok{
		// if this message hasn't been seen
		rr.readies[root] = make(map[string]struct{})
	}
	if _, ok := rr.readies[root][ip]; ok{
		return len(rr.readies[root]), errors.New("Duplicate ECHO from same IP carrying same message")
	}
	rr.readies[root][ip] = struct{}{}
	return len(rr.readies[root]), nil
}

// TODO: PERFORMANCE we probably want to keep the sessions with each peer
type Peer struct{
	IP string
	PORT int
}

// Common is a building block of follower and leader
type Common struct {
	pb.UnimplementedEchoServer
	pb.UnimplementedReadyServer
	EchoClientWrapper
	ReadyClientWrapper

	KnownPeers []Peer
	ByzantineLimit int

	EchosReceived EchoReceived
	ReadiesReceived ReadyReceived

	ReadiesSent sync.Map
}

func (c *Common) readyIsSent(merkleroot []byte) bool{
	if _, ok := c.ReadiesSent.Load(merkle.MerkleRootToString(merkleroot)); !ok{
		c.ReadiesSent.Store(merkle.MerkleRootToString(merkleroot), struct{}{})
		return false
	}
	return true
}

// Echo serves echo messages from other nodes
func (c *Common) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	if !merkle.VerifyProof(*req.MerkleProof, merkle.BytesContent(req.Data)){
		return nil, merkle.InvalidProof{}
	}
	// Echo calls
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Can't get PeerInfo")
	}
	e, err := c.EchosReceived.Add(p.Addr.String(), req)
	if err != nil{
		return nil, err
	}
	if e == len(c.KnownPeers) - c.ByzantineLimit{
		// TODO: interpolate {s'j} from any N-2f leaves received
		// TODO: recompute Merkle root h' and if h'!=h then abort
		if !c.readyIsSent(req.MerkleProof.Root){
			for _, p := range c.KnownPeers{
				err = c.SendReady(p, req.MerkleProof.Root)
				if err != nil{
					return nil, err
				}
			}

		}
	}

	return &pb.EchoResponse{}, nil
}

// Ready serves ready messages from other nodes
func (c *Common) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Can't get PeerInfo")
	}

	// TODO: after getting f+1 READY: Send Ready if not Sent
	r, err := c.ReadiesReceived.Add(p.Addr.String(), req)
	if err != nil{
		return nil, err
	}

	if r == c.ByzantineLimit + 1{
		if !c.readyIsSent(req.MerkleRoot){
			for _, p := range c.KnownPeers{
				// TODO: PERFORMANCE probably need an unblocking call for performance
				if err = c.SendReady(p, req.MerkleRoot); err != nil{
					return nil, err
				}
			}
		}
	}
	return &pb.ReadyResponse{}, nil
}

func (c *Common) Name() string{
	return "Name() should be implemented by Leader/Follower"
}
