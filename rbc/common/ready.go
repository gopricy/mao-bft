package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
)

type ReadyClientWrapper struct{}

// Send Ready upon receiving N-f distinct Echos and successfully validated or f+1 READY
func (ReadyClientWrapper) SendReady(peer Peer, merkleRoot []byte) error {
	request := &pb.ReadyRequest{MerkleRoot: merkleRoot}
	conn, err := createConnection(peer.IP, peer.PORT)
	if err != nil{
		return err
	}
	_, err = pb.NewReadyClient(conn).Ready(context.Background(), request)
	return err
}
