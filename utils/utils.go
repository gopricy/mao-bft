package mao_utils

import (
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/pb"
)

func IsSameBytes(left []byte, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for i, val := range left {
		if right[i] != val {
			return false
		}
	}
	return true
}

func FromBytesToBlock(bytes []byte) (pb.Block, error) {
	var Block pb.Block
	err := proto.Unmarshal(bytes, &Block)
	return Block, err
}
