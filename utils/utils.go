package mao_utils

import (
	"crypto/sha256"
	"encoding/hex"

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

func DecodeBlock(bytes []byte) (*pb.Block, error) {
	block := new(pb.Block)
	err := proto.Unmarshal(bytes, block)
	return block, err
}

func EncodeBlock(b *pb.Block) ([]byte, error) {
	return proto.Marshal(b)
}

// Create a block from:
// 1. A list of transactions
// 2. Previous Hash
func CreateBlockFromTxsAndPrevHash(txs []*pb.Transaction, prevHash []byte) (*pb.Block, error) {
	block := pb.Block{Content: &pb.BlockContent{Txs: txs, PrevHash: prevHash}}
	bytes, err := proto.Marshal(block.Content)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if _, err := h.Write(bytes); err != nil {
		return nil, nil
	}
	block.CurHash = h.Sum(nil)
	return &block, nil
}

func IsValidBlockHash(block *pb.Block) bool {
	h := sha256.New()
	byteContent, _ := proto.Marshal(block.Content)
	if _, err := h.Write(byteContent); err != nil {
		return false
	}
	return IsSameBytes(h.Sum(nil), block.CurHash)
}

func GetLastBlockFromArray(blocks []*pb.Block) *pb.Block {
	return blocks[len(blocks)-1]
}

func IsSameBlock(left *pb.Block, right *pb.Block) bool {
	if !IsSameBytes(left.CurHash, right.CurHash) || !IsValidBlockHash(left) || !IsValidBlockHash(right) {
		return false
	}
	return true
}

// File name is of format <state>_<hex CurHash>
func GetFileNameFromBlockDump(dump pb.BlockDump) string {
	stage := dump.State.String()
	hash := hex.EncodeToString(dump.Block.CurHash)
	return stage + "_" + hash
}
