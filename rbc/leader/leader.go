package leader

import (
	"github.com/gopricy/mao-bft/erasure"
	"github.com/gopricy/mao-bft/merkle"
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

func (l *Leader) RBCSend(data []byte) error{
	splits, err := erasure.Split(data, l.ByzantineLimit, len(l.AllPeers))
	if err != nil{
		return err
	}
	var contents []merkle.Content
	for _, s := range splits{
		contents = append(contents, merkle.BytesContent(s))
	}

	merkleTree := &merkle.MerkleTree{}
	merkleTree.Init(contents)

	for i, p := range l.AllPeers{
		proof, err := merkle.GetProof(merkleTree, contents[i])
		if err != nil{
			return err
		}
		l.SendPrepare(p, proof, splits[i])
	}
	return nil
}

