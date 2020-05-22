package follower

import (
	"context"

	"github.com/gopricy/mao-bft/rbc/common"

	"github.com/gopricy/mao-bft/pb"
)

type Follower struct {
	name string
	common.Common
	pb.UnimplementedPrepareServer
}

func NewFollower(name string, app common.Application) Follower {
	return Follower{name: name, Common: common.Common{App: app}}
}

// Prepare serves Prepare messages sent from Leader
func (f *Follower) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	for _, p := range f.AllPeers {
		f.SendEcho(p, req.MerkleProof, req.Data)
	}
	return &pb.PrepareResponse{}, nil
}

func (f *Follower) Name() string {
	return f.name
}
