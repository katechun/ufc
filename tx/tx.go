package tx

import (
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/rpc/client/http"
	"ufc/lib"
)

func Issue(cli *http.HTTP, to string) error {
	if to == "" {
		return errors.New("Account addr is null!")
	}
	wallet := lib.LoadWallet()
	tx := lib.NewTx(lib.NewIssuePayload(
		wallet.GetAddress("issuer"),
		wallet.GetAddress(to),
		10000))
	fmt.Println("tx.PubKey:", tx.PubKey)
	fmt.Println("tx.payload:", tx.Payload)
	fmt.Println("tx.sequence:", tx.Sequence)
	_ = tx.Sign(wallet.GetPrivKey("issuer"))
	fmt.Println("tx.signature:", tx.Signature)

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

func Transfer(cli *http.HTTP, from string, to string) error {
	if from == "" || to == "" {
		return errors.New("Account addr is null!")
	}

	wallet := lib.LoadWallet()
	tx := lib.NewTx(lib.NewTransferPayload(
		wallet.GetAddress(from),
		wallet.GetAddress(to),
		1000))
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
	fmt.Println(wallet.GetAddress(label))
	ret, err := cli.ABCIQuery("", wallet.GetAddress(label))
	if err != nil {
		panic(err)
	}
	fmt.Printf("ret => %+v\n", ret)

	fmt.Println(string(ret.Response.GetLog()))
}
