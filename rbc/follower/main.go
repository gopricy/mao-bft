package follower

import (
	"fmt"
	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
	"net"
	"os"
)

func main(){
	followerServer := Follower{}
	port := os.Getenv("RBC_PORT")
	if port == ""{
		port = "8000"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil{
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterEchoServer(s, &followerServer)
	pb.RegisterReadyServer(s, &followerServer)
	pb.RegisterPrepareServer(s, &followerServer)
	s.Serve(lis)
}
