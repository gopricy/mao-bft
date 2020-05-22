package leader

import (
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/erasure"
	"github.com/gopricy/mao-bft/rbc/merkle"
)

type Leader struct {
	name string
	common.Common
	common.PrepareClientWrapper
}

func NewLeader(name string, app common.Application) Leader {
	return Leader{name: name, Common: common.Common{App: app}}
}

func (l *Leader) Name() string {
	return l.name
}

func (l *Leader) RBCSend(data []byte) {
	splits, err := erasure.Split(data, l.ByzantineLimit, len(l.AllPeers))
	if err != nil {
		panic(err)
	}
	var contents []merkle.Content
	for _, s := range splits {
		contents = append(contents, merkle.BytesContent(s))
	}

	merkleTree := &merkle.MerkleTree{}
	merkleTree.Init(contents)

	for i, p := range l.AllPeers {
		proof, err := merkle.GetProof(merkleTree, contents[i])
		if err != nil {
			panic(err)
		}
		l.SendPrepare(p, proof, splits[i])
	}
}
