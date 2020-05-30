package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/merkle"
	"github.com/pkg/errors"
)

// Echo serves echo messages from other nodes
func (c *Common) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	// TODO: validate after merkle is fixed
	//valid := merkle.VerifyProof(req.MerkleProof, merkle.BytesContent(req.Data))
	//if !valid {
	//	return nil, merkle.InvalidProof{}
	//}
	// Echo calls
	actualData, verified, name := c.Verify(ctx, req.Data)
	if !verified{
		return nil, errors.New("signature invalid")
	}
	c.Debugf(`ECHO Message: "%.4s"`, actualData)
	req.Data = actualData
	e, err := c.EchosReceived.Add(name, req.MerkleProof.Root, req)
	if err != nil {
		return nil, err
	}
	if e == len(c.RBCSetting.AllPeers)-c.ByzantineLimit {
		// TODO: interpolate {s'j} from any N-2f leaves received
		// TODO: recompute Merkle root h' and if h'!=h then abort
		if !c.readyIsSent(req.MerkleProof.Root) {
			for _, p := range c.RBCSetting.AllPeers {
				c.Debugf("Send READY to %#v", p)
				c.SendReady(p, req.MerkleProof.Root)
			}
		}
	}
	rootString := merkle.MerkleRootToString(req.MerkleProof.Root)
	// 2f + 1 Ready and N - 2f Echo, decode and apply
	if e == len(c.RBCSetting.AllPeers)-2*c.ByzantineLimit {
		if len(c.ReadiesReceived.rec[rootString]) >= 2*c.ByzantineLimit+1 {
			c.Infof("Get enough READY and ECHO to decode")
			data, err := c.reconstructData(rootString)
			if err != nil {
				return nil, err
			}
			// TODO: add it back when block and app is ready
			//block, err := maoUtils.FromBytesToBlock(data)
			//if err != nil {
			//	return nil, err
			//}
			c.Infof("Data reconstructed %.6s", data)
			if err := c.App.RBCReceive(data); err != nil {
				return nil, errors.Wrap(err, "Failed to apply the transaction")
			}
		}
	}
	return &pb.EchoResponse{}, nil
}

// Send Echo when a Prepare message is received
func (c *Common) SendEcho(p *Peer, merkleProof *pb.MerkleProof, data []byte) {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        c.Sign(data),
	}

	/*go func() {
		for {
			_, err := pb.NewEchoClient(p.GetConn()).Echo(context.Background(), payload)
			if err == nil {
				break
			}
		}
	}()*/
	_, err := pb.NewEchoClient(p.GetConn()).Echo(c.CreateContext(), payload)
	if err != nil {
		panic(err)
	}
}


