package maobft

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc"
)

type common struct{
	name string
	pb.UnimplementedEchoServer
	pb.UnimplementedReadyServer
	rbc.EchoClientWrapper
	rbc.ReadyClientWrapper
}

// Echo serves echo messages from other nodes
func (c *common) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{}, nil
}

// Ready serves ready messages from other nodes
func (c *common) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	return &pb.ReadyResponse{}, nil
}

// Name is the node's name
func (c *common) Name() string {
	return c.name
}

