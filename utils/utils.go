package mao_utils

import (
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/pb"
)

func IsSameBytes(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i, val := range left {
		if right[i] != val {
			return false
		}
	}
	return true
}

func FromBytesToBlock(bytes []byte) (pb.Block, error) {
	var Block pb.Block
	err := proto.Unmarshal(bytes, &Block)
	return Block, err
}

// Create a block from:
// 1. A list of transactions
// 2. Previous Hash
// 3. Sequence number
func CreateBlockFromTxsAndPrevHash(txs []pb.Transaction, prevHash []byte, seq int) (pb.Block, error) {
	block := pb.Block{}
	block.Content.PrevHash = prevHash
	block.Content.SeqNumber = int32(seq)
	for _, tx := range txs {
		block.Content.Txs = append(block.Content.Txs, &tx)
	}
	bytes, err := proto.Marshal(&block)
	if err != nil {
		return pb.Block{}, err
	}
	h := sha256.New()
	if _, err := h.Write(bytes); err != nil {
		return pb.Block{}, nil
	}
	block.CurHash = h.Sum(nil)
	return block, nil
}