package rbc

import (
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"

	pb "github.com/gopricy/mao-bft/pb"
)

type Common interface {
	Name() string
	SendEcho(common.Peer, *pb.MerkleProof, []byte)
	SendReady(common.Peer, []byte)
	pb.ReadyServer
	pb.EchoServer
}

var _ Common = &common.Common{}

type Mao interface {
	SendPrepare(common.Peer, *pb.MerkleProof, []byte)
	Common
	RBCSend([]byte)
}

var _ Mao = &leader.Leader{}

type MaoFollower interface {
	pb.PrepareServer
	Common
}

var _MaoFollower = &follower.Follower{}
