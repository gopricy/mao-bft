package maobft

import (
	"encoding"

	pb "github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)

type Message interface {
	encoding.BinaryMarshaler
}

type Common interface {
	Name() string
	SendEcho(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error
	SendReady(conn *grpc.ClientConn, merkleRoot []byte) error
	pb.ReadyServer
	pb.EchoServer
}

type Mao interface {
	SendPrepare(conn *grpc.ClientConn, merkleProof *pb.MerkleProof, data []byte) error
	Common
}

type MaoFollower interface {
	pb.PrepareServer
	Common
}
