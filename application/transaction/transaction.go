package transaction

import (
	"github.com/gopricy/mao-bft/blockchain"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/follower"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"github.com/pkg/errors"
)

const MaximumTxn = 1000000

// Application manages blockchain & transaction, it also maintains internal App specific data structure for state machine.
type Application interface {
	// Once a message is RBC'ed, this function will be called to apply this block.
	// This function should be thread safe.
	RBCReceive(bytes []byte) error
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
	Ledger *Ledger
}

func newcommon() *common{
	res := new(common)
	res.Queue = new(EventQueue)
	res.Ledger = NewLedger()
	res.Blockchain = blockchain.NewBlockchain()
	return res
}

var _ Application = &common{}

func (c *common) RBCReceive(bytes []byte) error {
	block, err := mao_utils.DecodeBlock(bytes)
	if err != nil{
		return errors.Wrap(err, "Can't decode Block")
	}
	blocks, err := c.Blockchain.CommitBlock(block)
	for _, b := range blocks{
		for _, t := range b.Content.Txs{
			if err := c.Ledger.CommitTxn(t); err != nil{
				return err
			}
		}
	}
	return nil
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
}

func (l *Leader) SetRBCLeader(leader RBCLeader){
	l.Leader = leader
}

func NewLeader(blocksize int) *Leader{
	res := new(Leader)
	res.common = newcommon()
	res.MaxBlockSize = blocksize
	return res
}

func (l *Leader) ProposeTransfer(from, to string, dollar, cents int) (string, error){
	if dollar > MaximumTxn || cents >= 100 || cents < 0{
		return "", errors.New("invalid amount, transaction limit is 1M")
	}
	txn := &pb.Transaction{
		Message: &pb.Transaction_WireMsg{
			&pb.WireMessage{
				FromId: from,
				ToId: to,
				Amount: int32(dollar * 100 + cents),
			},
		},
	}
	u, t, err := l.Queue.AddTxToEventQueue(txn)
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
	u, t, err := l.Queue.AddTxToEventQueue(txn)
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


func (l *Leader) createBlockAndSend() error{
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

func NewFollower() *Follower{
	res := new(Follower)
	res.common = newcommon()
	return res
}
