package maobft

import (
	"context"
	pb "github.com/gopricy/mao-bft/rbc"
)

type Client struct {
	name string
	pb.UnimplementedEchoServer
	pb.UnimplementedReadyServer
	pb.ReadyClientWrapper
	pb.EchoClientWrapper
}

func NewClient(name string) Client{
	return Client{name: name}
}

var _ RBCClient = &Client{}

func (s *Client) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	//TODO: implement Echo
	return &pb.PrepareResponse{}, nil
}

func (s *Client) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	//TODO: implement Echo
	return &pb.EchoResponse{}, nil
}

func (s *Client) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	//TODO: implement Ready
	return &pb.ReadyResponse{}, nil
}

// Apply take the transaction into effect
func (s *Client) Apply(msg Message) error{
	//TODO: implement Apply
	return nil
}

func (s *Client) Name() string{
	return s.name
}
