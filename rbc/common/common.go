package common

import (
	"context"
	"github.com/gopricy/mao-bft/merkle"
	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc/peer"
	"github.com/pkg/errors"
	"sync"
)

type EchoReceived struct{
	echos map[merkle.RootString]map[string]*pb.Payload
	mu sync.Mutex
}

func (er *EchoReceived) Add(ip string, pl *pb.Payload) (int, error){
	er.mu.Lock()
	root := merkle.MerkleRootToString(pl.MerkleProof.Root)
	if _, ok := er.echos[root]; !ok{
		// if this message hasn't been seen
		er.echos[root] = make(map[string]*pb.Payload)
	}
	if _, ok := er.echos[root][ip]; ok{
		return len(er.echos[root]), errors.New("Duplicate ECHO from same IP carrying same message")
	}
	er.echos[root][ip] = pl
	er.mu.Unlock()
	return len(er.echos[root]), nil
}

type ReadyReceived struct{
	readies map[merkle.RootString]map[string]struct{}
	mu sync.Mutex
}

func (rr *ReadyReceived) Add(ip string, pl *pb.ReadyRequest) (int, error){
	rr.mu.Lock()
	root := merkle.MerkleRootToString(pl.MerkleRoot)
	if _, ok := rr.readies[root]; !ok{
		// if this message hasn't been seen
		rr.readies[root] = make(map[string]struct{})
	}
	if _, ok := rr.readies[root][ip]; ok{
		return len(rr.readies[root]), errors.New("Duplicate ECHO from same IP carrying same message")
	}
	rr.readies[root][ip] = struct{}{}
	rr.mu.Unlock()
	return len(rr.readies[root]), nil
}


// Common is a building block of follower and leader
type Common struct {
	pb.UnimplementedEchoServer
	pb.UnimplementedReadyServer
	EchoClientWrapper
	ReadyClientWrapper
	KnownNodes []Common

	Echos EchoReceived
	Readies ReadyReceived
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
	c.Echos.Add(p.Addr.String(), req)

	return &pb.EchoResponse{}, nil
}

// Ready serves ready messages from other nodes
func (c *Common) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Can't get PeerInfo")
	}

	// TODO: after getting f+1 READY: Send Ready if not Sent
	_, err := c.Readies.Add(p.Addr.String(), req)
	if err != nil{
		return nil, err
	}
	return &pb.ReadyResponse{}, nil
}

func (c *Common) Name() string{
	return "no name yet"
}
