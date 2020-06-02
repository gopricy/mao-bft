package common

import "github.com/gopricy/mao-bft/pb"

// Application manages blockchain & transaction, it also maintains internal App specific data structure for state machine.
type Application interface {
	// Once a message is RBC'ed, this function will be called to apply this block.
	// This function should be thread safe.
	// This function returns staged area's length
	// TODO?: can change it to block *pb.Block when we finalize it
	RBCReceive([]byte) (bool, error)

	// GetSyncQuestion will return sync request constructed from App blockchain.
	GetSyncQuestion() (pb.SyncRequest, error)
	// GetSyncAnswer will return sync answer for the corresponding sync question.
	GetSyncAnswer(request pb.SyncRequest) (pb.SyncResponse, error)
}
