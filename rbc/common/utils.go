package common

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
)

// TODO: add TLS
func createConnection(ip string, port int) (*grpc.ClientConn, error) {
	//TODO: PERFORMANCE WithBlock is a blocking call, probably need unblocking call for performance
	return grpc.Dial(fmt.Sprintf("%s:%d", ip, port), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(30*time.Second))
}
