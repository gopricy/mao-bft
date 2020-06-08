package common

import (
	"context"
	"encoding/hex"

	"github.com/fatih/color"
	"github.com/gopricy/mao-bft/pb"
	"github.com/pkg/errors"
)

// Prepare serves Prepare messages sent from Leader
func (c *Common) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	c.SetColor(color.FgBlue)
	defer c.UnsetColor()
	c.Debugf(`------PREPARE Server------`)

	actualData, verified, name := c.Verify(ctx, req.Data)
	if !verified {
		return nil, errors.New("invalid signature")
	}
	if !c.PrevHashValid(req.PrevHash, req.MerkleProof.Root) {
		return nil, errors.New("can't vote on two blocks with same prevHash")
	}
	c.Debugf(`Get PREPARE: "%.4s" from %s`, hex.EncodeToString(actualData), name)
	for _, p := range c.AllPeers {
		c.Debugf(`Send ECHO "%.4s" to %#v`, hex.EncodeToString(actualData), p)
		c.SendEcho(p, req.MerkleProof, actualData)
		if c.Mode == 3 {
			c.Infof("Byzantine Mode 3(send ready when not): send ready to %s", p.Name)
			c.SendReady(p, req.MerkleProof.Root)
		}
	}

	return &pb.PrepareResponse{}, nil
}
