package rbc

import (
	"context"
	"google.golang.org/grpc"
)

type ReadyClientWrapper struct{}

func (ReadyClientWrapper) SendReady(conn *grpc.ClientConn, merkleRoot string) error{
	request := &ReadyRequest{MerkleRoot: merkleRoot}
	_, err := NewReadyClient(conn).Ready(context.Background(), request)
	return err
}

