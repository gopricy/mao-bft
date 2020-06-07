package blockchain

import (
	"encoding/hex"
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func constructDepositTransaction(txUuid string, amount int, userId string) *pb.Transaction {
	return &pb.Transaction{
		TransactionUuid: txUuid,
		Message: &pb.Transaction_DepositMsg{
			DepositMsg: &pb.DepositMessage{
				Amount: int32(amount),
				AccountId: userId,
			},
		},
	}
}

func constructWireTransaction(txUuid string, amount int, from string, to string) *pb.Transaction {
	return &pb.Transaction{
		TransactionUuid: txUuid,
		Message: &pb.Transaction_WireMsg{
			WireMsg: &pb.WireMessage{
				Amount: int32(amount),
				FromId: from,
				ToId: to,
			},
		},
	}
}

// Create a sample non-persistent blockchain that contains a single block in each of staged/committed/pending area.
func getSampleBlockchain() *Blockchain {
	bc := NewBlockchain("")
	block, err := mao_utils.CreateBlockFromTxsAndPrevHash(
		[]*pb.Transaction{
			constructDepositTransaction("1", 10, "user1"),
			constructDepositTransaction("2", 15, "user2"),
		},
		mao_utils.GetLastBlockFromArray(bc.Chain).CurHash)
	if err != nil {
		panic("Fail to construct block.")
	}
	bc.Chain = append(bc.Chain, block)

	// Create 2 pending block.
	pending1, err := mao_utils.CreateBlockFromTxsAndPrevHash(
		[]*pb.Transaction{
			constructWireTransaction("3", 10, "user2", "user1"),
		},
		mao_utils.GetLastBlockFromArray(bc.Chain).CurHash)
	pending2, err := mao_utils.CreateBlockFromTxsAndPrevHash(
		[]*pb.Transaction{
			constructWireTransaction("4", 5, "user1", "user2"),
		},
		pending1.CurHash)
	bc.Pending.PushBack(pending1)
	bc.Pending.PushBack(pending2)
	// Set Pending 2 as staged block.
	bc.Staged[hex.EncodeToString(pending2.Content.PrevHash)] = pending2

	// Setup tx status.
	bc.TxStatus["1"] = pb.TransactionStatus_COMMITTED
	bc.TxStatus["2"] = pb.TransactionStatus_COMMITTED
	bc.TxStatus["3"] = pb.TransactionStatus_PENDING
	bc.TxStatus["4"] = pb.TransactionStatus_STAGED

	return bc
}

func TestBlockchain_Init(t *testing.T) {
	bc := NewBlockchain("")
	assert.NotNil(t, bc.Pending)
	assert.Equal(t, len(bc.Chain), 1)
	assert.True(t, mao_utils.IsSameBytes(bc.Chain[0].CurHash, []byte{0}))
	assert.NotNil(t, bc.Staged)
	assert.NotNil(t, bc.TxStatus)
}

func TestBlockchain_CommitBlock_CommitSingleBlock(t *testing.T) {
	bc := NewBlockchain("")

	block, err := mao_utils.CreateBlockFromTxsAndPrevHash(
		[]*pb.Transaction{
			constructWireTransaction("3", 10, "user2", "user1"),
		},
		mao_utils.GetLastBlockFromArray(bc.Chain).CurHash)
	assert.Nil(t, err)
	committedBlocks, _, err := bc.CommitBlock(block)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(committedBlocks))
	assert.Equal(t, bc.Pending.Len(), 0)
	assert.Equal(t, len(bc.Staged), 0)
	assert.Equal(t, len(bc.Chain), 2)
	assert.Equal(t, bc.Chain[1].Content.Txs[0].TransactionUuid, "3")
	assert.Equal(t, len(bc.TxStatus), 1)
	assert.Equal(t, bc.TxStatus["3"], pb.TransactionStatus_COMMITTED)
}

func TestBlockchain_CommitBlock_Commit2Block(t *testing.T) {
	bc := getSampleBlockchain()
	candidate := bc.Pending.Front().Value.(*pb.Block)
	committed, _, err := bc.CommitBlock(candidate)
	assert.Nil(t, err)
	assert.Equal(t, len(committed), 2)
	assert.Equal(t, committed[0].Content.Txs[0].TransactionUuid, "3")
	assert.Equal(t, committed[1].Content.Txs[0].TransactionUuid, "4")
	// Test status
	assert.Equal(t, len(bc.TxStatus), 4)
	for _, key := range []string{"1", "2", "3", "4"} {
		assert.Equal(t, bc.TxStatus[key], pb.TransactionStatus_COMMITTED)
	}
}

