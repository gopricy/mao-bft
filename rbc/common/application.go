package common

import (
	"github.com/gopricy/mao-bft/pb"
)

// Application manages blockchain & transaction, it also maintains internal App specific data structure for state machine.
type Application interface {
	// Once a message is RBC'ed, this function will be called to apply this block.
	// This function should be thread safe.
	RBCReceive(block pb.Block) error
	// Get status of a transaction by its uuid.
	GetTransactionStatus(txUuid string) (pb.TransactionStatus, error)
	// TODO(chenweilunster): Add validation functionality
}
