package rbc

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)

type EchoClientWrapper struct {
}

func (EchoClientWrapper) SendEcho(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	_, err := pb.NewEchoClient(conn).Echo(context.Background(), payload)
	return err
}
