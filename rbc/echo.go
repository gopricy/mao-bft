package rbc

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)

type EchoClientWrapper struct {
}

// Send Echo when a Prepare message is received
func (EchoClientWrapper) SendEcho(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	_, err := pb.NewEchoClient(conn).Echo(context.Background(), payload)
	return err
}
