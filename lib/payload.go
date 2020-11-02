package lib

import (
	"encoding/json"
	"github.com/tendermint/tendermint/crypto"
)

//转账接口
type Payload interface{
	//获取转账人地址
	GetSigner() crypto.Address
	//序列化转账人信息
	GetSignBytes() []byte
	//获取转账类型，是发行货币还是普通转账交易
	GetType() string
}

//发行货币结构体
type IssuePayload struct {
	Issuer crypto.Address
	To crypto.Address
	Value int
}

//创建发行货币实例
func NewIssuePayload(issuer,to crypto.Address,value int)*IssuePayload{
	return &IssuePayload{issuer,to,value}
}


//获取系统账户签名者的账户地址
func (pld *IssuePayload)GetSigner() crypto.Address{
	return pld.Issuer
}

//获取系统账户签名字节类型
func (pld *IssuePayload)GetSignBytes()[]byte{
	bz,err := json.Marshal(pld)
	if err != nil {
		return []byte{}
	}

	return bz
}

//获取发行货币的账户类型
func (pld *IssuePayload)GetType() string{
	return "issue"
}



//普通交易结构体
type TransferPayload struct{
	From crypto.Address
	To crypto.Address
	Value int
}

//创建正常交易实例化
func NewTransferPayload(from,to crypto.Address,value int)*TransferPayload{
	return &TransferPayload{from,to,value}
}

//获取转账人地址
func (pld *TransferPayload)GetSigner() crypto.Address{
	return pld.From
}


//序列化转账人信息
func (pld *TransferPayload)GetSignBytes() []byte{
	bz,err := json.Marshal(pld)
	if err != nil {
		return []byte{}
	}
	return bz
}

//获取转账类型
func (pld *TransferPayload)GetType() string{
	return "transfer"
}


























