package rbc

import (
	"testing"
	"time"

	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/mock"
	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

var g errgroup.Group

const byzantineLimit = 1
const followerNum = byzantineLimit * 3

type mockApp struct {
	trans []string
}

func newMockApp() *mockApp {
	res := new(mockApp)
	res.trans = []string{}
	return res
}

var _ common.Application = &mockApp{}

func (m *mockApp) RBCReceive(bytes []byte) error {
	m.trans = append(m.trans, string(bytes))
	return nil
}

func createApps() []*mockApp {
	res := make([]*mockApp, followerNum+1)
	for i := 0; i < followerNum+1; i++ {
		res[i] = newMockApp()
	}
	return res
}

func TestIntegration(t *testing.T) {
	var g errgroup.Group
	logging.SetLevel(logging.INFO, "RBC")
	rs, keys, closer := mock.InitPeers(byzantineLimit)
	defer closer()
	var stoppers []func()
	const testTrans = "Hello RBC!"
	apps := createApps()
	l, s := mock.StartLeader(t, apps[0], keys[0], rs, &g)
	stoppers = append(stoppers, s)
	var Apps []common.Application
	for i := 0; i < len(apps) - 1; i++{
		Apps = append(Apps, apps[i + 1])
	}
	stoppers = append(stoppers, mock.StartFollowers(t, Apps, keys[1:], rs, &g)...)

	l.RBCSend([]byte(testTrans))

	time.Sleep(time.Second * 1)
	for _, s := range stoppers {
		s()
	}

	assert.Nil(t, g.Wait())

	for _, a := range apps {
		assert.Equal(t, 1, len(a.trans))
		assert.Equal(t, a.trans[0], testTrans)
	}

}
