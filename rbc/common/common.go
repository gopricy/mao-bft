package common

import (
	"context"
	"github.com/gopricy/mao-bft/erasure"
	"github.com/gopricy/mao-bft/merkle"
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"github.com/pkg/errors"
	"google.golang.org/grpc/peer"
	"sync"
)


type Received struct{
	// TODO: improve the efficiency with better locking
	rec map[merkle.RootString]map[string]interface{}
	mu sync.Mutex
}

func (er *Received) Add(ip string, merkleRoot []byte, rec interface{}) (int, error){
	er.mu.Lock()
	defer er.mu.Unlock()
	root := merkle.MerkleRootToString(merkleRoot)
	if _, ok := er.rec[root]; !ok{
		// if this message hasn't been seen
		er.rec[root] = make(map[string]interface{})
	}
	if _, ok := er.rec[root][ip]; ok{
		return len(er.rec[root]), errors.New("Duplicate ECHO from same IP carrying same message")
	}
	er.rec[root][ip] = rec
	return len(er.rec[root]), nil
}

type RBCSetting struct{
	AllPeers []Peer
	ByzantineLimit int
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

	RBCSetting

	EchosReceived   Received
	ReadiesReceived Received

	ReadiesSent sync.Map

	// Below are related to transaction system.
	App Application
	pb.UnimplementedTransactionServiceServer
	Queue EventQueue
}

func (c *Common) reconstructData(root merkle.RootString) ([]byte, error){
	payloads := []*pb.Payload{}
	for _, m := range c.EchosReceived.rec[root]{
		payloads = append(payloads, m.(*pb.Payload))
	}
	return erasure.Reconstruct(payloads, c.ByzantineLimit, len(c.AllPeers))
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
	valid := merkle.VerifyProof(req.MerkleProof, merkle.BytesContent(req.Data))
	if !valid{
		return nil, merkle.InvalidProof{}
	}
	// Echo calls
	p, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("Can't get PeerInfo")
	}
	e, err := c.EchosReceived.Add(p.Addr.String(), req.MerkleProof.Root, req)
	if err != nil{
		return nil, err
	}
	if e == len(c.RBCSetting.AllPeers) - c.ByzantineLimit{
		// TODO: interpolate {s'j} from any N-2f leaves received
		// TODO: recompute Merkle root h' and if h'!=h then abort
		if !c.readyIsSent(req.MerkleProof.Root){
			for _, p := range c.RBCSetting.AllPeers{
				err = c.SendReady(p, req.MerkleProof.Root)
				if err != nil{
					return nil, err
				}
			}
		}
	}
	rootString := merkle.MerkleRootToString(req.MerkleProof.Root)
	// 2f + 1 Ready and N - 2f Echo, decode and apply
	if e == len(c.RBCSetting.AllPeers) - 2 * c.ByzantineLimit{
		if len(c.ReadiesReceived.rec[rootString]) >= 2 * c.ByzantineLimit + 1{
			data, err := c.reconstructData(rootString)
			if err != nil{
				return nil, err
			}
			block, err := mao_utils.FromBytesToBlock(data)
			if err != nil {
				return nil, err
			}
			if err := c.App.RBCReceive(block); err != nil{
				return nil, err
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
	r, err := c.ReadiesReceived.Add(p.Addr.String(), req.MerkleRoot, struct{}{})
	if err != nil{
		return nil, err
	}

	if r == c.ByzantineLimit + 1{
		if !c.readyIsSent(req.MerkleRoot){
			for _, p := range c.RBCSetting.AllPeers{
				// TODO: PERFORMANCE probably need an unblocking call for performance
				if err = c.SendReady(p, req.MerkleRoot); err != nil{
					return nil, err
				}
			}
		}
	}

	merkleRoot := merkle.MerkleRootToString(req.MerkleRoot)
	if r == 2 * c.ByzantineLimit + 1{
		if len(c.EchosReceived.rec[merkleRoot]) >= len(c.RBCSetting.AllPeers) - 2 * c.ByzantineLimit{
			data, err := c.reconstructData(merkleRoot)
			if err != nil{
				return nil, err
			}
			block, err := mao_utils.FromBytesToBlock(data)
			if err != nil {
				return nil, err
			}
			if err := c.App.RBCReceive(block); err != nil{
				return nil, err
			}
		}
	}

	return &pb.ReadyResponse{}, nil
}

func (c *Common) Name() string{
	return "Name() should be implemented by Leader/Follower"
}

func (c *Common) ProposeTransaction(
	ctx context.Context, in *pb.ProposeTransactionRequest)(*pb.ProposeTransactionResponse, error) {
	// TODO(chenweilunster): IMPLEMENT ME
	return nil, nil
}

func (c *Common) GetTransactionStatus(
	ctx context.Context, in *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusResponse, error) {
	// TODO(chenweilunster): IMPLEMENT ME
	return nil, nil
}

