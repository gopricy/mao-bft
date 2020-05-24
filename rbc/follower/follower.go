package follower

import (
	"github.com/gopricy/mao-bft/rbc/common"
)

type Follower struct {
	common.Common
}

func NewFollower(name string, app common.Application, faultLimit int, peers []*common.Peer) Follower {
	setting := common.RBCSetting{AllPeers: peers, ByzantineLimit: faultLimit}
	return Follower{Common: common.NewCommon(name, setting, app)}
}



