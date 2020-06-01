package transaction

import (
	"github.com/gopricy/mao-bft/blockchain"
	"github.com/gopricy/mao-bft/pb"
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

func TestCommon_InitNonPersistentCommonWillSucceed(t *testing.T) {
	common := newcommon("")
	assert.NotNil(t, common)
}

func TestCommon_InitWithEmptyStorage(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)
	common := newcommon(tmpDir)
	assert.NotNil(t, common)
	os.Remove(tmpDir)
}

func TestCommon_InitWithNonEmptyStorage(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "*")
	assert.Nil(t, err)

	// Construct a blockchain with persistent storage.
	bc := blockchain.NewBlockchain(tmpDir)
	// Get non-exist.
	assert.Equal(t, bc.GetTransactionStatus("3"), pb.TransactionStatus_REJECTED)

	_, err = bc.CreateNewPendingBlock([]*pb.Transaction{
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
	committed, err := bc.CommitBlock(pending2)
	assert.Nil(t, err)
	assert.Equal(t, len(committed), 0)
	assert.Equal(t, len(bc.TxStatus), 2)
	assert.Equal(t, bc.GetTransactionStatus("2"), pb.TransactionStatus_STAGED)
	assert.Equal(t, bc.GetTransactionStatus("1"), pb.TransactionStatus_PENDING)

	// Now failover
	bc = nil

	// Cold start new common.
	common := newcommon(tmpDir)
	assert.NotNil(t, common)
	assert.Equal(t, len(common.Ledger.Accounts), 0)
	assert.Equal(t, len(common.PendingLedger.Accounts), 2)
	assert.Equal(t, common.PendingLedger.Accounts["user1"], int32(10))
	assert.Equal(t, common.PendingLedger.Accounts["user2"], int32(10))

	os.Remove(tmpDir)
}