func TestBlockchain_CommitBlock_StageOnly(t *testing.T) {
	bc := getSampleBlockchain()
	block, err := mao_utils.CreateBlockFromTxsAndPrevHash(
		[]*pb.Transaction{
			constructWireTransaction("5", 10, "user2", "user1"),
		},
		[]byte{1})
	assert.Nil(t, err)
	blocks, _, err := bc.CommitBlock(block)
	assert.Nil(t, err)
	assert.Equal(t, len(blocks), 0)
	assert.Equal(t, len(bc.Staged), 2)
}

func TestBlockchain_CommitBlock_IdempotentCommit(t *testing.T) {
	bc := getSampleBlockchain()
	block, err := mao_utils.CreateBlockFromTxsAndPrevHash(
		[]*pb.Transaction{
			constructWireTransaction("5", 10, "user2", "user1"),
		},
		[]byte{1})
	assert.Nil(t, err)
	blocks, _, err := bc.CommitBlock(block)
	assert.Nil(t, err)
	assert.Equal(t, len(blocks), 0)
	assert.Equal(t, len(bc.Staged), 2)

	// Commit again.
	blocks, _, err = bc.CommitBlock(block)
	assert.Nil(t, err)
	assert.Equal(t, len(blocks), 0)
	assert.Equal(t, len(bc.Staged), 2)
}

func TestBlockchain_CreateNewPendingBlock(t *testing.T) {
	bc := getSampleBlockchain()
	block, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructWireTransaction("5", 1, "user1", "user2"),
		constructWireTransaction("6", 1, "user1", "user2"),
	})
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsValidBlockHash(block))
	assert.Equal(t, bc.Pending.Len(), 3)
	assert.Equal(t, len(bc.Staged), 1)
	assert.Equal(t, len(bc.Chain), 2)
	assert.Equal(t, bc.TxStatus["5"], pb.TransactionStatus_PENDING)
	assert.Equal(t, bc.TxStatus["6"], pb.TransactionStatus_PENDING)
}

// Do some fancy operation and test status.
func TestBlockchain_GetTransactionStatus(t *testing.T) {
	bc := NewBlockchain("")
	// Get non-exist.
	assert.Equal(t, bc.GetTransactionStatus("3"), pb.TransactionStatus_REJECTED)

	pending1, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("1", 10, "user1")})
	assert.Nil(t, err)
	assert.Equal(t, len(bc.TxStatus), 1)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_PENDING)

	pending2, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("2", 10, "user2")})
	assert.Nil(t, err)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_PENDING)

	// Commit 2, this should make pending 2 to staged.
	committed, _, err := bc.CommitBlock(pending2)
	assert.Nil(t, err)
	assert.Equal(t, len(committed), 0)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_STAGED)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_PENDING)

	// Commit 1, this should make everything committed.
	committed, _, err = bc.CommitBlock(pending1)
	assert.Nil(t, err)
	assert.Equal(t, len(committed), 2)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_COMMITTED)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_COMMITTED)
}

func TestBlockchain_GetAllTxInOrder(t *testing.T) {
	bc := getSampleBlockchain()
	blocks, isCommit := bc.GetAllBlocksInOrder()
	assert.Equal(t, len(blocks), 4)
	assert.Equal(t, len(isCommit), 4)
}

// This test tests that blockchain failover can reconstruct original state.
func TestBlockchain_Reconcile(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	// Construct a blockchain with persistent storage.
	bc := NewBlockchain(tmpDir)
	// Get non-exist.
	assert.Equal(t, bc.GetTransactionStatus("3"), pb.TransactionStatus_REJECTED)

	pending1, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("1", 10, "user1")})
	assert.Nil(t, err)
	assert.Equal(t, len(bc.TxStatus), 1)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_PENDING)

	pending2, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("2", 10, "user2")})
	assert.Nil(t, err)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_PENDING)

	// Commit 2, this should make pending 2 to staged.
	committed, _, err := bc.CommitBlock(pending2)
	assert.Nil(t, err)
	assert.Equal(t, len(committed), 0)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_STAGED)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_PENDING)

	// Now failover
	bc = nil
	bc = NewBlockchain(tmpDir)

	// Commit 1, this should make everything committed.
	committed, _, err = bc.CommitBlock(pending1)
	assert.Nil(t, err)
	assert.Equal(t, len(committed), 2)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_COMMITTED)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_COMMITTED)

	// Clean up.
	os.Remove(tmpDir)
}

