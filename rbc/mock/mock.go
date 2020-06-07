package mock

import (
	"fmt"
	"net"
	"testing"

	"github.com/fatih/color"
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
	connCloser = func() error {
		for _, p := range rbcSetting.AllPeers {
			if err := p.CONN.Close(); err != nil {
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
		err, stopper := NewFollower(apps[i], i+1, privKeys[i], rs, g)
		assert.Nil(t, err)
		stoppers = append(stoppers, stopper)
	}
	return stoppers
}

// if g is provided, it is a nonblocking call. if g is nil, it is a blocking call
func NewFollower(app common.Application, index int, privKey sign.PrivateKey, rs common.RBCSetting, g *errgroup.Group) (error, func()) {
	name := fmt.Sprintf("f%d", index)
	p := rs.AllPeers[name].PORT
	f := follower.NewFollower(name, app, rs.ByzantineLimit, rs.AllPeers, privKey)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, p))

	if err != nil {
		return err, func() {}
	}
	s := grpc.NewServer()

	pb.RegisterReadyServer(s, f)
	pb.RegisterEchoServer(s, f)
	pb.RegisterPrepareServer(s, f)
	pb.RegisterSyncServer(s, f)
	if g == nil {
		f.Debugf(color.CyanString("Follower %d starts to listen on %s:%d", index, address, p))
		err = s.Serve(lis)
		return err, func() {}
	}
	f.Debugf("RBC Follower starts to listen on %s:%d", address, p)
	g.Go(func() error {
		return s.Serve(lis)
	})
	return nil, s.GracefulStop
}

func StartLeader(t *testing.T, app common.Application, privKey sign.PrivateKey, rs common.RBCSetting,
	g *errgroup.Group) (mao *leader.Leader, stopper func()) {
	mao, stopper, err := NewLeader(app, privKey, rs, g)
	assert.Nil(t, err)
	return
}

// if g is provided, it is a nonblocking call. if g is nil, it is a blocking call
func NewLeader(app common.Application, privKey sign.PrivateKey, rs common.RBCSetting, g *errgroup.Group) (
	mao *leader.Leader, stopper func(), err error) {
	l := leader.NewLeader("mao", app, rs.ByzantineLimit, rs.AllPeers, privKey)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	if err != nil {
		return nil, func() {}, err
	}
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, l)
	pb.RegisterReadyServer(s, l)
	pb.RegisterPrepareServer(s, l)
	pb.RegisterSyncServer(s, l)
	l.Debugf("RBC Leader starts to listen on %s:%d", address, leaderPort)
	if g == nil {
		err = s.Serve(lis)
		return l, func() {}, nil
	}
	g.Go(func() error {
		return s.Serve(lis)
	})
	return l, s.GracefulStop, nil

}
