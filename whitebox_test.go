package maobft

import (
	"context"
	"github.com/gopricy/mao-bft/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"testing"
)

type MockServer struct{
	Leader
	savedPayload []*pb.Payload
}


func (ms *MockServer) Echo(ctx context.Context, req *pb.Payload) (*pb.EchoResponse, error) {
	ms.savedPayload = append(ms.savedPayload, req)
	return &pb.EchoResponse{}, nil
}

type MockClient struct{
	Follower
}

const address = "localhost:8000"

func TestEcho(t *testing.T){
	client := MockClient{NewFollower("F")}
	server := MockServer{NewLeader("L"), []*pb.Payload{}}
	lis, err := net.Listen("tcp", address)
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, &server)
	go s.Serve(lis)

	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	assert.Nil(t, err)
	defer conn.Close()
	client.SendEcho(conn, &pb.MerkleProof{Root:[]byte{'1', '2'}}, []byte{'a', 'b', 'c'})
	s.GracefulStop()
	assert.Equal(t, 1, len(server.savedPayload))
	assert.Equal(t, 2, len(server.savedPayload[0].MerkleProof.Root))
}
