package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
)

type PrepareClientWrapper struct {
}

// Leader sends Prepare messages to all Followers
func (PrepareClientWrapper) SendPrepare(p Peer, merkleProof *pb.MerkleProof, data []byte) error {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	conn, err := createConnection(p.IP, p.PORT)
	if err != nil{
		return err
	}
	_, err = pb.NewPrepareClient(conn).Prepare(context.Background(), payload)
	return err
}
