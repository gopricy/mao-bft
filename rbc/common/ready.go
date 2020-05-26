package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/merkle"
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
	c.Debugf(`Get READY from "%s" with root "%.4s"`, name, merkle.MerkleRootToString(req.MerkleRoot))

	// TODO: after getting f+1 READY: Send Ready if not Sent
	r, err := c.ReadiesReceived.Add(name, req.MerkleRoot, struct{}{})
	if err != nil {
		return nil, errors.Wrap(err, "Can't add this MerkleRoot to readiesReceived")
	}

	if r == c.ByzantineLimit+1 {
		if !c.readyIsSent(req.MerkleRoot) {
			for _, p := range c.RBCSetting.AllPeers {
				c.Debugf("Send READY (in Ready) to %#v", p)
				// TODO: Don't understand why this SendReady always fail in GRPC
				c.SendReady(p, req.MerkleRoot)
			}
		}
	}

	merkleRoot := merkle.MerkleRootToString(req.MerkleRoot)
	if r == 2*c.ByzantineLimit+1 {
		if len(c.EchosReceived.rec[merkleRoot]) >= len(c.RBCSetting.AllPeers)-2*c.ByzantineLimit {
			c.Infof("Get enough READY and ECHO to decode")
			data, err := c.reconstructData(merkleRoot)
			if err != nil {
				return nil, err
			}
			c.Infof("Data reconstructed")
			// TODO: add it back when block and app is finished
			//block, err := mao_utils.FromBytesToBlock(data)
			//if err != nil {
			//	return nil, err
			//}

			c.Infof("RBC Receive with data %.4s", data)
			if err := c.App.RBCReceive(data); err != nil {
				return nil, errors.Wrap(err, "failed to apply the apply the transaction")
			}
		}
	}

	return &pb.ReadyResponse{}, nil
}
