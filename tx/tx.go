package tx

import (
	"fmt"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/rpc/client/http"
	"ufc/lib"
)

var codec = amino.NewCodec()

func Issue(cli *http.HTTP) {
	fmt.Println("1")
	wallet := lib.LoadWallet("./wallet")
	tx := lib.NewTx(lib.NewIssuePayload(
		wallet.GetAddress("issuer"),
		wallet.GetAddress("michael"),
		1000))
	_ = tx.Sign(wallet.GetPrivKey("issuer"))
	bz, err := lib.MarshalBinary(tx)
	if err != nil {
		panic(err)
	}
	fmt.Println(bz)
	ret, err := cli.BroadcastTxCommit(bz)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Printf("issue ret => %+v\n", ret)
}

func Transfer(cli *http.HTTP) {
	wallet := lib.LoadWallet("./wallet")
	tx := lib.NewTx(lib.NewTransferPayload(
		wallet.GetAddress("michael"),
		wallet.GetAddress("britney"),
		100))
	_ = tx.Sign(wallet.GetPrivKey("michael"))
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
}

func Query(label string, cli *http.HTTP) {
	wallet := lib.LoadWallet("./wallet")
	fmt.Println(wallet.GetAddress(label))
	ret, err := cli.ABCIQuery("", wallet.GetAddress(label))
	if err != nil {
		panic(err)
	}
	fmt.Printf("ret => %+v\n", ret)
	fmt.Printf("%v", ret.Response.Value)
}
