package blockchain

import (
	"container/list"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"strconv"
	"sync"
)

type Blockchain struct {
	// This structure stores the blocks in chain.
	Chain []*pb.Block
	// This list stores sorted blocks that are not committed. All blocks are sorted by sequence number.
	Staged map[string]*pb.Block
	// This list stores pending blocks that is/will be broadcast. This should be concatenated to Chain.
	Pending *list.List
	// A mapping from transaction status to its status.
	TxStatus map[string]pb.TransactionStatus
	// Access to blockchain should always be thread-safe.
	Mu sync.RWMutex
}

// Init internal data structure.
func (bc *Blockchain) Init() {
	bc.Pending = list.New()
	// The first block in blockchain should have hash of 0.
	bc.Chain = append(bc.Chain, &pb.Block{CurHash: []byte{0}})
	bc.TxStatus = make(map[string]pb.TransactionStatus)
	bc.Staged = make(map[string]*pb.Block)
}

// Add block to staged area, key to it's previous block's CurHash.
func (bc *Blockchain) addToStagedArea(block *pb.Block, overwrite bool) error {
	hexHash := hex.EncodeToString(block.Content.PrevHash)
	if exist, ok := bc.Staged[hexHash]; ok {
		if !mao_utils.IsSameBytes(exist.CurHash, block.CurHash) {
			fmt.Println()
		}
	}
	bc.Staged[hexHash] = block
	if err := bc.setTxsStatus(block.Content.Txs, pb.TransactionStatus_STAGED, overwrite); err != nil {
		return err
	}
	return nil
}

// Returns whether a blockchain has uncommitted (by ready to commit) blocks in staged area.
func (bc *Blockchain) dirty() (bool, *pb.Block) {
	lastCommit := mao_utils.GetLastBlockFromArray(bc.Chain)
	lastCommitHash := hex.EncodeToString(lastCommit.CurHash)
	if staged, ok := bc.Staged[lastCommitHash]; ok {
		return true, staged
	}
	return false, nil
}

// Set a list of transactions as status.
func (bc *Blockchain) setTxsStatus(txs []*pb.Transaction, status pb.TransactionStatus, overwrite bool) error {
	for _, tx := range txs {
		_, ok := bc.TxStatus[tx.TransactionUuid]
		// Either overwrite an existing value, or write for the first time.
		if ok != overwrite {
			return errors.New("Transaction status doesn't match overwrite specification" + strconv.FormatBool(overwrite))
		}
		bc.TxStatus[tx.TransactionUuid] = status
	}
	return nil
}

// CommitBlock tries to apply a single block to block chain, return all blocks get applied, removed,
// and whether the input block is committed.
// - input
// Block: The block that we're trying to commit.
// - output
// 1. Successfully committed new blocks. Empty if nothing gets committed.
// 2. Error
// This function is thread safe.
func (bc *Blockchain) CommitBlock(block *pb.Block) ([]*pb.Block, error) {
	bc.Mu.Lock()
	defer bc.Mu.Unlock()

	// 0. Validate block.
	if isValid := mao_utils.IsValidBlockHash(block); isValid == false {
		return nil, errors.New("The block is invalid.")
	}
	// 1. Add the block to staged area in order by sequence number.
	isLeader := bc.Pending.Len() != 0
	if err := bc.addToStagedArea(block, isLeader); err != nil {
		return nil, nil
	}

	var committed []*pb.Block
	// 2. Scan staged area, try to commit if it's dirty.
	for isDirty, candidate := bc.dirty(); isDirty; isDirty, candidate = bc.dirty() {
		// a. Append to Chain.
		bc.Chain = append(bc.Chain, candidate)
		committed = append(committed, candidate)
		if err := bc.setTxsStatus(candidate.Content.Txs, pb.TransactionStatus_COMMITTED, true); err != nil {
			return nil, err
		}

		// b. Remove from pending if it has. Note that, only leader contains pending section.
		if bc.Pending.Len() != 0 {
			iter := bc.Pending.Front()
			nextPending := iter.Value.(*pb.Block)
			if !mao_utils.IsSameBlock(nextPending, candidate) {
				return nil, errors.New("Candidate block must be the head of pending.")
			}
			bc.Pending.Remove(iter)
		}

		// c. Remove from staged.
		delete(bc.Staged, hex.EncodeToString(candidate.Content.PrevHash))
	}

	return committed, nil
}

// CreateNewPendingBlock creates a block at pending chain. Append the block to pending chain and returns.
// This function is thread safe.
func (bc *Blockchain) CreateNewPendingBlock(txs []*pb.Transaction) (*pb.Block, error) {
	bc.Mu.Lock()
	defer bc.Mu.Unlock()

	lastBlock := mao_utils.GetLastBlockFromArray(bc.Chain)
	if bc.Pending.Len() != 0 {
		lastBlock = bc.Pending.Back().Value.(*pb.Block)
	}
	newBlock, err := mao_utils.CreateBlockFromTxsAndPrevHash(txs, lastBlock.CurHash)
	if err != nil {
		return nil, err
	}
	bc.Pending.PushBack(newBlock)
	// Assign all TX as status PENDING.
	if err := bc.setTxsStatus(txs, pb.TransactionStatus_PENDING, false); err != nil {
		return nil, err
	}
	return newBlock, nil
}

// Returns the status of a transaction, REJECT if the transaction is not found in chain.
// This function is thread safe.
func (bc *Blockchain) GetTransactionStatus(txUuid string) pb.TransactionStatus {
	bc.Mu.RLock()
	defer bc.Mu.RUnlock()

	if status, ok := bc.TxStatus[txUuid]; ok {
		return status
	}
	return pb.TransactionStatus_REJECTED
}
