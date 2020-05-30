package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"github.com/pkg/errors"
)

// Prepare serves Prepare messages sent from Leader
func (c *Common) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	c.Debugf(`Get PREPARE: "%.4s"`, req.Data)
	actualData, verified, _ := c.Verify(ctx, req.Data)
	if !verified{
		return nil, errors.New("invalid signature")
	}
	for _, p := range c.AllPeers {
		c.Debugf(`Send ECHO "%.4s" to %#v`, actualData, p)
		c.SendEcho(p, req.MerkleProof, actualData)
	}

	return &pb.PrepareResponse{}, nil
}
