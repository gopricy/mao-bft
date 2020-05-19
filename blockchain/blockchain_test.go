package blockchain

import (
	"github.com/gopricy/mao-bft/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOrderedInsertToList(t *testing.T) {
	bc := Blockchain{}
	bc.Init()
	// Try insert sequence 1
	OrderedInsertToList(&bc.Staged, pb.Block{Content: &pb.BlockContent{SeqNumber: 1}})
	assert.Equal(t, 1, GetStagedAreaSize(bc.Staged))

	// Try insert sequence 3
	OrderedInsertToList(&bc.Staged, pb.Block{Content: &pb.BlockContent{SeqNumber: 3}})
	assert.Equal(t, 2, GetStagedAreaSize(bc.Staged))

	lastStaged, err := GetLastBlockInStagedArea(bc.Staged)
	assert.Nil(t, err)
	assert.Equal(t, int32(3), lastStaged.Value.Content.SeqNumber)

	// Try insert sequence 2
	// Try insert sequence 3
	OrderedInsertToList(&bc.Staged, pb.Block{Content: &pb.BlockContent{SeqNumber: 2}})
	assert.Equal(t, 3, GetStagedAreaSize(bc.Staged))

	lastStaged, err = GetLastBlockInStagedArea(bc.Staged)
	assert.Nil(t, err)
	assert.Equal(t, int32(3), lastStaged.Value.Content.SeqNumber)
}

func TestRemoveStagedBlock(t *testing.T) {
	bc := Blockchain{}
	bc.Init()
	// Try insert sequences
	assert.True(t, OrderedInsertToList(&bc.Staged, pb.Block{Content: &pb.BlockContent{SeqNumber: 1}}))
	assert.True(t, OrderedInsertToList(&bc.Staged, pb.Block{Content: &pb.BlockContent{SeqNumber: 5}}))
	assert.True(t, OrderedInsertToList(&bc.Staged, pb.Block{Content: &pb.BlockContent{SeqNumber: 3}}))

	// remove 1
	RemoveStagedBlock(bc.Staged.Next)

	assert.Equal(t, 2, GetStagedAreaSize(bc.Staged))
	lastBlock, err := GetLastBlockInStagedArea(bc.Staged)
	assert.Nil(t, err)
	assert.Equal(t, int32(5), lastBlock.Value.Content.SeqNumber)
	assert.Equal(t, int32(3), bc.Staged.Next.Value.Content.SeqNumber)
}