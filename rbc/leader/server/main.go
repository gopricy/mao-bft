package main

import (
	"fmt"
	"net"
	"os"

	"github.com/gopricy/mao-bft/rbc/leader"

	"github.com/gopricy/mao-bft/pb"
	"google.golang.org/grpc"
)

func main() {
	leaderServer := leader.NewLeader("Mao", nil)
	port := os.Getenv("RBC_PORT")
	if port == "" {
		port = "8000"
	}
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	pb.RegisterEchoServer(s, &leaderServer)
	pb.RegisterReadyServer(s, &leaderServer)
	s.Serve(lis)
}
