package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)

type ReadyClientWrapper struct{}

// Send Ready upon receiving N-f distinct Echos and successfully validated
func (ReadyClientWrapper) SendReady(conn *grpc.ClientConn, merkleRoot []byte) error {
	request := &pb.ReadyRequest{MerkleRoot: merkleRoot}
	_, err := pb.NewReadyClient(conn).Ready(context.Background(), request)
	return err
}
