package common

import (
	"context"

	"github.com/gopricy/mao-bft/pb"
)

type PrepareClientWrapper struct {
}

// Leader sends Prepare messages to all Followers
func (PrepareClientWrapper) SendPrepare(p *Peer, merkleProof *pb.MerkleProof, data []byte) {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	/*
	go func() {
		for {
			_, err := pb.NewPrepareClient(p.GetConn()).Prepare(context.Background(), payload)
			if err == nil {
				break
			}
		}
	}()
	*/
	_, err := pb.NewPrepareClient(p.GetConn()).Prepare(context.Background(), payload)
	if err != nil {
		panic(err)
	}
}
