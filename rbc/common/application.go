package common

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
	RBCReceive(block pb.Block) error
	// Add transaction to event queue for processing. Return a unique id for the transaction & length of event queue.
	// This function should be thread safe.
	AddTransactionToEventQueue(transaction pb.Transaction) (string, int, error)
	// Get a block to RBC, caller provides the maximum transaction a block can contain.
	GetBlockToRBC(maxTx int) (pb.Block, error)
	// Get status of a transaction by its uuid.
	GetTransactionStatus(txUuid string) (pb.TransactionStatus, error)
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

type WireSystem struct {
	Application
	// This is the blockchain the Wire system manages.
	Blockchain blockchain.Blockchain
	Queue EventQueue
	Acnt Accounts
}

func (ws *WireSystem) Init() {
	ws.Queue.Q = *list.New()
	ws.Blockchain.Init()
}

func (ws *WireSystem) RBCReceive(block pb.Block) error {
	ws.Blockchain.Mu.Lock()
	defer ws.Blockchain.Mu.Unlock()
	_, _, _, err := ws.Blockchain.CommitBlock(block)
	return err
}

func (ws *WireSystem) AddTransactionToEventQueue(transaction pb.Transaction) (string, int, error) {
	if transaction.TransactionUuid != "" {
		return "", 0, errors.New("uuid can not be set by client.")
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return "", 0, err
	}
	uuidStr := id.String()
	transaction.TransactionUuid = uuidStr

	ws.Queue.Mu.Lock()
	defer ws.Queue.Mu.Unlock()
	ws.Queue.Q.PushBack(transaction)
	return uuidStr, ws.Queue.Q.Len(), nil
}

func (ws *WireSystem) GetBlockToRBC(maxTx int) (pb.Block, error) {
	if maxTx == 0 || ws.Queue.Q.Len() == 0 {
		return pb.Block{}, errors.New("Block must contain more that 0 transactions")
	}
	ws.Queue.Mu.Lock()
	defer ws.Queue.Mu.Unlock()
	count := 0

	block := pb.Block{}
	for ws.Queue.Q.Len() != 0 && count < maxTx {
	}
	return block, nil
}