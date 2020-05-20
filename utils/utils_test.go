package mao_utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsSameBytes(t *testing.T) {
	assert.Equal(t, true, IsSameBytes([]byte{1, 1}, []byte{1, 1}))
	assert.Equal(t, false, IsSameBytes([]byte{1, 1}, []byte{1, 2}))
}

func TestFromBytesToBlock(t *testing.T) {
	block := pb.Block{
		Content: &pb.BlockContent{SeqNumber: 1},
		CurHash: []byte{1, 2},
	}
	bytes, err := proto.Marshal(&block)
	assert.Nil(t, err)
	reBlock, reErr := FromBytesToBlock(bytes)
	assert.Nil(t, reErr)
	assert.True(t, IsSameBytes(reBlock.CurHash, block.CurHash))
	assert.Equal(t, reBlock.Content.SeqNumber, block.Content.SeqNumber)
}