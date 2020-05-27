package maobft

import (
	"fmt"
	"github.com/gopricy/mao-bft/application/transaction"
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
	"strings"
	"testing"
	"time"
)

const leaderPort = 8000
const address = "127.0.0.1"
const faultLimit = 1
const followerNum = 3
var g errgroup.Group

var allPeers [followerNum + 1]*common.Peer
var leaderApp *transaction.Leader
var followerApps [followerNum]*transaction.Follower

var trans []string

func initPeers(){
	allPeers[0] = &common.Peer{Name: "mao", PORT: leaderPort, IP: address}
	leaderApp = transaction.NewLeader(1)
	for i := 0; i < followerNum; i ++{
		allPeers[i + 1] = &common.Peer{Name: fmt.Sprintf("f%d", i+1), PORT: leaderPort + 1 + i, IP: address}
		followerApps[i] = transaction.NewFollower()
	}
}

func mockTransactions() map[string]int32{
	propose := func(id string, err error){
		if err != nil{
			panic(err)
		}
		trans = append(trans, id)
	}
	propose(leaderApp.ProposeDeposit("001", 50, 50))
	propose(leaderApp.ProposeDeposit("002", 100, 0))
	propose(leaderApp.ProposeTransfer("001", "002", 30, 0))
	propose(leaderApp.ProposeDeposit("002", 0, 50))
	expected := map[string]int32{}
	expected["001"] = 2050
	expected["002"] = 13050
	return expected
}

func mockInvalidTransactions() (map[string]int32, []error){
	var errs []error
	propose := func(id string, err error){
		if err != nil{
			errs = append(errs, err)
		}
		trans = append(trans, id)
	}
	propose(leaderApp.ProposeDeposit("001", 50, 50))
	propose(leaderApp.ProposeDeposit("002", 100, 0))
	propose(leaderApp.ProposeTransfer("001", "002", 30, 0))
	propose(leaderApp.ProposeTransfer("001", "002", 30, 0))
	expected := map[string]int32{}
	expected["001"] = 2050
	expected["002"] = 13000
	return expected, errs
}

func startFollower(t *testing.T, index int) (stopper func(), teardown func()){
	f := follower.NewFollower(fmt.Sprintf("f%d", index), followerApps[index-1], faultLimit, allPeers[:])
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort + index))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterReadyServer(s, f)
	pb.RegisterEchoServer(s, f)
	pb.RegisterPrepareServer(s, f)
	g.Go(func() error{
		return s.Serve(lis)
	})
	return s.GracefulStop, s.Stop
}

func startLeader(t *testing.T) (mao rbc.Mao, stopper func(), teardown func()){
	l := leader.NewLeader("mao", leaderApp, faultLimit, allPeers[:])
	leaderApp.SetRBCLeader(l)
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", address, leaderPort))
	assert.Nil(t, err)
	s := grpc.NewServer()

	pb.RegisterEchoServer(s, l)
	pb.RegisterReadyServer(s, l)
	pb.RegisterPrepareServer(s, l)
	g.Go(func() error{
		return s.Serve(lis)
	})
	return l, s.GracefulStop, s.Stop
}

func TestIntegration_ValidSingleTxPerBlock(t *testing.T) {
	initPeers()
	var stoppers []func()
	var teardown []func()
	_, s, tear := startLeader(t)
	stoppers = append(stoppers, s)
	teardown = append(teardown, tear)

	logging.SetLevel(logging.INFO, "RBC")
	for i := 1; i < 4; i ++{
		s, tear := startFollower(t, i)
		stoppers = append(stoppers, s)
		teardown = append(teardown, tear)
	}

	exp := mockTransactions()

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())

	ledgers := []*transaction.Ledger{leaderApp.Ledger}
	for _, f := range followerApps{
		ledgers = append(ledgers, f.Ledger)
	}

	for _, l := range ledgers{
		assert.Equal(t, exp, l.Accounts)
	}
}

func TestIntegration_InvalidTransaction(t *testing.T) {
	initPeers()
	var stoppers []func()
	_, s, _ := startLeader(t)
	stoppers = append(stoppers, s)

	logging.SetLevel(logging.INFO, "RBC")
	for i := 1; i < 4; i ++{
		s, _ := startFollower(t, i)
		stoppers = append(stoppers, s)
	}

	exp, errs := mockInvalidTransactions()

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())

	ledgers := []*transaction.Ledger{leaderApp.Ledger}
	for _, f := range followerApps{
		ledgers = append(ledgers, f.Ledger)
	}

	for _, l := range ledgers{
		assert.Equal(t, exp, l.Accounts)
	}
	assert.Equal(t, len(errs), 1)
	assert.True(t, strings.Contains(errs[0].Error(), "Invalid transaction:"))
}