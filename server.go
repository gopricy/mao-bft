package maobft

import (
	"context"
	pb "github.com/gopricy/mao-bft/rbc"
)

type Server struct {
	name string
	pb.UnimplementedEchoServer
	pb.UnimplementedReadyServer
	pb.EchoClientWrapper
	pb.ReadyClientWrapper
	pb.PrepareClientWrapper
}

var _ RBCServer = &Server{}

// TODO: probably need more arguments in the future
func NewServer(name string) Server{
	return Server{name: name}
}

func (s *Server) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	return &pb.EchoResponse{}, nil
}

func (s *Server) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	return &pb.ReadyResponse{}, nil
}

// Apply take the transaction into effect
func (s *Server) Apply(data Message) error{
	return nil
}

// Name is the server name
func (s *Server) Name() string{
	return s.name
}

