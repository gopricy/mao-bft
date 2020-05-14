package rbc

import (
	"context"
	"google.golang.org/grpc"
)

type PrepareClientWrapper struct {
}

func (PrepareClientWrapper) SendPrepare(conn *grpc.ClientConn, merkleRoot string, merkleBranch []string, data []byte) error{
	payload := &Payload{
		MerkleRoot:           merkleRoot,
		MerkleBranch:         merkleBranch,
		Data:                 data,
	}
	_, err := NewPrepareClient(conn).Prepare(context.Background(), payload)
	return err
}


