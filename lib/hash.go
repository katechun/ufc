package lib

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
	"github.com/tendermint/tendermint/crypto/merkle"
	"strconv"
)

type Balance int

func (b Balance) Hash() []byte{
	v,_ := codec.MarshalBinaryBare(b)
	return tmhash.Sum(v)
}

func (app *TokenApp) stateToHasherMap() map[string][]byte {
	hashers := map[string][]byte{}
	for addr,val := range app.Accounts {
		hashers[addr]=[]byte(strconv.Itoa(val))
	}
	return hashers
}

func (app *TokenApp) getRootHash() []byte {
	hashers := app.stateToHasherMap()
	return merkle.SimpleHashFromMap(hashers)
}

func (app *TokenApp) getProofBytes(addr string) []byte {
	hashers := app.stateToHasherMap()
	_,proofs,_ := merkle.SimpleProofsFromMap(hashers)
	bz,err := codec.MarshalBinaryBare(proofs[addr])
	if err != nil  { return  []byte{} }
	return bz
}