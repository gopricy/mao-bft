package maobft

import (
	"github.com/gopricy/mao-bft/rbc"
)

type Leader struct {
	common
	rbc.PrepareClientWrapper
}

var _ Mao = &Leader{}

func NewLeader(name string) Leader {
	return Leader{common:common{name: name}}
}

