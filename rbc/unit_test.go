package rbc

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type MockLeader struct {
	*leader.Leader
}

type MockFollower struct {
	*follower.Follower
	savedPrepare []*pb.Payload
	savedEcho    []*pb.Payload
	savedReady   []*pb.ReadyRequest
}

func (mf *MockFollower) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	mf.savedEcho = append(mf.savedEcho, req)
	return &pb.EchoResponse{}, nil
}

func (mf *MockFollower) Ready(ctx context.Context, req *pb.ReadyRequest) (*pb.ReadyResponse, error) {
	mf.savedReady = append(mf.savedReady, req)
	return &pb.ReadyResponse{}, nil
}

func (mf *MockFollower) Prepare(ctx context.Context, req *pb.Payload) (*pb.PrepareResponse, error) {
	mf.savedPrepare = append(mf.savedEcho, req)
	return &pb.PrepareResponse{}, nil
}

//TODO: change it to 8000 and deal with connection failure
const port = 8009

func TestEcho(t *testing.T) {
	client := MockLeader{leader.NewLeader("L", nil, 1, nil, nil)}
	server := MockFollower{Follower: follower.NewFollower("F", nil, 1, nil, nil)}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	assert.Nil(t, err)
	//defer lis.Close()
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, &server)
	pb.RegisterReadyServer(s, &server)
	pb.RegisterPrepareServer(s, &server)
	var g errgroup.Group
	g.Go(func() error {
		return s.Serve(lis)
	})

	peer := &common.Peer{IP: "127.0.0.1", PORT: port}

	client.SendPrepare(peer, &pb.MerkleProof{Root: []byte("root")}, []byte("prepare"))
	// SendPrepare is async call, let's wait for 0.1s
	//time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 1, len(server.savedPrepare))
	assert.Equal(t, []byte("root"), server.savedPrepare[0].MerkleProof.Root)
	assert.Equal(t, []byte("prepare"), server.savedPrepare[0].Data)
	client.SendEcho(peer, &pb.MerkleProof{Root: []byte("root")}, []byte("echo"))
	// SendEcho is async call, let's wait for 0.1s
	//time.Sleep(time.Millisecond * 100)
	assert.Equal(t, 1, len(server.savedEcho))
	assert.Equal(t, []byte("root"), server.savedEcho[0].MerkleProof.Root)
	assert.Equal(t, []byte("echo"), server.savedEcho[0].Data)

	s.GracefulStop()
	assert.Nil(t, g.Wait())
	lis.Close()
}
