package mock

import (
	"fmt"
	"net"
	"testing"

	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	"github.com/gopricy/mao-bft/rbc/sign"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const leaderPort = 8010
const address = "127.0.0.1"

func InitPeers(byzantineLimit int) (rbcSetting common.RBCSetting, allPrivateKeys []*[64]byte, connCloser func() error) {
	rbcSetting.ByzantineLimit = byzantineLimit
	followerNum := byzantineLimit * 3
	pub, priv := sign.GenerateKey()
	rbcSetting.AllPeers = make(map[string]*common.Peer)
	rbcSetting.AllPeers["mao"] = &common.Peer{Name: "mao", PORT: leaderPort, IP: address, PubKey: pub}
	allPrivateKeys = append(allPrivateKeys, priv)
	for i := 0; i < followerNum; i++ {
		name := fmt.Sprintf("f%d", i+1)
		pub, priv := sign.GenerateKey()
		rbcSetting.AllPeers[name] = &common.Peer{Name: fmt.Sprintf("f%d", i+1), PORT: leaderPort + 1 + i, IP: address, PubKey: pub}
		allPrivateKeys = append(allPrivateKeys, priv)
	}
	connCloser = func() error{
		for _, p := range rbcSetting.AllPeers {
			if err := p.CONN.Close(); err != nil{
				return err
			}
		}
		return nil
	}
	return
}

func StartFollowers(t *testing.T, apps []common.Application, privKeys []*[64]byte, rs common.RBCSetting, g *errgroup.Group) (stoppers []func()) {
	if len(apps) != len(privKeys) {
		panic("apps and privKeys should have same length")
	}
	followerNum := len(apps)
	for i := 0; i < followerNum; i++ {
		err, stopper := NewFollower(apps[i], i, privKeys[i], rs, g)
		assert.Nil(t, err)
		stoppers = append(stoppers, stopper)
	}
	return stoppers
}

func NewFollower(app common.Application, index int, privKey sign.PrivateKey, rs common.RBCSetting, g *errgroup.Group) (error, func()){
	f := follower.NewFollower(fmt.Sprintf("f%d", index+1), app, rs.ByzantineLimit, rs.AllPeers, privKey)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort+index+1))
	if err != nil{
		return err, func(){}
	}
	s := grpc.NewServer()

	pb.RegisterReadyServer(s, f)
	pb.RegisterEchoServer(s, f)
	pb.RegisterPrepareServer(s, f)
	g.Go(func() error {
		return s.Serve(lis)
	})
	return nil, s.GracefulStop
}

func StartLeader(t *testing.T, app common.Application, privKey sign.PrivateKey, rs common.RBCSetting, g *errgroup.Group) (mao *leader.Leader, stopper func()) {
	l := leader.NewLeader("mao", app, rs.ByzantineLimit, rs.AllPeers, privKey)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, l)
	pb.RegisterReadyServer(s, l)
	pb.RegisterPrepareServer(s, l)
	g.Go(func() error {
		return s.Serve(lis)
	})
	return l, s.GracefulStop
}

func NewLeader(app common.Application, privKey sign.PrivateKey, rs common.RBCSetting, g *errgroup.Group) (mao *leader.Leader, stopper func()) {
	l := leader.NewLeader("mao", app, rs.ByzantineLimit, rs.AllPeers, privKey)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	if err != nil{
		return nil, func(){}
	}
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, l)
	pb.RegisterReadyServer(s, l)
	pb.RegisterPrepareServer(s, l)
	g.Go(func() error {
		return s.Serve(lis)
	})
	return l, s.GracefulStop
}
