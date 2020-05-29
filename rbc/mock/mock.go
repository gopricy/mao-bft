package mock

import (
	"fmt"
	"net"
	"testing"

	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/follower"
	"github.com/gopricy/mao-bft/rbc/leader"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/sign"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const leaderPort = 8010
const address = "127.0.0.1"

func InitPeers(byzantineLimit int) (
	rbcSetting common.RBCSetting,
	allPrivateKeys []*[64]byte,
	connCloser func()) {
	rbcSetting.ByzantineLimit = byzantineLimit
	followerNum := byzantineLimit * 3
	pub, priv, _ := sign.GenerateKey(nil)
	rbcSetting.AllPeers = make(map[string]*common.Peer)
	rbcSetting.AllPeers["map"] = &common.Peer{Name: "mao", PORT: leaderPort, IP: address, PubKey: *pub}
	allPrivateKeys = append(allPrivateKeys, priv)
	for i := 0; i < followerNum; i++ {
		name := fmt.Sprintf("f%d", i+1)
		pub, priv, _ := sign.GenerateKey(nil)
		rbcSetting.AllPeers[name] = &common.Peer{Name: fmt.Sprintf("f%d", i+1), PORT: leaderPort + 1 + i, IP: address, PubKey: *pub}
		allPrivateKeys = append(allPrivateKeys, priv)
	}
	connCloser = func() {
		for _, p := range rbcSetting.AllPeers {
			p.CONN.Close()
		}
	}
	return
}

func StartFollowers(t *testing.T, apps []common.Application, privKeys []*[64]byte, rs common.RBCSetting, g *errgroup.Group) (stoppers []func()) {
	if len(apps) != len(privKeys) {
		panic("apps and privKeys should have same length")
	}
	followerNum := len(apps)
	for i := 0; i < followerNum; i++ {
		f := follower.NewFollower(fmt.Sprintf("f%d", i+1), apps[i], rs.ByzantineLimit, rs.AllPeers, privKeys[i])
		lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort+i+1))
		assert.Nil(t, err)
		s := grpc.NewServer()

		pb.RegisterReadyServer(s, f)
		pb.RegisterEchoServer(s, f)
		pb.RegisterPrepareServer(s, f)
		g.Go(func() error {
			return s.Serve(lis)
		})
		stoppers = append(stoppers, s.Stop)
	}
	return stoppers
}

func StartLeader(t *testing.T, app common.Application, privKey *[64]byte, rs common.RBCSetting, g *errgroup.Group) (mao *leader.Leader, stopper func()) {
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
	return l, s.Stop
}