func TestBlockchain_GetLastStagedAndCommitBlock(t *testing.T) {
	bc := NewBlockchain("")
	lastCommitBlock, err := mao_utils.DecodeBlock(bc.GetLastCommittedBytes())

	// Empty blockchain's last commit should be blockchain head.
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsSameBytes(lastCommitBlock.CurHash, []byte{0}))

	pending1, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("1", 10, "user1")})
	pending2, err := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("2", 10, "user2")})

	// Pending chain doesn't affect last commit.
	lastCommitBlock, err = mao_utils.DecodeBlock(bc.GetLastCommittedBytes())
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsSameBytes(lastCommitBlock.CurHash, []byte{0}))

	// Commit 2, this should make pending 2 to staged.
	committed, _, err := bc.CommitBlock(pending2)
	assert.Equal(t, len(committed), 0)
	lastCommitBlock, err = mao_utils.DecodeBlock(bc.GetLastCommittedBytes())
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsSameBytes(lastCommitBlock.CurHash, []byte{0}))
	lastStagedBlock, err := mao_utils.DecodeBlock(bc.GetLastStagedBlock())
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsSameBytes(lastStagedBlock.CurHash, pending2.CurHash))

	// Commit 2, this should make all committed
	committed, _, err = bc.CommitBlock(pending1)
	assert.Equal(t, len(committed), 2)
	lastCommitBlock, err = mao_utils.DecodeBlock(bc.GetLastCommittedBytes())
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsSameBytes(lastCommitBlock.CurHash, pending2.CurHash))
	lastStagedBlock, err = mao_utils.DecodeBlock(bc.GetLastStagedBlock())
	assert.Nil(t, err)
	assert.True(t, mao_utils.IsSameBytes(lastStagedBlock.CurHash, pending1.CurHash))
}

func TestBlockchain_GetAnswerForSyncRequest(t *testing.T) {
	bc := NewBlockchain("")
	pending1, _ := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("1", 10, "user1")})
	pending2, _ := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("2", 10, "user2")})
	// Commit 1, 2.
	_, _, _ = bc.CommitBlock(pending2)
	_, _, _ = bc.CommitBlock(pending1)

	headBytes, err := mao_utils.EncodeBlock(bc.Chain[0])
	assert.Nil(t, err)
	tailBytes := bc.GetLastCommittedBytes()
	// Get answer with head & pending 2.
	answerBlocks := bc.GetAnswerForSyncRequest(headBytes, tailBytes)
	assert.Equal(t, len(answerBlocks), 1)
	assert.True(t, mao_utils.IsSameBlock(answerBlocks[0], pending1))
}

func TestBlockchain_GetAnswerForSyncRequest2(t *testing.T) {
	// In this case we test 3 block in a chain.
	bc := NewBlockchain("")
	pending1, _ := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("1", 10, "user1")})
	pending2, _ := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("2", 10, "user2")})
	pending3, _ := bc.CreateNewPendingBlock([]*pb.Transaction{
		constructDepositTransaction("3", 10, "user3")})
	// Commit 1, 2.
	_, _, _ = bc.CommitBlock(pending1)
	_, _, _ = bc.CommitBlock(pending2)
	_, _, _ = bc.CommitBlock(pending3)

	headBytes, err := mao_utils.EncodeBlock(bc.Chain[0])
	assert.Nil(t, err)
	tailBytes := bc.GetLastCommittedBytes()
	// Get answer with head & pending 3.
	answerBlocks := bc.GetAnswerForSyncRequest(headBytes, tailBytes)
	assert.Equal(t, len(answerBlocks), 2)
	assert.True(t, mao_utils.IsSameBlock(answerBlocks[0], pending1))
	assert.True(t, mao_utils.IsSameBlock(answerBlocks[1], pending2))
}
