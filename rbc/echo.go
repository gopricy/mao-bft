package rbc

import (
	"context"
	"google.golang.org/grpc"
)

type EchoClientWrapper struct {

}

func (EchoClientWrapper) SendEcho(conn *grpc.ClientConn, merkleRoot string, merkleBranch []string, data []byte) error{
	payload := &Payload{
		MerkleRoot:           merkleRoot,
		MerkleBranch:         merkleBranch,
		Data:                 data,
	}
	_, err := NewEchoClient(conn).Echo(context.Background(), payload)
	return err
}
