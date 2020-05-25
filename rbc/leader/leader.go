package leader

import (
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/erasure"
	"github.com/gopricy/mao-bft/rbc/merkle"
)

type Leader struct {
	name string
	common.Common
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


func (l *Leader) SendPrepare(p *common.Peer, merkleProof *pb.MerkleProof, data []byte) {
	payload := &pb.Payload{
		MerkleProof: merkleProof,
		Data:        data,
	}
	/*
		go func() {
			for {
				_, err := pb.NewPrepareClient(p.GetConn()).Prepare(context.Background(), payload)
				if err == nil {
					break
				}
			}
		}()
	*/
	_, err := pb.NewPrepareClient(p.GetConn()).Prepare(l.CreateContext(), payload)
	if err != nil {
		panic(err)
	}
}
