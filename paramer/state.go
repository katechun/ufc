package paramer

import (
	"ufc/lib"
)

//var (
//	Issuer = kf.GenPrivKey()
//	SYSTEM_ISSUER = crypto.Address("KING_OF_TOKEN")
//)

func InitWallet() {
	wallet := lib.NewWallet()
	wallet.GenPrivKey("issuer")
	wallet.GenPrivKey("michael")
	wallet.GenPrivKey("britney")
	wallet.Save("./wallet")
}
