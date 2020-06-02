package common

import (
"context"

"github.com/gopricy/mao-bft/pb"
)

// Sync sends out request and get response by
func (c *Common) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncResponse, error) {
	return nil, nil
}

// TODO(Sync): send sync request.
