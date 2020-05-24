package common

import (
	"context"

	"github.com/gopricy/mao-bft/pb"
)

type EchoClientWrapper struct {
}

// Send Echo when a Prepare message is received
func (EchoClientWrapper) SendEcho(p *Peer, merkleProof *pb.MerkleProof, data []byte) {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}

	/*go func() {
		for {
			_, err := pb.NewEchoClient(p.GetConn()).Echo(context.Background(), payload)
			if err == nil {
				break
			}
		}
	}()*/
	_, err := pb.NewEchoClient(p.GetConn()).Echo(context.Background(), payload)
	if err != nil {
		panic(err)
	}
}
