package blockchain

import (
	"container/list"
	"encoding/hex"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"log"
	"strconv"
	"sync"
)

// if staged is more than this size, a sync should be triggered.
const MaxStagedBuffer = 3

type Blockchain struct {
	// This structure stores the blocks in chain.
	Chain []*pb.Block
	// This list stores sorted blocks that are not committed, the mapping is <prevHash -> CurBlock>.
	Staged map[string]*pb.Block
	// This list stores pending blocks that is/will be broadcast. This should be concatenated to Chain.
	Pending *list.List
	// Latest staged seen. This is used for sync.
	LastStaged *pb.Block
	// A mapping from transaction status to its status.
	TxStatus map[string]pb.TransactionStatus
	// Access to blockchain should always be thread-safe.
	Mu sync.RWMutex
	// This flag determines whether a blockchain should be persistent (on disk).
	persistent bool
	// This is the path that blockchain stores persistent states.
	path string
	// This is the logger that blockchain will use to maintain a persistent storage.
	logger *Logger
}

// NewBlockchain takes in path as parameter, it will return a blockchain with initial state constructed from path.
// If path is empty string, blockchain
func NewBlockchain(path string) *Blockchain{
	res := new(Blockchain)
	res.Pending = list.New()
	// The first block in blockchain should have hash of 0 and content nil.
	res.Chain = append(res.Chain, &pb.Block{CurHash: []byte{0}})
	res.TxStatus = make(map[string]pb.TransactionStatus)
	res.Staged = make(map[string]*pb.Block)
	res.path = path
	// If path is empty, this blockchain is non-persistent, this is usually used for testing.
	res.persistent = path != ""

	if res.persistent {
		// initialize logger and reconcile existing persistent storage.
		res.logger = NewLogger(res.path)
		res.Reconcile()
	}

	return res
}

// Reconcile replicate blockchain to be same as state stored in persistent storage.
func (bc *Blockchain) Reconcile() {
	blockMap := make(map[pb.BlockState]map[string]*pb.Block)
	blockDumps, err := bc.logger.ReadAllBlocks()
	if err != nil {
		log.Fatalln("Fail to reconcile with blocks stored in: " + bc.path)
	}
	if len(blockDumps) == 0 {
		return
	}

	// 4 internal states to reconstruct from persistent storage.
	var chain []*pb.Block
	txStatus := make(map[string]pb.TransactionStatus)
	staged := make(map[string]*pb.Block)
	pending := list.New()
	chain = append(chain, &pb.Block{CurHash: []byte{0}})
	
	for _, dump := range blockDumps {
		switch dump.State {
		case pb.BlockState_BS_COMMITTED:
			fallthrough
		case pb.BlockState_BS_PENDING:
			fallthrough
		case pb.BlockState_BS_STAGED:
			if blockMap[dump.State] == nil {
				blockMap[dump.State] = make(map[string]*pb.Block)
			}
			stateMap := blockMap[dump.State]
			stateMap[hex.EncodeToString(dump.Block.Content.PrevHash)] = dump.Block
			break
		default:
			log.Fatalln("Unknown kind of block: " + proto.MarshalTextString(&dump))
		}
	}

	// Construct chain.
	committedMap := blockMap[pb.BlockState_BS_COMMITTED]
	// Observed committed block.
	committedSet := make(map[string]bool)
	tail := hex.EncodeToString(mao_utils.GetLastBlockFromArray(chain).CurHash)
	for block, contains := committedMap[tail]; contains; block, contains = committedMap[tail]{
		if len(chain) >= len(committedMap) + 1 {
			log.Fatalln("Created more committed blocks than what's stored in persistent storage.")
		}
		chain = append(chain, block)
		committedSet[hex.EncodeToString(block.CurHash)] = true
		tail = hex.EncodeToString(block.CurHash)
	}
	if len(chain) != len(committedMap) + 1 { // +1 because chain contains a head.
		log.Fatalln("There are leftover committed in persistent storage.")
	}

	// Construct pending. Reuse tail constructed above.
	pendingMap := blockMap[pb.BlockState_BS_PENDING]
	for block, contains := pendingMap[tail]; contains; block, contains = pendingMap[tail]{
		if pending.Len() >= len(pendingMap) {
			log.Fatalln("Created more pending blocks than what's stored in persistent storage.")
		}
		pending.PushBack(block)
		tail = hex.EncodeToString(block.CurHash)
	}

	// Construct staged.
	stagedMap := blockMap[pb.BlockState_BS_STAGED]
	for _, block := range stagedMap {
		// Only add block to staged if it's not in committed set.
		if _, exist := committedSet[hex.EncodeToString(block.CurHash)]; !exist {
			staged[hex.EncodeToString(block.Content.PrevHash)] = block
		}
	}

	// Reconstruct TX status.
	for iter := pending.Front(); iter != nil; iter = iter.Next() {
		block := iter.Value.(*pb.Block)
		if block.Content == nil {
			continue
		}
		for _, tx := range block.Content.Txs {
			txStatus[tx.TransactionUuid] = pb.TransactionStatus_PENDING
		}
	}
	for _, block := range staged {
		if block.Content == nil {
			continue
		}
		for _, tx := range block.Content.Txs {
			txStatus[tx.TransactionUuid] = pb.TransactionStatus_STAGED
		}
	}
	for _, block := range chain {
		if block.Content == nil {
			continue
		}
		for _, tx := range block.Content.Txs {
			txStatus[tx.TransactionUuid] = pb.TransactionStatus_COMMITTED
		}
	}

	bc.Chain = chain
	bc.Staged = staged
	bc.Pending = pending
	bc.TxStatus = txStatus
}

