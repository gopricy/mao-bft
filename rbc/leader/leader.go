package leader

import (
	"encoding/hex"
	"math"
	"time"

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
	block, err := mao_utils.DecodeBlock(bytes)
	if err != nil {
		panic(err)
	}

	splits := [][]byte{}
	l.Debugf("Split data into %d shards with any %d shards can reconstruct data",
		len(l.AllPeers), len(l.AllPeers)-2*l.ByzantineLimit)

	splits, err = erasure.Split(bytes, l.ByzantineLimit, len(l.AllPeers))
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
		switch l.Mode {
		case 1:
			l.Infof(`Byzantine Mode 1(send the same data shard to all peers): PREPARE "%.4s" to %#v`, hex.EncodeToString(splits[0]), p)
			proof, err := merkle.GetProof(merkleTree, contents[0])
			if err != nil {
				panic(err)
			}
			l.SendPrepare(p, proof, block.Content.PrevHash, splits[0])
		default:
			l.Debugf(`Send PREPARE "%.4s" to %#v`, hex.EncodeToString(splits[i]), p)
			proof, err := merkle.GetProof(merkleTree, contents[i])
			if err != nil {
				panic(err)
			}
			l.SendPrepare(p, proof, block.Content.PrevHash, splits[i])
		}
		i++
	}
}

func (l *Leader) SendPrepare(p *common.Peer, merkleProof *pb.MerkleProof, prevHash []byte, data []byte) {

	payload := &pb.Payload{
		MerkleProof: merkleProof,
		PrevHash:    prevHash,
		Data:        l.Sign(data),
	}
	if l.Mode == 2 {
		l.Infof(`Byzantine Mode 2(send data without signature)`)
		payload.Data = data
	}

	go func() {
		retry := 0
		for {
			_, err := pb.NewPrepareClient(p.GetConn()).Prepare(l.CreateContext(), payload)
			if err == nil {
				break
			}
			if l.Mode == 4 {
				retry++
				if retry > 10 {
					retry = 10
				}
				l.Debugf("SendPreapre failed. Wait for %d to reconnect.", int(math.Pow(2, float64(retry))))
				time.Sleep(time.Second * time.Duration(int(math.Pow(2, float64(retry)))))
			}
		}
	}()
	// _, err := pb.NewPrepareClient(p.GetConn()).Prepare(l.CreateContext(), payload)
	// if err != nil {
	// 	panic(err)
	// }
}
