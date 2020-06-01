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
	switch *t{
	case "leader":
		mock.NewLeader(transaction.NewLeader(2, ""), keys[i], rbcSetting, &g)
		var input string
		var input2 string
		n, err := fmt.Scanln(&input, &input2)
		fmt.Println(n, err, input, "|", input2)

	case "follower":
		err, _ := mock.NewFollower(transaction.NewFollower(""), i, keys[i], rbcSetting, &g)
		if err != nil{
			panic(err)
		}
	default:
		panic("not supported")
	}
}

func handleUserInput(l *transaction.Leader){
	for {
		var userInput string
		reader := bufio.NewReader(os.Stdin)
		userInput, _ = reader.ReadString('\n')
		deposit := regexp.MustCompile(`(?i)deposit (\d+)(\.\d+)? (?i)to (\S+)`)
		transfer := regexp.MustCompile(`(?i)transfer (\d+)(\.\d+) (?i)from (\S+) (?i)to (\S+)`)
		dep := deposit.FindSubmatch([]byte(userInput))
		trans := transfer.FindSubmatch([]byte(userInput))

		if len(dep) != 0{
			dollar, err := strconv.Atoi(string(dep[1]))
			if err != nil{
				fmt.Println("wrong dollar format")
				continue
			}
			var cents int
			if len(dep[2]) == 0{
				cents = 0
			}else{
				cents, err = strconv.Atoi(string(dep[2]))
				if err != nil{
					fmt.Println("wrong cents format")
					continue
				}
			}

			trans, err := l.ProposeDeposit(string(dep[3]), dollar, cents)
			if err != nil{
				fmt.Println("can't propose the deposit: ", err.Error())
				continue
			}
			fmt.Println("deposit proposed, txnID: ", trans)
		}
		if len(dep) == 0 && len(trans) == 0 {
			fmt.Println("wrong command")
		}
		fmt.Println(dep, trans)
	}
}