// Add block to staged area, key to it's previous block's CurHash.
func (bc *Blockchain) addToStagedArea(block *pb.Block, overwrite bool) error {
	// First mark a block seen as latest staged.
	bc.LastStaged = block

	hexHash := hex.EncodeToString(block.Content.PrevHash)

	// Write before apply.
	if bc.persistent {
		bc.logger.WriteBlock(*block, pb.BlockState_BS_STAGED)
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
// 2. Length of staged area.
// 3. Error
// This function is thread safe.
func (bc *Blockchain) CommitBlock(block *pb.Block) ([]*pb.Block, bool, error) {
	bc.Mu.Lock()
	defer bc.Mu.Unlock()

	// 0. Validate block:
	// a. Block should have valid hash.
	if isValid := mao_utils.IsValidBlockHash(block); isValid == false {
		return nil, false, errors.New("The block is invalid.")
	}
	// b. Skip block if already in chain
	if bc.IsBlockAlreadyInChain(block) {
		return nil, false, nil
	}

	// 1. Add the block to staged area in order by sequence number.
	isLeader := bc.Pending.Len() != 0
	if err := bc.addToStagedArea(block, isLeader); err != nil {
		return nil, false, nil
	}

	var committed []*pb.Block
	// 2. Scan staged area, try to commit if it's dirty.
	for isDirty, candidate := bc.dirty(); isDirty; isDirty, candidate = bc.dirty() {
		// a. Append to Chain.

		// Write before apply.
		if bc.persistent {
			bc.logger.WriteBlock(*candidate, pb.BlockState_BS_COMMITTED)
		}
		bc.Chain = append(bc.Chain, candidate)



		committed = append(committed, candidate)
		// Set transaction status only of block is valid.
		if err := bc.setTxsStatus(candidate.Content.Txs, pb.TransactionStatus_COMMITTED, true); err != nil {
				return nil, false, err
			}

		// b. Remove from pending if it has. Note that, only leader contains pending section.
		if bc.Pending.Len() != 0 {
			iter := bc.Pending.Front()
			nextPending := iter.Value.(*pb.Block)
			if !mao_utils.IsSameBlock(nextPending, candidate) {
				return nil, false, errors.New("Candidate block must be the head of pending.")
			}
			bc.Pending.Remove(iter)
		}

		// c. Remove from staged.
		delete(bc.Staged, hex.EncodeToString(candidate.Content.PrevHash))
	}

	return committed, len(bc.Staged) > MaxStagedBuffer, nil
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

	// Write before apply.
	if bc.persistent {
		bc.logger.WriteBlock(*newBlock, pb.BlockState_BS_PENDING)
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

// GetAllBlocksInOrder returns 2 lists,
// the first list is all blocks that is either committed or pending,
// the second list indicates whether block is committed.
func (bc *Blockchain) GetAllBlocksInOrder() ([]*pb.Block, []bool) {
	bc.Mu.RLock()
	defer bc.Mu.RUnlock()

	var allBlocks []*pb.Block
	var isBlockCommitted []bool
	for _, block := range bc.Chain {
		allBlocks = append(allBlocks, block)
		isBlockCommitted = append(isBlockCommitted, true)
	}

	for iter := bc.Pending.Front(); iter != nil; iter = iter.Next() {
		block := iter.Value.(*pb.Block)
		allBlocks = append(allBlocks, block)
		isBlockCommitted = append(isBlockCommitted, false)

	}

	return allBlocks, isBlockCommitted
}

// GetLastCommitted returns the last committed block's bytes representation.
func (bc *Blockchain) GetLastCommittedBytes() []byte {
	bc.Mu.RLock()
	defer bc.Mu.RUnlock()
	bytes, err := mao_utils.EncodeBlock(mao_utils.GetLastBlockFromArray(bc.Chain))
	if err != nil {
		log.Fatalln(err.Error())
	}
	return bytes
}

// GetLastStagedBlock returns the latest staged block's bytes representation.
func (bc *Blockchain) GetLastStagedBlock() []byte {
	bc.Mu.RLock()
	defer bc.Mu.RUnlock()
	if bc.LastStaged == nil {
		log.Fatalln("Staged area is empty while calling GetLastStagedBlockHash")
	}
	bytes, err := mao_utils.EncodeBlock(bc.LastStaged)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return bytes
}

// TODO: Optimize this function to be O(1)
// IsBlockAlreadyInChain returns whether block is in either staged area or committed area.
func (bc *Blockchain) IsBlockAlreadyInChain(block *pb.Block) bool {
	for _, b := range bc.Chain {
		if mao_utils.IsSameBytes(b.CurHash, block.CurHash) {
			return true
		}
	}
	for _, b := range bc.Staged {
		if mao_utils.IsSameBytes(b.CurHash, block.CurHash) {
			return true
		}
	}
	return false
}

// GetBlocksForSyncRequest will read the committed chain and try to answer the question.
// It returns an empty array or nil if not able to answer.
func (bc *Blockchain) GetAnswerForSyncRequest(lastCommit []byte, latestStaged []byte) []*pb.Block {
	foundBegin := false
	lastCommitBlock, err := mao_utils.DecodeBlock(lastCommit)
	if err != nil {
		return nil
	}
	lastStagedBlock, err := mao_utils.DecodeBlock(latestStaged)
	if err != nil {
		return nil
	}
	var res []*pb.Block
	for _, committedBlock := range bc.Chain {
		if mao_utils.IsSameBlock(committedBlock, lastStagedBlock) {
			return res
		}
		if foundBegin {
			res = append(res, committedBlock)
		} else if mao_utils.IsSameBlock(committedBlock, lastCommitBlock) {
				foundBegin = true
		}
	}
	return nil
}