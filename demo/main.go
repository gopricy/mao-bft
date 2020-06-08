// [_Command-line flags_](http://en.wikipedia.org/wiki/Command-line_interface#Command-line_option)
// are a common way to specify options for command-line
// programs. For example, in `wc -l` the `-l` is a
// command-line flag.

package main

// Go provides a `flag` package supporting basic
// command-line flag parsing. We'll use this package to
// implement our example command-line program.
import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/gopricy/mao-bft/application/transaction"
	"github.com/gopricy/mao-bft/pb"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/mock"
	"github.com/gopricy/mao-bft/rbc/sign"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const rbcSetting = "rbc_setting.json"
const privateKeys = "private_keys.json"

func main() {
	t := flag.String("t", "", "type of app")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		panic(fmt.Sprintf("only one arg is premitted, either init or index, got %v", args))
	}
	write := func(fileName string, content []byte) {
		err := ioutil.WriteFile(fileName, content, 0644)
		if err != nil {
			panic(err)
		}
	}

	if args[0] == "init" {
		rbcsetting, allpks, _ := mock.InitPeers(1)
		bytes, err := json.Marshal(rbcsetting)
		if err != nil {
			panic(err)
		}
		write(rbcSetting, bytes)

		keys, err := json.Marshal(allpks)
		if err != nil {
			panic(err)
		}
		write(privateKeys, keys)
		return
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		panic("arg should be int")
	}

	rbcbytes, err := ioutil.ReadFile(rbcSetting)
	if err != nil {
		panic("should call init first")
	}
	rbcSetting := common.RBCSetting{}
	err = json.Unmarshal(rbcbytes, &rbcSetting)
	if err != nil {
		panic(err)
	}

	var keys []sign.PrivateKey
	keyBytes, err := ioutil.ReadFile(privateKeys)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(keyBytes, &keys)
	if err != nil {
		panic(err)
	}
	var g errgroup.Group
	logging.SetLevel(logging.DEBUG, "RBC")
	switch *t {
	case "leader":
		leaderApp := transaction.NewLeader(1, "pstl")
		l, s, err := mock.NewLeader(leaderApp, keys[0], rbcSetting, &g)
		defer s()
		if err != nil {
			panic(err)
		}
		leaderApp.SetRBCLeader(l)
		handleLeaderUserInput(leaderApp)

	case "follower":
		followerApp := transaction.NewFollower(fmt.Sprintf("pstf%d", i))
		err, s := mock.NewFollower(followerApp, i, keys[i], rbcSetting, &g)
		defer s()
		if err != nil {
			panic(err)
		}
		handleFollowerUserInput(followerApp)
	default:
		panic("not supported")
	}
}

type handler interface {
	GetTransactionStatus(string) pb.TransactionStatus
}

type rbcHandler interface {
	SetMode(int)
}

func parseCommand(userInput string, ledger *transaction.Ledger, h handler, rh rbcHandler) (string, [][]byte) {
	// three types of commands
	deposit := regexp.MustCompile(`(?i)deposit (\d+)(\.\d+)? (?i)to (\S+)(?:\s*)?`)
	transfer := regexp.MustCompile(`(?i)transfer (\d+)(\.\d+)? (?i)from (\S+)(?:\s*) (?i)to (\S+)(?:\s*)`)
	getStatus := regexp.MustCompile(`(?i)status (\S+)`)
	getBalance := regexp.MustCompile(`(?i)balance (\S+)`)
	setLevel := regexp.MustCompile(`(?i)level (?i)(INFO|DEBUG)`)
	byzantineMode := regexp.MustCompile(`(?i)mode (0|1|2|3|4)`)

	dep := deposit.FindSubmatch([]byte(userInput))
	trans := transfer.FindSubmatch([]byte(userInput))
	stat := getStatus.FindSubmatch([]byte(userInput))
	blc := getBalance.FindSubmatch([]byte(userInput))
	level := setLevel.FindSubmatch([]byte(userInput))
	mode := byzantineMode.FindSubmatch([]byte(userInput))
	switch {
	case len(dep) != 0:
		return "deposit", dep
	case len(trans) != 0:
		return "transfer", trans
	case len(stat) != 0:
		res := h.GetTransactionStatus(string(stat[1]))
		fmt.Println("status: " + res.String())
		return "handled", nil
	case len(blc) != 0:
		act := string(blc[1])
		if _, ok := ledger.Accounts[act]; !ok {
			fmt.Println("account not exist")
			return "handled", nil
		}
		fmt.Printf("%d.%d\n", ledger.Accounts[act]/100, ledger.Accounts[act]%100)
		return "handled", nil
	case len(level) != 0:
		if strings.ToLower(string(level[1])) == "info" {
			logging.SetLevel(logging.INFO, "RBC")
			fmt.Println("Level set to INFO")
			return "handled", nil
		}
		logging.SetLevel(logging.DEBUG, "RBC")
		fmt.Println("Level set to DEBUG")
		return "handled", nil
	case len(mode) != 0:
		m, _ := strconv.Atoi(string(mode[1]))
		rh.SetMode(m)
		fmt.Printf("Set byzantine mode to %d\n", m)
		return "handled", nil
	default:
		return "unknown", nil
	}
}

func parseNum(dollarMatch, centsMatch []byte) (int, int, error) {
	dollar, err := strconv.Atoi(string(dollarMatch))
	if err != nil {
		return 0, 0, errors.Wrap(err, "wrong dollar format:"+string(dollarMatch))
	}
	var cents int
	if len(centsMatch) == 0 {
		cents = 0
	} else {
		cents, err = strconv.Atoi(string(centsMatch[1:]))
		if err != nil {
			return 0, 0, errors.Wrap(err, "wrong cents format:"+string(centsMatch))
		}
	}
	return dollar, cents, nil
}

func handleLeaderUserInput(l *transaction.Leader) {
	for {
		var userInput string
		reader := bufio.NewReader(os.Stdin)
		userInput, _ = reader.ReadString('\n')

		switch t, sub := parseCommand(userInput, l.Ledger, l, l.Leader); t {
		case "deposit", "transfer":
			if l == nil {
				fmt.Println("deposit and transfer can only be proposed by leader")
				continue
			}
			dollar, cents, err := parseNum(sub[1], sub[2])
			if err != nil {
				fmt.Println("wrong money format", err)
				continue
			}

			var id string
			if t == "deposit" {
				id, err = l.ProposeDeposit(string(sub[3]), dollar, cents)
			} else {
				id, err = l.ProposeTransfer(string(sub[3]), string(sub[4]), dollar, cents)
			}

			if err != nil {
				fmt.Printf("can't propose the %s: %s\n", t, err.Error())
				continue
			}
			fmt.Println(color.HiCyanString("%s proposed, txnID: %s", t, id))
		case "handled":
		default:
			fmt.Println("unsupported command")
		}
	}
}

func handleFollowerUserInput(f *transaction.Follower) {
	for {
		var userInput string
		reader := bufio.NewReader(os.Stdin)
		userInput, _ = reader.ReadString('\n')

		switch t, _ := parseCommand(userInput, f.Ledger, f, &f.Follower); t {
		case "handled":
		default:
			fmt.Println("unsupported command")
		}
	}
}
