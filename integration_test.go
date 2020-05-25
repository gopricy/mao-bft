package maobft

import (
	"fmt"
	"github.com/gopricy/mao-bft/application/wire"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

const leaderPort = 8000
const address = "127.0.0.1"
const faultLimit = 1
const followerNum = 3
var g errgroup.Group

var allPeers []*common.Peer

func init(){
	allPeers = append(allPeers, &common.Peer{Name: "mao", PORT: leaderPort, IP: address})
	for i := 0; i < followerNum; i ++{
		allPeers = append(allPeers, &common.Peer{Name: fmt.Sprintf("f%d", i+1), PORT: leaderPort + 1 + i, IP: address})
	}
}

func startFollower(t *testing.T, index int) (stopper func()){
	f := follower.NewFollower(fmt.Sprintf("f%d", index), new(wire.WireSystem), faultLimit, allPeers)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort + index))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterReadyServer(s, &f)
	pb.RegisterEchoServer(s, &f)
	pb.RegisterPrepareServer(s, &f)
	g.Go(func() error{
		return s.Serve(lis)
	})
	return s.GracefulStop
}

func startLeader(t *testing.T) (mao rbc.Mao, stopper func()){
	l := leader.NewLeader("mao", new(wire.WireSystem), faultLimit, allPeers)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, &l)
	pb.RegisterReadyServer(s, &l)
	pb.RegisterPrepareServer(s, &l)
	g.Go(func() error{
		return s.Serve(lis)
	})
	return &l, s.GracefulStop
}

func TestIntegration(t *testing.T) {
	var stoppers []func()
	l, s := startLeader(t)
	stoppers = append(stoppers, s)

	logging.SetLevel(logging.INFO, "RBC")
	for i := 1; i < 4; i ++{
		s := startFollower(t, i)
		stoppers = append(stoppers, s)
	}


	const testString = "Hello RBC!"
	l.RBCSend([]byte(testString))

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())
}