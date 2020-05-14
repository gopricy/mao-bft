package maobft

import (
	"encoding"
	pb "github.com/gopricy/mao-bft/rbc"
	"google.golang.org/grpc"
)

type Message interface{
	encoding.BinaryMarshaler
}

type node interface{
	Apply(Message) error
	Name() string
	SendEcho(conn *grpc.ClientConn, merkleRoot string, merkleBranch []string, data []byte) error
	SendReady(conn *grpc.ClientConn, merkleRoot string) error
	pb.ReadyServer
	pb.EchoServer
}

type RBCServer interface{
	SendPrepare(conn *grpc.ClientConn, merkleRoot string, merkleBranch []string, data []byte) error
	node
}

type RBCClient interface{
	pb.PrepareServer
	node
}
