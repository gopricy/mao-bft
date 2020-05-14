package rbc

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)

type PrepareClientWrapper struct {
}

func (PrepareClientWrapper) SendPrepare(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	_, err := pb.NewPrepareClient(conn).Prepare(context.Background(), payload)
	return err
}
