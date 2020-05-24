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

func NewLeader(name string, app common.Application, faultLimit int, peers []*common.Peer) Leader {
	setting := common.RBCSetting{AllPeers: peers, ByzantineLimit: faultLimit}
	return Leader{Common: common.NewCommon(name, setting, app)}
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
	if err := merkleTree.Init(contents); err != nil{
		panic(err)
	}

	for i, p := range l.AllPeers {
		l.Infof(`Send PREPARE "%s" to %#v`, splits[i], p)
		proof, err := merkle.GetProof(merkleTree, contents[i])
		if err != nil {
			panic(err)
		}
		l.SendPrepare(p, proof, splits[i])
	}
}
