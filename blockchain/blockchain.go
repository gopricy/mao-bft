package blockchain

import (
	sha256 "crypto/sha256"
	"errors"
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"google.golang.org/protobuf/proto"
)

type StagedBlock struct {
	Prev *StagedBlock
	Next *StagedBlock
	Value *pb.Block
}

// RemoveStagedBlock removes a staged block from blockchain. Return prev block.
func RemoveStagedBlock(Block *StagedBlock) *StagedBlock {
	prevBlock := Block.Prev
	if Block.Next != nil {
		Block.Next.Prev = Block.Prev
	}
	Block.Prev.Next = Block.Next
	return prevBlock
}

// Insert to list in sorted order.
func OrderedInsertToList(head *StagedBlock, Block pb.Block) bool {
	cur := head
	for cur.Next != nil && cur.Next.Value.Content.SeqNumber < Block.Content.SeqNumber {
		cur = cur.Next
	}
	// Insert after cur node.
	newBlock := StagedBlock{
		Prev: cur,
		Next: cur.Next,
		Value: &Block,
	}
	if cur.Next != nil {
		cur.Next.Prev = &newBlock
	}
	cur.Next = &newBlock
	return true
}

func GetStagedAreaSize(head StagedBlock) int {
	cur := head.Next
	count := 0
	for cur != nil {
		count += 1
		cur = cur.Next
	}
	return count
}

func GetLastBlockInStagedArea(head StagedBlock) (StagedBlock, error) {
	if head.Next == nil {
		return StagedBlock{}, errors.New("Not staged block in staged area.")
	}
	cur := head.Next
	for cur.Next != nil {
		cur = cur.Next
	}
	return *cur, nil
}

type Blockchain struct {
	// This structure stores the blocks in chain.
	Chain []*pb.Block
	// This list stores sorted blocks that are not committed. All blocks are sorted by sequence number.
	Staged StagedBlock
}

func (bc *Blockchain) Init() {
	// Add a sentinel node for both staged and blockchain.
	bc.Chain = append(bc.Chain, &pb.Block{})
	bc.Staged = StagedBlock{}
}

func (bc *Blockchain) GetLastBlock() *pb.Block {
	return bc.Chain[len(bc.Chain) - 1]
}

func (bc *Blockchain) GetLastIndex() int {
	return len(bc.Chain) - 1
}

// Add to staged area in sorted order.
func (bc *Blockchain) addToStagedArea(Block pb.Block) bool {
	return OrderedInsertToList(&bc.Staged, Block)
}

// ValidateBlock validates that a block's cur_hash matches the actual hash.
func ValidateBlock(Block pb.Block) bool {
	h := sha256.New()
	byteContent, _ := proto.Marshal(Block.Content)
	if _, err := h.Write(byteContent); err != nil {
		return false
	}
	return mao_utils.IsSameBytes(h.Sum(nil), Block.CurHash)
}

// CommitBlock tries to apply a single block to block chain, return all blocks get applied.
// Note that, there could be multiple block gets applied in one shot.
func (bc *Blockchain) CommitBlock(Block pb.Block) ([]*pb.Block, []*pb.Block, error) {
	// 0. Validate block.
	if isValid := ValidateBlock(Block); isValid == false {
		return nil, nil, errors.New("The block is not valid.")
	}
	// 1. Add the block to staged area in order.
	bc.addToStagedArea(Block)

	var committed []*pb.Block
	var deleted []*pb.Block
	// 2. try commit to chain if it matches the last block.
	if mao_utils.IsSameBytes(Block.Content.PrevHash, bc.GetLastBlock().CurHash) {
		// Scan staging area and 1. Remove all invalid. 2. commit all can be connected.
		cur := bc.Staged.Next
		for cur != nil && cur.Next != nil {
			if cur.Value.Content.SeqNumber > int32(bc.GetLastIndex() + 1) {
				break
			}

			if cur.Value.Content.SeqNumber <= int32(bc.GetLastIndex()) {
				// The staged area conflict with chain, thus remove.
				deleted = append(deleted, cur.Value)
			 	cur = RemoveStagedBlock(cur)
			} else if cur.Value.Content.SeqNumber == int32(bc.GetLastIndex() + 1) {
				// Try to commit if hash matches.
				if mao_utils.IsSameBytes(cur.Value.Content.PrevHash, bc.GetLastBlock().CurHash) {
					committed = append(committed, cur.Value)
					bc.Chain = append(bc.Chain, cur.Value)
				}
				cur = RemoveStagedBlock(cur)
			}
			cur = cur.Next
		}
	}
	return committed, deleted, nil
}
