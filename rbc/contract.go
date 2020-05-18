package rbc

import (
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"

	pb "github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)


type Common interface {
	Name() string
	SendEcho(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error
	SendReady(conn *grpc.ClientConn, merkleRoot []byte) error
	pb.ReadyServer
	pb.EchoServer
}

var _ Common = &common.Common{}

type Mao interface {
	SendPrepare(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error
	Common
}

var _ Mao = &leader.Leader{}

type MaoFollower interface {
	pb.PrepareServer
	Common
}

var _MaoFollower = &follower.Follower{}
