package common

import (
	"context"
	"errors"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"log"

	"github.com/gopricy/mao-bft/pb"
)

// Sync receives a sync request, and return a sync response.
func (c *Common) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	res, err := c.App.GetSyncAnswer(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Send sync towards given peer.
func (c *Common) SendSync(p *Peer) (*pb.SyncResponse, error) {
	req, err := c.App.GetSyncQuestion()
	if err != nil {
		log.Fatalln("GetSyncQuestion fails: " + err.Error())
	}
	res, err := pb.NewSyncClient(p.GetConn()).Sync(c.CreateContext(), req)
	if err != nil {
		return nil, err
	}
	// Validate the response is actually valid. This is very important because no one can fake answer.
	begin, err := mao_utils.DecodeBlock(req.LastCommit)
	for _, blockBytes := range append(res.Response, req.LatestStaged) {
		next, err := mao_utils.DecodeBlock(blockBytes)
		if err != nil ||
			!mao_utils.IsValidBlockHash(next) ||
			!mao_utils.IsSameBytes(begin.CurHash, next.Content.PrevHash) {
			return nil, errors.New("Peer's answer is not valid. Skip this peer: " + p.Name)
		}
		begin = next
	}
	// The result is valid, we return the response to Sync issuer.
	return res, nil
}

func (c *Common) Synchronize() {
	// We do best effort round robin sync request to each peer.
	for _, peer := range c.AllPeers {
		res, err := c.SendSync(peer)
		if err != nil {
			c.Infof("Skip since no valid answer or RPC is rejected for peer: " +
				peer.Name + ". Error is: " + err.Error())
			continue
		}
		for _, bytes := range res.Response {
			_, err := c.App.RBCReceive(bytes)
			if err != nil {
				log.Fatalln("Fail to apply sync's response.")
			}
		}
		return
	}
}