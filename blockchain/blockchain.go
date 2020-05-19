package blockchain

import (
	sha256 "crypto/sha256"
	"errors"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"google.golang.org/protobuf/proto"
)

type StagedBlock struct {
	Prev *StagedBlock
	Next *StagedBlock
	Value *pb.Block
}

// RemoveStagedBlock removes a staged block from blockchain. Return next block.
func RemoveStagedBlock(Block *StagedBlock) *StagedBlock {
	nextBlock := Block.Next
	if Block.Next != nil {
		Block.Next.Prev = Block.Prev
	}
	Block.Prev.Next = Block.Next
	return nextBlock
}

// Insert to list in sorted order.
func OrderedInsertToList(head StagedBlock, Block pb.Block) bool {
	cur := &head
	for cur.Next != nil && cur.Next.Value.Content.SeqNumber < head.Value.Content.SeqNumber {
		cur = cur.Next
	}
	// Insert after cur node.
	newBlock := StagedBlock{
		Prev: cur,
		Next: cur.Next,
		Value: &Block,
	}
	cur.Next = &newBlock
	if cur.Next != nil {
		cur.Next.Prev = &newBlock
	}
	return true
}

type Blockchain struct {
	// This structure stores the blocks in chain.
	Chain []*pb.Block
	// This list stores sorted blocks that are not committed. All blocks are sorted by sequence number.
	Staged StagedBlock
}

func (b *Blockchain) RBCReceive([]byte) error{
	return nil
}

type BlockchainFollower struct {
	Blockchain
	follower.Follower
}

func NewBlockchainFollower(name string) BlockchainFollower{
	res := BlockchainFollower{}
	res.Follower = follower.NewFollower(name, &res)
	return res
}

type BlockchainLeader struct{
	Blockchain
	leader.Leader
}

// Client will call this
func (bl *BlockchainLeader) Send(p *pb.Block) error{
	b, _ := proto.Marshal(p)
	return bl.RBCSend(b)
}

func NewBlockchainLeader(name string) BlockchainLeader{
	res := BlockchainLeader{}
	res.Leader = leader.NewLeader(name, &res)
	return res
}

func (bc *Blockchain) Init() {
	// Add a sentinel node for both staged and blockchain.
	bc.Chain = append(bc.Chain, &pb.Block{})
	bc.Staged = StagedBlock{}
}

func (bc *Blockchain) GetLastBlock() *pb.Block {
	return bc.Chain[len(bc.Chain) - 1]
}

// Add to staged area in sorted order.
func (bc *Blockchain) addToStagedArea(Block pb.Block) bool {
	return OrderedInsertToList(bc.Staged, Block)
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
	// 2. try commit to chain if it matches the last block.
	if mao_utils.IsSameBytes(Block.Content.PrevHash, bc.GetLastBlock().CurHash) {
		// Scan staging area and 1. Remove all invalid. 2. commit all can be connected.
		cur := bc.Staged
		for cur.Next != nil {
			// if cur.Value.Content.SeqNumber < len(bc.Chain) - 1 {
			// 	RemoveStagedBlock
			// }
		}
	}
	return nil, nil, nil
}
