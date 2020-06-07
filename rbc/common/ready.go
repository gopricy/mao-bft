package common

import (
	"context"
	"encoding/hex"

	"github.com/fatih/color"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/merkle"
	"github.com/pkg/errors"
)

func (c *Common) SendReady(p *Peer, root []byte) {
	readyReq := &pb.ReadyRequest{
		MerkleRoot: c.Sign(root),
	}
	go func() {
		for {
			_, err := pb.NewReadyClient(p.GetConn()).Ready(c.CreateContext(), readyReq)
			if err == nil {
				break
			}
		}
	}()
	// _, err := pb.NewReadyClient(p.GetConn()).Ready(c.CreateContext(), readyReq)
	// if err != nil {
	// 	panic(err)
	// }
}

// Ready serves ready messages from other nodes
func (c *Common) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	root, verified, name := c.Verify(ctx, req.MerkleRoot)
	if !verified {
		return nil, errors.New("invalid signature")
	}
	if !c.PrevHashValid(req.PrevHash, req.MerkleRoot) {
		return nil, errors.New("block with same prevHash already voted")
	}

	c.SetColor(color.FgGreen)
	defer c.UnsetColor()
	c.Debugf(`------Ready Server------`)
	c.Debugf(`Get READY from "%s" with root "%.4s"`, name, merkle.MerkleRootToString(req.MerkleRoot))

	// TODO: after getting f+1 READY: Send Ready if not Sent
	r, err := c.ReadiesReceived.Add(name, root, struct{}{})
	if err != nil {
		return nil, errors.Wrap(err, "Can't add this MerkleRoot to readiesReceived")
	}

	if r == c.ByzantineLimit+1 {
		if !c.readyIsSent(root) {
			for _, p := range c.RBCSetting.AllPeers {
				c.Debugf("Send READY (in Ready) to %#v", p)
				// TODO: Don't understand why this SendReady always fail in GRPC
				c.SendReady(p, root)
			}
		}
	}

	rootString := merkle.MerkleRootToString(root)
	if r == 2*c.ByzantineLimit+1 {
		if len(c.EchosReceived.rec[rootString]) >= len(c.RBCSetting.AllPeers)-2*c.ByzantineLimit {
			c.Debugf("Get enough READY and ECHO to decode")
			data, err := c.reconstructData(rootString)
			if err != nil {
				return nil, err
			}
			// c.Debugf("Data reconstructed")
			// TODO: add it back when block and app is finished
			//block, err := mao_utils.FromBytesToBlock(data)
			//if err != nil {
			//	return nil, err
			//}

			c.Debugf("Data reconstructed %.4s", hex.EncodeToString(data))
			shouldSync, err := c.App.RBCReceive(data)
			if err != nil {
				return nil, errors.Wrap(err, "failed to apply the transaction")
			}
			if shouldSync {
				c.Synchronize()
			}
		}
	}

	return &pb.ReadyResponse{}, nil
}
