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
	"github.com/gopricy/mao-bft/application/transaction"
	"github.com/gopricy/mao-bft/rbc/common"
	"github.com/gopricy/mao-bft/rbc/mock"
	"github.com/gopricy/mao-bft/rbc/sign"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
)
const rbcSetting = "rbc_setting.json"
const privateKeys = "private_keys.json"

func main() {
	t := flag.String("t", "", "type of app")
	flag.Parse()
	args := flag.Args()
	if len(args) != 1{
		panic(fmt.Sprintf("only one arg is premitted, either init or index, got %v", args))
	}
	write := func(fileName string, content []byte){
		err := ioutil.WriteFile(fileName, content, 0644)
		if err != nil{
			panic(err)
		}
	}

	if args[0] == "init"{
		rbcsetting, allpks, _ := mock.InitPeers(1)
		bytes, err := json.Marshal(rbcsetting)
		if err != nil{
			panic(err)
		}
		write(rbcSetting, bytes)

		keys, err := json.Marshal(allpks)
		if err != nil{
			panic(err)
		}
		write(privateKeys, keys)
		return
	}


	i, err := strconv.Atoi(args[0])
	if err != nil{
		panic("arg should be int")
	}

	rbcbytes, err := ioutil.ReadFile(rbcSetting)
	if err != nil{
		panic("should call init first")
	}
	rbcSetting := common.RBCSetting{}
	err = json.Unmarshal(rbcbytes, &rbcSetting)
	if err != nil{
		panic(err)
	}

	var keys []sign.PrivateKey
	keyBytes, err := ioutil.ReadFile(privateKeys)
	if err != nil{
		panic(err)
	}
	err = json.Unmarshal(keyBytes, &keys)
	if err != nil{
		panic(err)
	}
	var g errgroup.Group
	logging.SetLevel(logging.DEBUG, "RBC")
	switch *t{
	case "leader":
		leaderApp := transaction.NewLeader(1, "")
		l, s, err := mock.NewLeader(leaderApp, keys[0], rbcSetting, &g)
		if err != nil{
			panic(err)
		}
		defer s()
		leaderApp.SetRBCLeader(l)
		handleUserInput(leaderApp)

	case "follower":
		err, s := mock.NewFollower(transaction.NewFollower(""), i, keys[i], rbcSetting, nil)
		defer s()
		if err != nil{
			panic(err)
		}

	default:
		panic("not supported")
	}
}

func handleBalance(l *transaction.Follower){
	for{
		var userInput string
		reader := bufio.NewReader(os.Stdin)
		userInput, _ = reader.ReadString('\n')
		getBalance := regexp.MustCompile(`(?i)balance (\S+)`)

		blc := getBalance.FindSubmatch([]byte(userInput))

		if len(blc) == 0{
			fmt.Println("invalid command")
		}
		act := string(blc[1])

		fmt.Println(l.Ledger.Accounts[act])
	}
}

func handleUserInput(l *transaction.Leader){
	for {
		var userInput string
		reader := bufio.NewReader(os.Stdin)
		userInput, _ = reader.ReadString('\n')

		// three types of commands
		deposit := regexp.MustCompile(`(?i)deposit (\d+)(\.\d+)? (?i)to (\S+)`)
		transfer := regexp.MustCompile(`(?i)transfer (\d+)(\.\d+) (?i)from (\S+) (?i)to (\S+)`)
		getStatus := regexp.MustCompile(`(?i)status (\S+)`)
		getBalance := regexp.MustCompile(`(?i)balance (\S+)`)

		// l.GetTransactionStatus()

		dep := deposit.FindSubmatch([]byte(userInput))
		trans := transfer.FindSubmatch([]byte(userInput))
		stat := getStatus.FindSubmatch([]byte(userInput))
		blc := getBalance.FindSubmatch([]byte(userInput))

		if len(dep) == 0 && len(trans) == 0 && len(stat) == 0 && len(blc) == 0{
			fmt.Println("invalid command")
		}

		parseNum := func(dollarMatch, centsMatch []byte) (int, int, error){
			dollar, err := strconv.Atoi(string(dollarMatch))
			if err != nil {
				return 0, 0, errors.Wrap(err, "wrong dollar format")
			}
			var cents int
			if len(centsMatch) == 0 {
				cents = 0
			} else {
				cents, err = strconv.Atoi(string(centsMatch))
				if err != nil {
					return 0, 0, errors.Wrap(err,"wrong cents format")
				}
			}
			return dollar, cents, nil
		}

		if len(dep) != 0{
			dollar, cents, err := parseNum(dep[1], dep[2])
			id, err := l.ProposeDeposit(string(dep[3]), dollar, cents)
			if err != nil{
				fmt.Println("can't propose the deposit: ", err.Error())
				continue
			}
			fmt.Println("deposit proposed, txnID: ", id)
			continue
		}

		if len(trans) != 0{
			dollar, cents, err := parseNum(trans[1], trans[2])
			id, err := l.ProposeTransfer(string(trans[3]), string(trans[4]), dollar, cents)
			if err != nil{
				fmt.Println("can't propose the transfer: ", err.Error())
				continue
			}
			fmt.Println("transfer proposed, txnID: ", id)
			continue
		}

		if len(stat) != 0{
			res := l.GetTransactionStatus(string(stat[1]))
			fmt.Println("status: " + res.String())
			continue
		}

		act := string(blc[1])
		fmt.Println(l.Ledger.Accounts[act])

	}
}
