package maobft

import (
	"fmt"
	"github.com/gopricy/mao-bft/application/wire"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"net"
	"sync"
	"testing"
	"time"
)

const leaderPort = 8000
const address = "127.0.0.1"
const faultLimit = 1
const followerNum = 3
var wg sync.WaitGroup

var allPeers []*common.Peer

func init(){
	for i := 0; i < followerNum; i ++{
		allPeers = append(allPeers, &common.Peer{PORT: leaderPort + 1 + i, IP: address})
	}
	allPeers = append(allPeers, &common.Peer{PORT: leaderPort, IP: address})
}

func startFollower(t *testing.T, index int) (stopper func()){
	f := follower.NewFollower(fmt.Sprintf("f%d", 8000 + index), new(wire.WireSystem), faultLimit, allPeers)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort + index))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterReadyServer(s, &f)
	pb.RegisterEchoServer(s, &f)
	pb.RegisterPrepareServer(s, &f)
	wg.Add(1)
	go func(){
		assert.Nil(t, s.Serve(lis))
		wg.Done()
	}()
	return s.Stop
}

func startLeader(t *testing.T) (mao rbc.Mao, stopper func()){
	l := leader.NewLeader("mao", new(wire.WireSystem), faultLimit, allPeers)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, &l)
	pb.RegisterReadyServer(s, &l)
	pb.RegisterPrepareServer(s, &l)
	wg.Add(1)
	go func(){
		assert.Nil(t, s.Serve(lis))
		wg.Done()
	}()
	return &l, s.Stop
}

func TestIntegration(t *testing.T) {
	var stoppers []func()
	l, s := startLeader(t)
	stoppers = append(stoppers, s)

	for i := 1; i < 4; i ++{
		s := startFollower(t, i)
		stoppers = append(stoppers, s)
	}

	defer func() {
		for _, s := range stoppers {
			s()
		}
	}()

	const testString = "Hello RBC!"
	l.RBCSend([]byte(testString))

	time.Sleep(time.Second * 5)
}