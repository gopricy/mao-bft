package common

import (
	"context"

	"github.com/gopricy/mao-bft/pb"
)

type ReadyClientWrapper struct{}

// SendReady is an async call upon receiving N-f distinct Echos and successfully validated or f+1 READY
func (ReadyClientWrapper) SendReady(peer Peer, merkleRoot []byte) {
	request := &pb.ReadyRequest{MerkleRoot: merkleRoot}
	go func() {
		for {
			_, err := pb.NewReadyClient(peer.GetConn()).Ready(context.Background(), request)
			if err == nil {
				break
			}
		}
	}()
}
