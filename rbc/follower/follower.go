package follower

import (
	"github.com/gopricy/mao-bft/rbc/common"
)

type Follower struct {
	common.Common
}

func NewFollower(name string, app common.Application, faultLimit int, peers map[string]*common.Peer, privateKey *[64]byte) *Follower {
	setting := common.RBCSetting{AllPeers: peers, ByzantineLimit: faultLimit}
	return &Follower{Common: common.NewCommon(name, setting, app, privateKey)}
}
