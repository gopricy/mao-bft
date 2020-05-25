package wire

import (
	"container/list"
	"errors"
	"github.com/google/uuid"
	"github.com/gopricy/mao-bft/blockchain"
	"github.com/gopricy/mao-bft/pb"
	"sync"
)

// Application manages blockchain & transaction, it also maintains internal App specific data structure for state machine.
type Application interface {
	// Once a message is RBC'ed, this function will be called to apply this block.
	// This function should be thread safe.
	RBCReceive(bytes []byte) error
	// Get status of a transaction by its uuid.
	GetTransactionStatus(txUuid string) (pb.TransactionStatus, error)
	// TODO(chenweilunster): Add validation functionality
}

// #############################################################
// Define Wire System
// #############################################################

// Wire system manages this state machine. This state machine simply tracks account balance of each person.
type Accounts struct {
	Balance map[string]int
	Mu sync.RWMutex
}

// A event queue is a double sided queue that buffers the client proposed transaction.
type EventQueue struct {
	Q list.List
	Mu sync.RWMutex
}

// Add a transaction to event queue, it assigns a UUID to input transaction.
// This queue is managed by TransactionService.
func (q *EventQueue) AddTxToEventQueue(tx pb.Transaction) (string, int, error) {
	if tx.TransactionUuid != "" {
		return "", 0, errors.New("uuid can not be set by client.")
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return "", 0, err
	}
	uuidStr := id.String()
	tx.TransactionUuid = uuidStr

	q.Mu.Lock()
	defer q.Mu.Unlock()
	q.Q.PushBack(tx)
	return uuidStr, q.Q.Len(), nil
}

// Get a list of transactions to form a block. It returns a list of TXs to
func (q *EventQueue) GetTxsToFormBlock(maxTx int) ([]pb.Transaction, error) {
	q.Mu.Lock()
	defer q.Mu.Unlock()

	if maxTx == 0 || q.Q.Len() == 0 {
		return nil, errors.New("Block must contain more that 0 transactions")
	}

	count := 0
	var res []pb.Transaction
	for q.Q.Len() != 0 && count < maxTx {
		tx := q.Q.Front()
		res = append(res, tx.Value.(pb.Transaction))
		q.Q.Remove(tx)
	}
	return res, nil
}

type WireSystem struct {
	// This is the blockchain the Wire system manages.
	Blockchain blockchain.Blockchain
	Acnt Accounts
}

var _ Application = &WireSystem{}

func (ws *WireSystem) Init() {
	ws.Blockchain.Init()
}

func (ws *WireSystem) RBCReceive(bytes []byte) error {
	// TODO(chenweilunster): IMPLEMENT ME
	return nil
}

func (ws *WireSystem) GetTransactionStatus(txUuid string) (pb.TransactionStatus, error) {
	// TODO(chenweilunster): IMPLEMENT ME
	return pb.TransactionStatus_UNKNOWN, nil
}
