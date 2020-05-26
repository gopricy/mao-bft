package rbc

import (
	"fmt"
	"github.com/gopricy/mao-bft/pb"
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

const leaderPort = 8010
const address = "127.0.0.1"
const faultLimit = 1
const followerNum = 3
var g errgroup.Group

type mockApp struct{
	trans []string
}

func newMockApp() *mockApp{
	res := new(mockApp)
	res.trans = []string{}
	return res
}

var _ common.Application = &mockApp{}

func (m *mockApp) RBCReceive(bytes []byte) error{
	m.trans = append(m.trans, string(bytes))
	return nil
}

var allPeers [followerNum + 1]*common.Peer
var allApps [followerNum + 1]*mockApp

func init(){
	allPeers[0] = &common.Peer{Name: "mao", PORT: leaderPort, IP: address}
	allApps[0] = newMockApp()
	for i := 0; i < followerNum; i ++{
		allPeers[i + 1] = &common.Peer{Name: fmt.Sprintf("f%d", i+1), PORT: leaderPort + 1 + i, IP: address}
		allApps[i + 1] = newMockApp()
	}
}

func startFollower(t *testing.T, index int) (stopper func()){
	f := follower.NewFollower(fmt.Sprintf("f%d", index), allApps[index], faultLimit, allPeers[:])
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort + index))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterReadyServer(s, f)
	pb.RegisterEchoServer(s, f)
	pb.RegisterPrepareServer(s, f)
	g.Go(func() error{
		return s.Serve(lis)
	})
	return s.GracefulStop
}

func startLeader(t *testing.T) (mao Mao, stopper func()){
	l := leader.NewLeader("mao", allApps[0], faultLimit, allPeers[:])
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, l)
	pb.RegisterReadyServer(s, l)
	pb.RegisterPrepareServer(s, l)
	g.Go(func() error{
		return s.Serve(lis)
	})
	return l, s.GracefulStop
}

func TestIntegration(t *testing.T) {
	var stoppers []func()
	const testTrans = "Hello RBC!"
	l, s := startLeader(t)
	stoppers = append(stoppers, s)

	logging.SetLevel(logging.INFO, "RBC")
	for i := 1; i < 4; i ++{
		s := startFollower(t, i)
		stoppers = append(stoppers, s)
	}

	l.RBCSend([]byte(testTrans))

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())

	for _, a := range allApps{
		assert.Equal(t, 1, len(a.trans))
		assert.Equal(t, a.trans[0], testTrans)
	}

}

