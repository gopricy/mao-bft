package transaction

import (
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/blockchain"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/follower"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"github.com/pkg/errors"
	"sync"
)

const MaximumTxn = 1000000

// Application manages blockchain & transaction, it also maintains internal App specific data structure for state machine.
type Application interface {
	// Once a message is RBC'ed, this function will be called to apply this block.
	// This function should be thread safe.
	RBCReceive(bytes []byte) (bool, error)

	GetSyncQuestion() (pb.SyncRequest, error)
	GetSyncAnswer(request pb.SyncRequest) (pb.SyncResponse, error)

	// Get status of a transaction by its uuid.
	GetTransactionStatus(txUuid string) pb.TransactionStatus
	// TODO(chenweilunster): Add validation functionality
}

type RBCLeader interface{
	RBCSend(bytes []byte)
}

type common struct{
	Queue *EventQueue
	Blockchain *blockchain.Blockchain
	// Ledger stores the committed states of transactions.
	Ledger *Ledger
	// PendingLedger stores the ledger after applying all Tx in event queue.
	PendingLedger *Ledger
	mu sync.Mutex
}

func newcommon(dir string) *common{
	res := new(common)
	res.Queue = new(EventQueue)
	res.Ledger = NewLedger()
	res.PendingLedger = NewLedger()
	res.Blockchain = blockchain.NewBlockchain(dir)
	blocks, isCommit := res.Blockchain.GetAllBlocksInOrder()
	res.Ledger.Reconcile(blocks, isCommit, true)
	res.PendingLedger.Reconcile(blocks, isCommit, false)

	return res
}

var _ Application = &common{}

func (c *common) GetSyncQuestion() (pb.SyncRequest, error) {
	return pb.SyncRequest{
		LastCommitHash: c.Blockchain.GetLastCommittedHash(),
		LatestStagedHash: c.Blockchain.GetLastStagedBlockHash(),
	}, nil
}

func (c *common) GetSyncAnswer(request pb.SyncRequest) (pb.SyncResponse, error) {
	return pb.SyncResponse{}, nil
}

func (c *common) RBCReceive(bytes []byte) (bool, error) {
	block, err := mao_utils.DecodeBlock(bytes)
	if err != nil{
		return false, errors.Wrap(err, "Can't decode Block")
	}

	// Below is critical section that only one thread can enter at the same time.
	c.mu.Lock()
	defer c.mu.Unlock()
	blocks, shouldSync, err := c.Blockchain.CommitBlock(block)
	if err != nil {
		return false, err
	}
	for _, b := range blocks{
		for _, t := range b.Content.Txs{
			if err := c.Ledger.CommitTxn(t); err != nil{
				return false, err
			}
		}
	}
	return shouldSync, nil
}

func (c *common) GetTransactionStatus(txUuid string) pb.TransactionStatus {
	if c.Queue.Exist(txUuid){
		return pb.TransactionStatus_UNKNOWN
	}
	return c.Blockchain.TxStatus[txUuid]
}

type Leader struct{
	Leader RBCLeader
	*common
	MaxBlockSize int
	mu sync.Mutex
}

func (l *Leader) SetRBCLeader(leader RBCLeader){
	l.Leader = leader
}

func NewLeader(blocksize int, dir string) *Leader{
	res := new(Leader)
	res.common = newcommon(dir)
	res.MaxBlockSize = blocksize
	return res
}

// TODO: expose this API in a binary.
func (l *Leader) ProposeTransfer(from, to string, dollar, cents int) (string, error){
	if dollar > MaximumTxn || cents >= 100 || cents < 0{
		return "", errors.New("invalid amount, transaction limit is 1M")
	}
	txn := &pb.Transaction{
		Message: &pb.Transaction_WireMsg{
			WireMsg: &pb.WireMessage{
				FromId: from,
				ToId: to,
				Amount: int32(dollar * 100 + cents),
			},
		},
	}
	if !l.PendingLedger.ValidateTransaction(txn) {
		return "", errors.New("Invalid transaction: " + proto.MarshalTextString(txn))
	}
	u, t, err := l.Queue.AddTxToEventQueue(txn, l.PendingLedger)
	if err != nil{
		return "", err
	}
	if t == l.MaxBlockSize{
		if err := l.createBlockAndSend(); err != nil{
			return "", err
		}
	}
	return u, nil
}

func (l *Leader) ProposeDeposit(id string, dollar, cents int) (string, error){
	if dollar > MaximumTxn || cents >= 100 || cents < 0{
		return "", errors.New("invalid amount, transaction limit is 1M")
	}
	txn := &pb.Transaction{
		Message: &pb.Transaction_DepositMsg{
			&pb.DepositMessage{
				AccountId: id,
				Amount: int32(dollar * 100 + cents),
			},
		},
	}
	u, t, err := l.Queue.AddTxToEventQueue(txn, l.PendingLedger)
	if err != nil{
		return "", err
	}
	if t == l.MaxBlockSize{
		if err := l.createBlockAndSend(); err != nil{
			return "", err
		}
	}
	return u, nil
}

// This function is critical section that permits only single entry.
func (l *Leader) createBlockAndSend() error{
	l.mu.Lock()
	defer l.mu.Unlock()

	txs, err := l.Queue.GetTransactions(l.MaxBlockSize)
	if err != nil{
		return err
	}
	block, err := l.Blockchain.CreateNewPendingBlock(txs)
	if err != nil{
		return err
	}
	enc, err := mao_utils.EncodeBlock(block)
	if err != nil{
		return err
	}
	l.Leader.RBCSend(enc)
	return nil
}

type Follower struct{
	Follower follower.Follower
	*common
}

func NewFollower(dir string) *Follower{
	res := new(Follower)
	res.common = newcommon(dir)
	return res
}
