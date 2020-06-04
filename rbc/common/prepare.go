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
	color.Set(color.FgBlue)
	defer color.Unset()
	c.Debugf(color.BlueString(`Get PREPARE: "%.4s"`, hex.EncodeToString(req.Data)))

	actualData, verified, _ := c.Verify(ctx, req.Data)
	if !verified {
		return nil, errors.New("invalid signature")
	}
	if !c.PrevHashValid(req.PrevHash, req.MerkleProof.Root){
		return nil, errors.New("can't vote on two blocks with same prevHash")
	}
	for _, p := range c.AllPeers {
		c.Debugf(`Send ECHO "%.4s" to %#v`, hex.EncodeToString(actualData), p)
		c.SendEcho(p, req.MerkleProof, actualData)
	}

	return &pb.PrepareResponse{}, nil
}
