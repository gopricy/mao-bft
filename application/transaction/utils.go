package transaction

import (
	"container/list"
	"encoding/hex"
	"errors"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/gopricy/mao-bft/pb"
	"github.com/op/go-logging"
)

// #############################################################
// Define Wire System
// #############################################################

// Wire system manages this state machine. This state machine simply tracks account balance of each person.
type Ledger struct {
	Accounts map[string]int32
	mu       sync.RWMutex
}

func NewLedger() *Ledger {
	res := new(Ledger)
	res.Accounts = map[string]int32{}
	return res
}

// ValidateBlock validates that after committing this block, the ledger is still valid.
func (l *Ledger) ValidateBlock(block *pb.Block) bool {
	copyMap := make(map[string]int32)
	for k, v := range l.Accounts {
		copyMap[k] = v
	}

	if block.Content == nil {
		return false
	}

	for _, tx := range block.Content.Txs {
		switch v := tx.Message.(type) {
		case *pb.Transaction_WireMsg:
			copyMap[v.WireMsg.FromId] -= v.WireMsg.Amount
			if copyMap[v.WireMsg.FromId] < 0 {
				return false
			}
			copyMap[v.WireMsg.ToId] += v.WireMsg.Amount
		case *pb.Transaction_DepositMsg:
			if v.DepositMsg.Amount < 0 {
				return false
			}
			if _, ok := copyMap[v.DepositMsg.AccountId]; ok {
				copyMap[v.DepositMsg.AccountId] += v.DepositMsg.Amount
			} else {
				copyMap[v.DepositMsg.AccountId] = v.DepositMsg.Amount
			}
		default:
			return false
		}
	}
	return true
}

// Reconcile will take what already stored in blockchain and reconstruct application internal state.
func (l *Ledger) Reconcile(blocks []*pb.Block, isCommit []bool, isCommitLedger bool) {
	for i, block := range blocks {
		if !l.ValidateBlock(block) {
			logging.MustGetLogger("app").Debug("block is not valid or no content, hash is: " + hex.EncodeToString(block.CurHash))
			continue
		}
		if isCommitLedger && !isCommit[i] {
			continue
		}
		// Commit a block to ledger.
		for _, tx := range block.Content.Txs {
			err := l.CommitTxn(tx)
			if err != nil {
				log.Fatalln("err: " + err.Error())
			}
		}
	}
}

func (l *Ledger) GetBalance(act string) (int, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	balance, ok := l.Accounts[act]
	return int(balance), ok
}

// CommitTxn will commit an transaction and panic if account ID doesn't exist
func (l *Ledger) CommitTxn(txn *pb.Transaction) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	switch v := txn.Message.(type) {
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

func (l *Ledger) ValidateTransaction(txn *pb.Transaction) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	switch v := txn.Message.(type) {
	case *pb.Transaction_WireMsg:
		// FROM account should exist and not over withdraw.
		val, ok := l.Accounts[v.WireMsg.FromId]
		if !ok || v.WireMsg.Amount < 0 || val < v.WireMsg.Amount {
			return false
		}
		// TO account should exist.
		val, ok = l.Accounts[v.WireMsg.ToId]
		if !ok {
			return false
		}
		break
	case *pb.Transaction_DepositMsg:
		// Deposit should be greater than 0.
		if v.DepositMsg.Amount < 0 {
			return false
		}
		break
	default:
		return false
	}
	return true
}

// A event queue is a double sided queue that buffers the client proposed transaction.
type EventQueue struct {
	Q  list.List
	Mu sync.RWMutex
}

// Add a transaction to event queue if it's valid, it assigns a UUID to input transaction.
// This queue is managed by TransactionService.
func (q *EventQueue) AddTxToEventQueue(tx *pb.Transaction, pendingLedger *Ledger) (string, int, error) {
	q.Mu.Lock()
	defer q.Mu.Unlock()

	if tx.TransactionUuid != "" {
		return "", -1, errors.New("uuid can not be set by client")
	}
	if !pendingLedger.ValidateTransaction(tx) {
		return "", -1, errors.New("Invalid Transaction: " + proto.MarshalTextString(tx))
	}

	id, err := uuid.NewUUID()
	if err != nil {
		return "", 0, err
	}
	uuidStr := id.String()
	tx.TransactionUuid = uuidStr

	q.Q.PushBack(tx)
	if err := pendingLedger.CommitTxn(tx); err != nil {
		return "", -1, errors.New("Cannot commit transaction in pending ledger.")
	}
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

func (q *EventQueue) Exist(uuid string) bool {
	q.Mu.RLock()
	defer q.Mu.RUnlock()

	p := q.Q.Front()
	for p != nil {
		if p.Value.(*pb.Transaction).TransactionUuid == uuid {
			return true
		}
	}
	return false
}
