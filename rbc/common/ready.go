package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/merkle"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"github.com/pkg/errors"
)

func (c *Common) SendReady(p *Peer, root []byte) {
	readyReq := &pb.ReadyRequest{
		MerkleRoot: root,
	}
	/*go func() {
		for {
			_, err := pb.NewEchoClient(p.GetConn()).Echo(context.Background(), payload)
			if err == nil {
				break
			}
		}
	}()*/
	_, err := pb.NewReadyClient(p.GetConn()).Ready(c.CreateContext(), readyReq)
	if err != nil {
		panic(err)
	}
}

// Ready serves ready messages from other nodes
func (c *Common) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	name, err := c.GetNameFromContext(ctx)
	if err != nil{
		return nil, errors.Wrap(err, "Can't get name in context")
	}
	c.Debugf(`Get READY "%s" with root "%s"`, name, merkle.MerkleRootToString(req.MerkleRoot))

	// TODO: after getting f+1 READY: Send Ready if not Sent
	r, err := c.ReadiesReceived.Add(name, req.MerkleRoot, struct{}{})
	if err != nil {
		return nil, err
	}

	if r == c.ByzantineLimit+1 {
		if !c.readyIsSent(req.MerkleRoot) {
			for _, p := range c.RBCSetting.AllPeers {
				c.Debugf("Send Ready to %#v", p)
				c.SendReady(p, req.MerkleRoot)
			}
		}
	}

	merkleRoot := merkle.MerkleRootToString(req.MerkleRoot)
	if r == 2*c.ByzantineLimit+1 {
		if len(c.EchosReceived.rec[merkleRoot]) >= len(c.RBCSetting.AllPeers)-2*c.ByzantineLimit {
			data, err := c.reconstructData(merkleRoot)
			if err != nil {
				return nil, err
			}
			block, err := mao_utils.FromBytesToBlock(data)
			if err != nil {
				return nil, err
			}
			if err := c.App.RBCReceive(block); err != nil {
				return nil, err
			}
		}
	}

	return &pb.ReadyResponse{}, nil
}
