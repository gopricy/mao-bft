package common

// Application manages blockchain & transaction, it also maintains internal App specific data structure for state machine.
type Application interface {
	// Once a message is RBC'ed, this function will be called to apply this block.
	// This function should be thread safe.
	// TODO?: can change it to block *pb.Block when we finalize it
	RBCReceive([]byte) error
}
