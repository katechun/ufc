package tx

import (
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/rpc/client/http"
	"strconv"
	"strings"
	"ubc/lib"
)

func Issue(cli *http.HTTP, to string, value int) error {
	if to == "" {
		return errors.New("Account addr is null!")
	}
	wallet := lib.LoadWallet()
	tx := lib.NewTx(lib.NewIssuePayload(
		wallet.GetAddress("issuer"),
		wallet.GetAddress(to),
		value))
	//fmt.Println("tx.PubKey:", tx.PubKey)
	//fmt.Println("tx.payload:", tx.Payload)
	//fmt.Println("tx.sequence:", tx.Sequence)
	_ = tx.Sign(wallet.GetPrivKey("issuer"))
	//fmt.Println("tx.signature:", tx.Signature)

	bz, err := lib.MarshalBinary(tx)
	if err != nil {
		panic(err)
	}
	_, err = cli.BroadcastTxCommit(bz)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	return nil
	//fmt.Printf("issue ret => %+v\n", ret)
}

func Transfer(cli *http.HTTP, from string, to string, val string) error {
	if from == "" || to == "" {
		return errors.New("Account addr is null!")
	}

	wallet := lib.LoadWallet()
	b, error := strconv.Atoi(val)

	v := QueryVal(from, cli)

	if v < b {
		return errors.New("余额不足")
	}

	if error != nil {
		return errors.New("输入错误")
	}

	tx := lib.NewTx(lib.NewTransferPayload(
		wallet.GetAddress(from),
		wallet.GetAddress(to),
		b))
	_ = tx.Sign(wallet.GetPrivKey(from))
	fmt.Println(tx)
	bz, err := lib.MarshalBinary(tx)
	if err != nil {
		panic(err)
	}
	ret, err := cli.BroadcastTxCommit(bz)
	if err != nil {
		panic(err)
	}
	fmt.Printf("issue ret => %+v\n", ret)
	return nil
}

func Query(label string, cli *http.HTTP) {
	wallet := lib.LoadWallet()
	//fmt.Println(wallet.GetAddress(label))
	ret, err := cli.ABCIQuery("", wallet.GetAddress(label))
	if err != nil {
		panic(err)
	}
	//fmt.Printf("ret => %+v\n", ret)

	fmt.Println(string(ret.Response.GetLog()))
}

func Addtx(cli *http.HTTP, key string) error {

	ret, err := cli.BroadcastTxCommit([]byte(key))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Add tx ret => %+v\n", ret)
	return nil
}

func QueryVal(label string, cli *http.HTTP) int {
	wallet := lib.LoadWallet()
	//fmt.Println(wallet.GetAddress(label))
	ret, err := cli.ABCIQuery("", wallet.GetAddress(label))
	if err != nil {
		panic(err)
	}
	//fmt.Printf("ret => %+v\n", ret)

	logs := string(ret.Response.GetLog())
	str := strings.Split(logs, "=>")
	str1 := strings.Replace(str[1], " ", "", -1)
	b, error := strconv.Atoi(str1)
	if error != nil {
		fmt.Println("输入错误")
		panic(error)
	}

	return b

}
