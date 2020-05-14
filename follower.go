package maobft

import (
	"context"

	"github.com/gopricy/mao-bft/pb"
)

type Follower struct {
	common
	pb.UnimplementedPrepareServer
}

func NewFollower(name string) Follower {
	return Follower{common: common{name: name}}
}

func (s *Follower) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	//TODO: implement
	return &pb.PrepareResponse{}, nil
}

var _ MaoFollower = &Follower{}

