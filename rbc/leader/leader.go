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

func (l *Leader) Name() string{
	return l.name
}

