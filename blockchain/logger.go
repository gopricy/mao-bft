package blockchain

import (
	"github.com/golang/protobuf/proto"
	"github.com/gopricy/mao-bft/pb"
	mao_utils "github.com/gopricy/mao-bft/utils"
	"io/ioutil"
	"log"
	"os"
)

// Give logger complete permission.
const RWX = 0777

// Logger is a separate go routine that dumps block data to disk before every blockchain operation.
type Logger struct {
	// The directory to dump block information.
	dir string
	// The channel that write request is sending to.
	wRequests chan *logRequest
}

// A log request send to writer go routine.
type logRequest struct {
	blockDump pb.BlockDump
	done chan int
}

// This creates a new logger.
func NewLogger(dir string) *Logger {
	err := os.MkdirAll(dir, RWX)
	if err != nil {
		panic("Cannot make directory.")
	}

	logger := &Logger{dir: dir, wRequests: make(chan *logRequest)}
	go logger.handlerRoutine()

	return logger
}

func (logger *Logger) handlerRoutine() {
	for {
		req := <- logger.wRequests
		dump := req.blockDump
		fileName := mao_utils.GetFileNameFromBlockDump(dump)
		bytes, err := proto.Marshal(&dump)
		if err != nil {
			log.Fatalln("Failed to encode address book:", err)
		}
		err = ioutil.WriteFile(logger.dir + "/" + fileName, bytes, RWX)
		if err != nil {
			log.Fatalln("Failed to write byte to disk: " + logger.dir + "/" + fileName, err)
		}
		// Mark
		req.done <- 1
	}
}

// WriteBlock writes a block to disk. System will exist if encounters any failure.
func (logger *Logger) WriteBlock(block pb.Block, state pb.BlockState) {
	dump := pb.BlockDump{
		Block: &block,
		State: state,
	}
	done := make(chan int)
	req := logRequest{
		dump,
		done,
	}
	// Send request to handler routine, wait for done.
	logger.wRequests <- &req
	<- done
}

// ReadAllBlocks read all block dumps from local disk, return a list of block dump.
func (logger *Logger) ReadAllBlocks() ([]pb.BlockDump, error) {
	files, err := ioutil.ReadDir(logger.dir)
	if err != nil {
		return nil, err
	}

	var res []pb.BlockDump
	for _, file := range files {
		fname := file.Name()
		bytes, err := ioutil.ReadFile(logger.dir + "/" + fname)
		if err != nil {
			log.Fatalln("Error reading file:", err)
		}
		dumpBlock := pb.BlockDump{}
		if err := proto.Unmarshal(bytes, &dumpBlock); err != nil {
			log.Fatalln("Failed to parse BlockDump:", err)
		}
		res = append(res, dumpBlock)
	}

	return res, nil
}