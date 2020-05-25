package common

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
)

// Prepare serves Prepare messages sent from Leader
func (c *Common) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	c.Debugf(`Get PREPARE: data "%s" in payload`, req.Data)
	for _, p := range c.AllPeers {
		c.Debugf(`Send ECHO with data "%s" to: %#v`, req.Data, p)
		c.SendEcho(p, req.MerkleProof, req.Data)
	}
	return &pb.PrepareResponse{}, nil
}
