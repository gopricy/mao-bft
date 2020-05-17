package leader

import (
	"github.com/gopricy/mao-bft/rbc/common"
)

type Leader struct {
	name string
	common.Common
	common.PrepareClientWrapper
}

func NewLeader(name string) Leader {
	return Leader{name: name, Common: common.Common{}}
}

func (l *Leader) Name() string{
	return l.name
}

