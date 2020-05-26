package transaction

import (
	"container/list"
	"errors"
	"github.com/google/uuid"
	"github.com/gopricy/mao-bft/pb"
	"sync"
)


// #############################################################
// Define Wire System
// #############################################################

// Wire system manages this state machine. This state machine simply tracks account balance of each person.
type Ledger struct{
	Accounts map[string]int32
	mu sync.RWMutex
}

func NewLedger() *Ledger{
	res := new(Ledger)
	res.Accounts = map[string]int32{}
	return res
}

func (l *Ledger) GetBalance(act string) (int, bool){
	l.mu.RLock()
	defer l.mu.RUnlock()

	balance, ok := l.Accounts[act]
	return int(balance), ok
}

// CommitTxn will commit an transaction and panic if account ID doesn't exist
func (l *Ledger) CommitTxn(txn *pb.Transaction) error{
	l.mu.Lock()
	defer l.mu.Unlock()

	switch v := txn.Message.(type){
	case *pb.Transaction_WireMsg:
		l.Accounts[v.WireMsg.FromId] -= v.WireMsg.Amount
		l.Accounts[v.WireMsg.ToId] += v.WireMsg.Amount
	case *pb.Transaction_DepositMsg:
		if _, ok := l.Accounts[v.DepositMsg.AccountId]; ok {
			l.Accounts[v.DepositMsg.AccountId] += v.DepositMsg.Amount
		} else {
			l.Accounts[v.DepositMsg.AccountId] = v.DepositMsg.Amount
		}

	default:
		return errors.New("unsupported txn type")
	}
	return nil
}

// A event queue is a double sided queue that buffers the client proposed transaction.
type EventQueue struct {
	Q list.List
	Mu sync.RWMutex
}

// Add a transaction to event queue, it assigns a UUID to input transaction.
// This queue is managed by TransactionService.
func (q *EventQueue) AddTxToEventQueue(tx *pb.Transaction) (string, int, error) {
	if tx.TransactionUuid != "" {
		return "", 0, errors.New("uuid can not be set by client")
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

// Get a list of transactions to form a block. It returns a list of TXs
func (q *EventQueue) GetTransactions(maxTx int) ([]*pb.Transaction, error) {
	q.Mu.Lock()
	defer q.Mu.Unlock()

	if maxTx == 0 {
		return nil, errors.New("Block must contain more that 0 transactions")
	}

	count := 0
	var res []*pb.Transaction
	for q.Q.Len() != 0 && count < maxTx {
		tx := q.Q.Front()
		res = append(res, tx.Value.(*pb.Transaction))
		q.Q.Remove(tx)
	}
	return res, nil
}

func (q *EventQueue) Exist(uuid string) bool{
	q.Mu.RLock()
	defer q.Mu.RUnlock()

	p := q.Q.Front()
	for p != nil{
		if p.Value.(*pb.Transaction).TransactionUuid == uuid{
			return true
		}
	}
	return false
}
