package leader

import (
	"encoding/hex"

	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/erasure"
	"github.com/gopricy/mao-bft/rbc/merkle"
	mao_utils "github.com/gopricy/mao-bft/utils"
)

type Leader struct {
	name string
	common.Common
}

func NewLeader(name string, app common.Application, faultLimit int, peers map[string]*common.Peer, privateKey *[64]byte) *Leader {
	setting := common.RBCSetting{AllPeers: peers, ByzantineLimit: faultLimit}
	return &Leader{Common: common.NewCommon(name, setting, app, privateKey)}
}

func (l *Leader) RBCSend(bytes []byte) {
	//bytes, err := proto.Marshal(block)
	//if err != nil{
	//	panic(err)
	//}
	block, err := mao_utils.DecodeBlock(bytes)
	if err != nil {
		panic(err)
	}

	splits, err := erasure.Split(bytes, l.ByzantineLimit, len(l.AllPeers))
	if err != nil {
		panic(err)
	}
	var contents []merkle.Content
	for _, s := range splits {
		contents = append(contents, merkle.BytesContent(s))
	}

	merkleTree := &merkle.MerkleTree{}
	if err := merkleTree.Init(contents); err != nil {
		panic(err)
	}

	i := 0
	for _, p := range l.AllPeers {
		l.Debugf(`Send PREPARE "%.4s" to %#v`, hex.EncodeToString(splits[i]), p)
		proof, err := merkle.GetProof(merkleTree, contents[i])
		if err != nil {
			panic(err)
		}
		l.SendPrepare(p, proof, block.Content.PrevHash, splits[i])
		i++
	}
}

func (l *Leader) SendPrepare(p *common.Peer, merkleProof *pb.MerkleProof, prevHash []byte, data []byte) {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		PrevHash:    prevHash,
		Data:        l.Sign(data),
	}
	go func() {
		for {
			_, err := pb.NewPrepareClient(p.GetConn()).Prepare(l.CreateContext(), payload)
			if err == nil {
				break
			}
		}
	}()
	// _, err := pb.NewPrepareClient(p.GetConn()).Prepare(l.CreateContext(), payload)
	// if err != nil {
	// 	panic(err)
	// }
}
