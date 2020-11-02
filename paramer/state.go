package paramer

import (

	"github.com/tendermint/tendermint/crypto"
	kf "github.com/tendermint/tendermint/crypto/secp256k1"
	"ufc/lib"
)

var (
	Issuer = kf.GenPrivKey()
	SYSTEM_ISSUER = crypto.Address("KING_OF_TOKEN")
)


func InitWallet(){
	wallet := lib.NewWallet()
	wallet.GenPrivKey("issuer")
	wallet.GenPrivKey("michael")
	wallet.GenPrivKey("britney")
	wallet.Save("./wallet")
}