package common

import (
	"context"
	"fmt"
	"google.golang.org/grpc/metadata"
	"sync"

	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/erasure"
	"github.com/gopricy/mao-bft/rbc/merkle"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type Received struct {
	// TODO: improve the efficiency with better locking
	rec map[merkle.RootString]map[string]interface{}
	mu  sync.Mutex
}

type contextKey string
const keyName =  contextKey("name")

func (er *Received) Add(ip string, merkleRoot []byte, rec interface{}) (int, error) {
	er.mu.Lock()
	defer er.mu.Unlock()
	root := merkle.MerkleRootToString(merkleRoot)
	if er.rec == nil{
		er.rec = make(map[merkle.RootString]map[string]interface{})
	}
	if _, ok := er.rec[root]; !ok {
		// if this message hasn't been seen
		er.rec[root] = make(map[string]interface{})
	}
	if _, ok := er.rec[root][ip]; ok {
		return len(er.rec[root]), errors.New("Duplicate ECHO from same IP carrying same message")
	}
	er.rec[root][ip] = rec
	return len(er.rec[root]), nil
}

type RBCSetting struct {
	AllPeers       []*Peer
	ByzantineLimit int
}

type Peer struct {
	Name string
	IP   string
	PORT int
	CONN *grpc.ClientConn
}

func (p *Peer) GoString() string{
	return fmt.Sprintf("%s", p.Name)
}

func (p *Peer) GetConn() *grpc.ClientConn {
	for p.CONN == nil || p.CONN.GetState() == connectivity.Shutdown {
		conn, err := createConnection(p.IP, p.PORT)
		if err == nil {
			p.CONN = conn
			break
		}
	}
	return p.CONN
}

// Common is a building block of follower and leader
type Common struct {
	RBCSetting

	EchosReceived   Received
	ReadiesReceived Received

	NodeName string
	ReadiesSent sync.Map

	// Below are related to transaction system.
	App Application

	Logger *logging.Logger

	// TODO: remove it from rbc
	//pb.UnimplementedTransactionServiceServer
	//Queue Event
}

func NewCommon(name string, setting RBCSetting, app Application) Common{
	//format := logging.MustStringFormatter(
	//	`%{time:15:05:05} %{module} %{message}`
	//)
	//log := logging.NewLogBackend(os.Stdout, "name", 0)
	return Common{RBCSetting: setting,
		NodeName: name,
		App: app,
		Logger: logging.MustGetLogger("RBC"),
	}
}

func (c *Common) reconstructData(root merkle.RootString) ([]byte, error) {
	payloads := []*pb.Payload{}
	for _, m := range c.EchosReceived.rec[root] {
		payloads = append(payloads, m.(*pb.Payload))
	}
	return erasure.Reconstruct(payloads, c.ByzantineLimit, len(c.AllPeers))
}

func (c *Common) readyIsSent(merkleroot []byte) bool {
	if _, ok := c.ReadiesSent.Load(merkle.MerkleRootToString(merkleroot)); !ok {
		c.ReadiesSent.Store(merkle.MerkleRootToString(merkleroot), struct{}{})
		return false
	}
	return true
}

func (c *Common) Name() string {
	return c.NodeName
}

func (c *Common) Debugf(format string, args ...interface{}){
	c.Logger.Debugf("%s:" + format, append([]interface{}{c.Name()}, args...)...)
}

func (c *Common) Infof(format string, args ...interface{}){
	c.Logger.Infof("%s:" + format, append([]interface{}{c.Name()}, args...)...)
}




func (c *Common) CreateContext() context.Context{
	md := metadata.Pairs("name", c.Name())
	return metadata.NewOutgoingContext(context.Background(), md)
}

func (c *Common) GetNameFromContext(ctx context.Context) (string, error){
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok{
		return "", errors.New("failed to decode context")
	}
	name, ok := md["name"]
	if !ok{
		return "", errors.New("context doesn't have name")
	}

	return name[0], nil
}


func (c *Common) ProposeTransaction(
	ctx context.Context, in *pb.ProposeTransactionRequest) (*pb.ProposeTransactionResponse, error) {
	// TODO(chenweilunster): IMPLEMENT ME
	return nil, nil
}

func (c *Common) GetTransactionStatus(
	ctx context.Context, in *pb.GetTransactionStatusRequest) (*pb.GetTransactionStatusResponse, error) {
	// TODO(chenweilunster): IMPLEMENT ME
	return nil, nil
}
