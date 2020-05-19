package leader

import (
	"github.com/gopricy/mao-bft/rbc/common"
)

type Leader struct {
	name string
	common.Common
	common.PrepareClientWrapper
}

func NewLeader(name string, app common.Application) Leader {
	return Leader{name: name, Common: common.Common{App: app}}
}

func (l *Leader) RBCSend(b []byte) error{
	// erasure b to N shards
	// send it out
	for _, p := range l.KnownPeers{
		l.SendPrepare(p, nil, nil)
	}
	return nil
}

func (l *Leader) Name() string{
	return l.name
}

