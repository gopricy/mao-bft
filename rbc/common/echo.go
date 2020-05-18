package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
)

type EchoClientWrapper struct {
}

// Send Echo when a Prepare message is received
func (EchoClientWrapper) SendEcho(p Peer, merkleProof *pb.MerkleProof, data []byte) error {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	conn, err := createConnection(p.IP, p.PORT)
	if err != nil{
		return err
	}
	_, err = pb.NewEchoClient(conn).Echo(context.Background(), payload)
	return err
}
