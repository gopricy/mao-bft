package common

import (
	"context"
	"fmt"
	"sync"

	"github.com/gopricy/mao-bft/rbc/sign"

	"google.golang.org/grpc/metadata"

	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/erasure"
	"github.com/gopricy/mao-bft/rbc/merkle"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Received struct {
	// TODO: improve the efficiency with better locking
	rec map[merkle.RootString]map[string]interface{}
	mu  sync.Mutex
}



func (er *Received) Add(ip string, merkleRoot []byte, rec interface{}) (int, error) {
	er.mu.Lock()
	defer er.mu.Unlock()
	root := merkle.MerkleRootToString(merkleRoot)
	if er.rec == nil {
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
	AllPeers       map[string]*Peer
	ByzantineLimit int
}

type Peer struct {
	Name   string
	IP     string
	PORT   int
	CONN   *grpc.ClientConn
	PubKey sign.PublicKey
}

func (p *Peer) GoString() string {
	return fmt.Sprintf("%s", p.Name)
}

func (p *Peer) GetConn() *grpc.ClientConn {
	//for p.CONN == nil || p.CONN.GetState() == connectivity.Shutdown {
	conn, err := createConnection(p.IP, p.PORT)
	if err == nil {
		p.CONN = conn
		//break
	}
	//}
	return p.CONN
}

// Common is a building block of follower and leader
type Common struct {
	RBCSetting

	EchosReceived   Received
	ReadiesReceived Received
	PrevHashVoted map[string]merkle.RootString

	NodeName    string
	ReadiesSent sync.Map

	// Below are related to transaction system.
	App Application

	Logger *logging.Logger

	// privatekey
	privateKey *[64]byte
}

func NewCommon(name string, setting RBCSetting, app Application, privateKey *[64]byte) Common {
	//format := logging.MustStringFormatter(
	//	`%{time:15:05:05} %{module} %{message}`
	//)
	//log := logging.NewLogBackend(os.Stdout, "name", 0)
	return Common{RBCSetting: setting,
		NodeName:   name,
		App:        app,
		Logger:     logging.MustGetLogger("RBC"),
		privateKey: privateKey,
	}
}

func (c *Common) Verify(ctx context.Context, message []byte) ([]byte, bool, string) {
	name, err := c.getNameFromContext(ctx)
	if err != nil {
		return nil, false, ""
	}
	data, verified := sign.Verify(c.AllPeers[name].PubKey, message)
	return data, verified, name
}

func (c *Common) PrevHashValid(prevHash []byte, merkleRoot []byte) bool{
	if root, ok := c.PrevHashVoted[string(prevHash)]; ok{
		return merkle.MerkleRootToString(merkleRoot) == root
	}
	return true
}

func (c *Common) Sign(message []byte) []byte {
	return sign.Sign(c.privateKey, message)
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

func (c *Common) Debugf(format string, args ...interface{}) {
	c.Logger.Debugf("%s:"+format, append([]interface{}{c.Name()}, args...)...)
}

func (c *Common) Infof(format string, args ...interface{}) {
	c.Logger.Infof("%s:"+format, append([]interface{}{c.Name()}, args...)...)
}

func (c *Common) CreateContext() context.Context {
	md := metadata.Pairs("name", c.Name())
	return metadata.NewOutgoingContext(context.Background(), md)
}

func (c *Common) getNameFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("failed to decode context")
	}
	name, ok := md["name"]
	if !ok {
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
