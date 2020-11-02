package lib

import (
	"bytes"
	"github.com/tendermint/tendermint/crypto"
	"time"
)


//事务结构
type Tx struct{
	Payload Payload
	Signature []byte
	PubKey crypto.PubKey
	Sequence int64
}

//创建新事务
func NewTx(payload Payload)*Tx{
	return &Tx{Payload:payload,Sequence: time.Now().Unix()}
}

//验证
func (tx *Tx) Verify() bool {
	//获取转账者地址
	signer := tx.Payload.GetSigner()
	//根据公钥生成转账者地址
	signerFromKey := tx.PubKey.Address()
	//判断获取的地址和生成地址是否一致
	if !bytes.Equal(signer,signerFromKey){
		return false
	}
	data := tx.Payload.GetSignBytes()
	sig := tx.Signature
	valid := tx.PubKey.VerifyBytes(data,sig)
	if !valid{
		return false
	}
	return true
}

//进行签名
func (tx *Tx)Sign(priv crypto.PrivKey)error{
	data := tx.Payload.GetSignBytes()
	var err error
	tx.Signature,err = priv.Sign(data)
	tx.PubKey = priv.PubKey()
	return err
}


























