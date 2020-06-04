package rbc

import (
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"

	pb "github.com/gopricy/mao-bft/pb"
)

type Common interface {
	Name() string
	SendEcho(*common.Peer, *pb.MerkleProof, []byte)
	SendReady(*common.Peer, []byte)
	pb.ReadyServer
	pb.EchoServer
	pb.PrepareServer
}

var _ Common = &common.Common{}

type Mao interface {
	SendPrepare(*common.Peer, *pb.MerkleProof, []byte, []byte)
	// TODO: we can change it to block
	RBCSend([]byte)
	Common
}

var _ Mao = &leader.Leader{}

type MaoFollower interface {
	Common
}

var _MaoFollower = &follower.Follower{}
