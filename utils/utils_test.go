package mao_utils

import (
	"crypto/sha256"
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test helper functions.
func hashProtoMessage(m proto.Message) ([]byte, error) {
	bytes, err := proto.Marshal(m)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	if _, err := h.Write(bytes); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}


func TestIsSameBytes(t *testing.T) {
	assert.Equal(t, true, IsSameBytes([]byte{1, 1}, []byte{1, 1}))
	assert.Equal(t, false, IsSameBytes([]byte{1, 1}, []byte{1, 2}))
}

func TestFromBytesToBlock(t *testing.T) {
	block := pb.Block{
		Content: &pb.BlockContent{Txs: []*pb.Transaction{{TransactionUuid: "abc"}}},
		CurHash: []byte{1, 2},
	}
	bytes, err := proto.Marshal(&block)
	assert.Nil(t, err)
	reBlock, reErr := FromBytesToBlock(bytes)
	assert.Nil(t, reErr)
	assert.True(t, IsSameBytes(reBlock.CurHash, block.CurHash))
	assert.Equal(t, reBlock.Content.Txs[0].TransactionUuid, "abc")
}

func TestFromBytesToBlock_ReconstructIsStillValid(t *testing.T) {
	txs := []*pb.Transaction{{TransactionUuid: "abc"}}
	curHash := []byte{1, 2}
	block, err := CreateBlockFromTxsAndPrevHash(txs, curHash)
	assert.Nil(t, err)
	assert.True(t, IsValidBlockHash(block))

	bytes, err := proto.Marshal(block)
	assert.Nil(t, err)
	reBlock, reErr := FromBytesToBlock(bytes)
	assert.Nil(t, reErr)
	assert.True(t, IsSameBytes(reBlock.CurHash, block.CurHash))
	assert.Equal(t, reBlock.Content.Txs[0].TransactionUuid, "abc")
	assert.True(t, IsValidBlockHash(block))
}

func TestCreateBlockFromTxsAndPrevHash(t *testing.T) {
	txs := []*pb.Transaction{{TransactionUuid: "a"}, {TransactionUuid: "b"}}
	prevHash := []byte{1, 2}
	expectedContent := pb.BlockContent{Txs: txs, PrevHash: prevHash}
	actual, err := CreateBlockFromTxsAndPrevHash(txs, prevHash)
	assert.Nil(t, err)
	assert.True(t, IsValidBlockHash(actual))
	assert.True(t, IsSameBytes(expectedContent.PrevHash, actual.Content.PrevHash))
	assert.Equal(t, len(actual.Content.Txs), len(expectedContent.Txs))
	for i, tx := range actual.Content.Txs {
		assert.Equal(t, tx.TransactionUuid, expectedContent.Txs[i].TransactionUuid)
	}
}

func TestIsValidBlockHash(t *testing.T) {
	txs := []*pb.Transaction{{TransactionUuid: "a"}, {TransactionUuid: "b"}}
	prevHash := []byte{1, 2}
	expectedContent := pb.BlockContent{Txs: txs, PrevHash: prevHash}
	hash, err := hashProtoMessage(&expectedContent)
	assert.Nil(t, err)
	block := pb.Block{Content: &expectedContent, CurHash: hash}
	assert.True(t, IsValidBlockHash(&block))
	block.CurHash = nil
	assert.False(t, IsValidBlockHash(&block))
}

func TestIsSameBlock(t *testing.T) {
	txs := []*pb.Transaction{{TransactionUuid: "abc"}}
	curHash := []byte{1, 2}
	block, err := CreateBlockFromTxsAndPrevHash(txs, curHash)
	assert.Nil(t, err)
	assert.True(t, IsValidBlockHash(block))

	bytes, err := proto.Marshal(block)
	assert.Nil(t, err)
	reBlock, reErr := FromBytesToBlock(bytes)
	assert.Nil(t, reErr)
	assert.True(t, IsValidBlockHash(reBlock))
	assert.True(t, IsSameBlock(block, reBlock))
}