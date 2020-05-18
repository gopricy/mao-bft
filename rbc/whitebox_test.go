package rbc

import (
	"context"
	"fmt"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"testing"
)

type MockLeader struct{
	leader.Leader
}

type MockFollower struct{
	follower.Follower
	savedPrepare []*pb.Payload
	savedEcho []*pb.Payload
	savedReady []*pb.ReadyRequest
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

const port = 8000

func TestEcho(t *testing.T){
	client := MockLeader{leader.NewLeader("L", nil)}
	server := MockFollower{Follower: follower.NewFollower("F", nil)}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, &server)
	pb.RegisterReadyServer(s, &server)
	pb.RegisterPrepareServer(s, &server)
	go s.Serve(lis)

	peer := common.Peer{IP:"127.0.0.1", PORT: 8000}

	err = client.SendPrepare(peer, &pb.MerkleProof{Root:[]byte("root")}, []byte("prepare"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(server.savedPrepare))
	assert.Equal(t, []byte("root"), server.savedPrepare[0].MerkleProof.Root)
	assert.Equal(t, []byte("prepare"), server.savedPrepare[0].Data)
	err = client.SendEcho(peer, &pb.MerkleProof{Root:[]byte("root")}, []byte("echo"))
	assert.Nil(t, err)
	assert.Equal(t, 1, len(server.savedEcho))
	assert.Equal(t, []byte("root"), server.savedEcho[0].MerkleProof.Root)
	assert.Equal(t, []byte("echo"), server.savedEcho[0].Data)
	s.GracefulStop()
}
