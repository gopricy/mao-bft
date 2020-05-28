package maobft

import (
	"github.com/gopricy/mao-bft/rbc/common"
	"strings"
	"testing"
	"time"

	"github.com/gopricy/mao-bft/application/transaction"
	"github.com/gopricy/mao-bft/rbc/mock"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

const faultLimit = 1
const followerNum = faultLimit * 3


var trans []string

func init(){
	logging.SetLevel(logging.INFO, "RBC")
}

func mockTransactions(leaderApp *transaction.Leader) map[string]int32 {
	propose := func(id string, err error) {
		if err != nil {
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

func mockInvalidTransactions(leaderApp *transaction.Leader) (map[string]int32, []error) {
	var errs []error
	propose := func(id string, err error) {
		if err != nil {
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

func createApps(num int) (res []common.Application){
	res = append(res, transaction.NewLeader(1, ""))
	for i := 0; i < num-1; i ++{
		res = append(res, transaction.NewFollower(""))
	}
	return
}

func TestIntegration_ValidSingleTxPerBlock(t *testing.T) {
	var g errgroup.Group

	rbcSetting, priKeys, cleaner := mock.InitPeers(faultLimit)
	var stoppers []func()
	apps := createApps(followerNum + 1)
	l, s := mock.StartLeader(t, apps[0], priKeys[0], rbcSetting, &g)
	apps[0].(*transaction.Leader).SetRBCLeader(l)
	stoppers = append(stoppers, s)
	mock.StartFollowers(t, apps[1:], priKeys[1:], rbcSetting, &g)

	exp := mockTransactions(apps[0].(*transaction.Leader))

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())

	ledgers := []*transaction.Ledger{apps[0].(transaction.Leader).Ledger}
	for _, f := range apps[1:] {
		ledgers = append(ledgers, f.(*transaction.Follower).Ledger)
	}

	for _, l := range ledgers {
		assert.Equal(t, exp, l.Accounts)
	}
	cleaner()
}

func TestIntegration_InvalidTransaction(t *testing.T) {
	var g errgroup.Group
	rbcSetting, priKeys, cleaner := mock.InitPeers(faultLimit)
	var stoppers []func()
	apps := createApps(followerNum + 1)
	l, s := mock.StartLeader(t, apps[0], priKeys[0], rbcSetting, &g)
	apps[0].(*transaction.Leader).SetRBCLeader(l)
	stoppers = append(stoppers, s)
	mock.StartFollowers(t, apps[1:], priKeys[1:], rbcSetting, &g)

	exp := mockTransactions(apps[0].(*transaction.Leader))

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	exp, errs := mockInvalidTransactions(apps[0].(*transaction.Leader))

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())

	ledgers := []*transaction.Ledger{apps[0].(*transaction.Leader).Ledger}
	for _, f := range apps[1:] {
		ledgers = append(ledgers, f.(*transaction.Follower).Ledger)
	}

	for _, l := range ledgers {
		assert.Equal(t, exp, l.Accounts)
	}
	assert.Equal(t, len(errs), 1)
	assert.True(t, strings.Contains(errs[0].Error(), "Invalid transaction:"))
	cleaner()
}
