package lib

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tendermint/tendermint/abci/example/code"
	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"strconv"
	"strings"
)

const (
	ValidatorSetChangePrefix string = "val:"
)

type Address = crypto.Address

var (
	//获取系统账户地址
	SYSTEM_ISSUER = crypto.Address("KING_OF_TOKEN")
	stateKey      = []byte("stateKey")
	//kvPairPrefixKey = []byte("kvPairKey:")
	//
	//ProtocolVersion version.Protocol = 0x1
)

//定义应用结构
type TokenApp struct {
	types.BaseApplication
	Accounts map[string]int
	state    State
}

type State struct {
	db      dbm.DB
	Size    int64  `json:"size"`
	Height  int64  `json:"height"`
	AppHash []byte `json:"app_hash"`
}

type PersistentUbcApplication struct {
	app *TokenApp

	// validator set
	ValUpdates []types.ValidatorUpdate

	valAddrToPubKeyMap map[string]types.PubKey

	logger log.Logger
}

//新建应用
//func NewTokenApp() *TokenApp{
//	return &TokenApp{Accounts:map[string]int{}}
//}

func NewTokenApp(dbDir string) *PersistentUbcApplication {
	name := "ubc"
	db, err := dbm.NewGoLevelDB(name, dbDir)
	if err != nil {
		panic(err)
	}

	state := loadState(db)

	return &PersistentUbcApplication{
		app:                &TokenApp{state: state, Accounts: make(map[string]int)},
		valAddrToPubKeyMap: make(map[string]types.PubKey),
		logger:             log.NewNopLogger(),
	}
}

func loadState(db dbm.DB) State {
	var state State
	state.db = db
	stateBytes, err := db.Get(stateKey)
	if err != nil {
		panic(err)
	}
	if len(stateBytes) == 0 {
		return state
	}
	err = json.Unmarshal(stateBytes, &state)
	if err != nil {
		panic(err)
	}
	return state
}

func (app *PersistentUbcApplication) Info(req types.RequestInfo) types.ResponseInfo {
	res := app.app.Info(req)
	res.LastBlockHeight = app.app.state.Height
	res.LastBlockAppHash = app.app.state.AppHash
	return res
}

func (app *PersistentUbcApplication) SetOption(req types.RequestSetOption) types.ResponseSetOption {
	return app.app.SetOption(req)
}

//查询操作
func (app *PersistentUbcApplication) Query(req types.RequestQuery) (rsp types.ResponseQuery) {
	//fmt.Println("crypto address:",req.Data)
	//获取账户地址
	addr := crypto.Address(req.Data)
	//rsp.Key = req.Data
	//获取账户信息并进行序列化
	rsp.Value, _ = MarshalBinary(app.app.Accounts[addr.String()])
	rsp.Key = req.Data
	rsp.Log = fmt.Sprintf("Query : U-%v => %v", addr, app.app.Accounts[addr.String()])
	//fmt.Println(rsp.Value)
	//rsp.Value=app.Accounts[addr.String()]
	return
}

func (app *PersistentUbcApplication) CheckTx(raw types.RequestCheckTx) (rsp types.ResponseCheckTx) {
	prefix := raw.Tx[0:len(ValidatorSetChangePrefix)]
	if string(prefix) == "val:" {
		return
	}

	//tx,err := app.decodeTx(raw.Tx)
	tx := Tx{}
	err := UnmarshalBinary(raw.Tx, &tx)
	if err != nil {
		rsp.Code = 1
		rsp.Log = "decode error"
	}

	if !tx.Verify() {
		rsp.Code = 2
		rsp.Log = "verify failed"
		return
	}
	return
}

//发布事务
func (app *PersistentUbcApplication) DeliverTx(raw types.RequestDeliverTx) (rsp types.ResponseDeliverTx) {

	//return  types.ResponseDeliverTx{
	//	Code:1,
	//	Log: fmt.Sprintf("Expected 'pubkey!power'. Got %v", raw.Tx),
	//}
	// if it starts with "val:", update the validator set
	// format is "val:pubkey!power"
	if isValidatorTx(raw.Tx) {
		// update validators in the merkle tree
		// and in app.ValUpdates
		app.logger.Info("exec add validator node!")
		return app.execValidatorTx(raw.Tx)
	}

	tx := Tx{}
	_ = UnmarshalBinary(raw.Tx, &tx)

	// otherwise, update the key-value store
	return func() (rsp types.ResponseDeliverTx) {

		app.logger.Info("tx......")
		tx := Tx{}
		_ = UnmarshalBinary(raw.Tx, &tx)
		switch tx.Payload.GetType() {
		case "issue":
			pld := tx.Payload.(*IssuePayload)
			err := app.Issue(pld.Issuer, pld.To, pld.Value)
			if err != nil {
				rsp.Log = err.Error()
			}
			rsp.Info = "issue tx applied"
		case "transfer":
			pld := tx.Payload.(*TransferPayload)
			err := app.Transfer(pld.From, pld.To, pld.Value)
			if err != nil {
				rsp.Log = err.Error()
			}
			rsp.Info = "transger tx applied"
		}

		return
	}()

}

// Commit will panic if InitChain was not called
func (app *PersistentUbcApplication) Commit() types.ResponseCommit {
	return app.app.Commit()
}

// Save the validators in the merkle tree
func (app *PersistentUbcApplication) InitChain(req types.RequestInitChain) types.ResponseInitChain {
	//var state map[string]int64
	//json.Unmarshal(req.AppStateBytes,&state)
	//app.Value,_ = state["priv_validator_key_file"]
	//app.Version = 0
	//

	for _, v := range req.Validators {
		r := app.updateValidator(v)
		if r.IsErr() {
			app.logger.Error("Error updating validators", "r", r)
		}
	}
	return types.ResponseInitChain{}
}

// Track the block hash and header information
func (app *PersistentUbcApplication) BeginBlock(req types.RequestBeginBlock) types.ResponseBeginBlock {
	// reset valset changes
	app.ValUpdates = make([]types.ValidatorUpdate, 0)

	for _, ev := range req.ByzantineValidators {
		if ev.Type == tmtypes.ABCIEvidenceTypeDuplicateVote {
			// decrease voting power by 1
			if ev.TotalVotingPower == 0 {
				continue
			}
			app.updateValidator(types.ValidatorUpdate{
				PubKey: app.valAddrToPubKeyMap[string(ev.Validator.Address)],
				Power:  ev.TotalVotingPower - 1,
			})
		}

	}

	//app.updateValidator(types.ValidatorUpdate{
	//	PubKey: LoadLocalPubKey(),
	//	Power: 10,
	//})

	return types.ResponseBeginBlock{}
}

//type PubKey1 struct {
//	Type string `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
//	Value string `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
//
//}
//// FilePVKey stores the immutable part of PrivValidator.
//type PrivValidatorKey struct {
//	Address Address  `json:"address"`
//	PubKey  PubKey1  `json:"pub_key"`
//	PrivKey PubKey1 `json:"priv_key"`
//}
//
//
//
//func LoadLocalPubKey() types.PubKey{
//	//validator_node_file:="..\\config\\priv_validator_key.json"
//	cfg := config.DefaultBaseConfig()
//	fmt.Println("..\\"+cfg.PrivValidatorKeyFile())
//	bz,err := ioutil.ReadFile("..\\"+cfg.PrivValidatorKeyFile())
//	if err != nil { panic(err)}
//	pvk:=PrivValidatorKey{}
//
//	err = codec.UnmarshalJSON(bz,&pvk)
//	fmt.Println(pvk)
//	if err != nil { panic(err)}
//	fmt.Println(pvk.PubKey.Value)
//	//pubkey2:=types.PubKey{Type:pvk.PubKey.Type,Data:pvk.PubKey.Value,}
//	//fmt.Println(pubkey2)
//
//	//return pubkey2
//	return types.PubKey{}
//}

// Update the validator set
func (app *PersistentUbcApplication) EndBlock(req types.RequestEndBlock) types.ResponseEndBlock {
	return types.ResponseEndBlock{ValidatorUpdates: app.ValUpdates}
}

//转账交易
func (app *PersistentUbcApplication) Transfer(from, to crypto.Address, value int) error {
	//如果账号余额不够就抛出错误
	if app.app.Accounts[from.String()] < value {
		return errors.New("balance low")
	}

	fmt.Println("before transger from app.app.Accounts[to.String()]", app.app.Accounts[from.String()])
	fmt.Println("before transger to app.app.Accounts[to.String()]", app.app.Accounts[to.String()])

	app.app.Accounts[from.String()] -= value
	app.app.Accounts[to.String()] += value

	fmt.Println("after transger from app.app.Accounts[to.String()]", app.app.Accounts[from.String()])
	fmt.Println("after transger to app.app.Accounts[to.String()]", app.app.Accounts[to.String()])

	return nil

}

//func (app *PersistentUbcApplication)decodeTx(raw []byte)(*Tx,error){
//	var tx Tx
//	err := lib.(raw,&tx)
//	fmt.Println(tx)
//	return &tx,err
//}

//发行货币 向系统账号增加货币数量
func (app *PersistentUbcApplication) Issue(issuer, to crypto.Address, value int) error {
	//导入钱包信息
	wallet := LoadWallet()
	//获取系统账号地址
	SYSTEM_ISSUER = wallet.GetAddress("issuer")

	//判断发行地址和系统地址是否一致
	if !bytes.Equal(issuer, SYSTEM_ISSUER) {
		return errors.New("invalid issuer")
	}

	fmt.Println("app.app.Accounts[to.String()]", app.app.Accounts[to.String()])
	fmt.Println("before transger to app.app.Accounts[to.String()]", app.app.Accounts[to.String()])

	//把发行的系统账号累加发行货币数量
	app.app.Accounts[to.String()] += value
	fmt.Println("app.app.Accounts[to.String()]", app.app.Accounts[to.String()])
	fmt.Println("after transger to app.app.Accounts[to.String()]", app.app.Accounts[to.String()])

	return nil
}

func (app *PersistentUbcApplication) Dump() {
	fmt.Printf("state => %v\n", app.app.Accounts)
}

// format is "val:pubkey!power"
// pubkey is a base64-encoded 32-byte ed25519 key
func (app *PersistentUbcApplication) execValidatorTx(tx []byte) types.ResponseDeliverTx {
	tx = tx[len(ValidatorSetChangePrefix):]

	//get the pubkey and power
	pubKeyAndPower := strings.Split(string(tx), "!")
	if len(pubKeyAndPower) != 2 {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Expected 'pubkey!power'. Got %v", pubKeyAndPower)}
	}

	pubkeyS, powerS := pubKeyAndPower[0], pubKeyAndPower[1]
	//pubkeyS, powerS := pubKeyAndPower, "10"

	// decode the pubkey
	pubkey, err := base64.StdEncoding.DecodeString(pubkeyS)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Pubkey (%s) is invalid base64", pubkeyS)}
	}

	// decode the power
	power, err := strconv.ParseInt(powerS, 10, 64)
	if err != nil {
		return types.ResponseDeliverTx{
			Code: code.CodeTypeEncodingError,
			Log:  fmt.Sprintf("Power (%s) is not an int", powerS)}
	}

	// update
	return app.updateValidator(types.Ed25519ValidatorUpdate(pubkey, power))
}

// add, update, or remove a validator
func (app *PersistentUbcApplication) updateValidator(v types.ValidatorUpdate) (rsp types.ResponseDeliverTx) {
	key := []byte("val:" + string(v.PubKey.Data))

	pubkey := ed25519.PubKeyEd25519{}
	copy(pubkey[:], v.PubKey.Data)

	if v.Power == 0 {
		// remove validator
		hasKey, err := app.app.state.db.Has(key)
		if err != nil {
			panic(err)
		}
		if !hasKey {
			pubStr := base64.StdEncoding.EncodeToString(v.PubKey.Data)
			return types.ResponseDeliverTx{
				Code: code.CodeTypeUnauthorized,
				Log:  fmt.Sprintf("Cannot remove non-existent validator %s", pubStr)}
		}
		app.app.state.db.Delete(key)
		delete(app.valAddrToPubKeyMap, string(pubkey.Address()))
	} else {
		// add or update validator
		value := bytes.NewBuffer(make([]byte, 0))
		if err := types.WriteMessage(&v, value); err != nil {
			return types.ResponseDeliverTx{
				Code: code.CodeTypeEncodingError,
				Log:  fmt.Sprintf("Error encoding validator: %v", err)}
		}
		app.app.state.db.Set(key, value.Bytes())
		app.valAddrToPubKeyMap[string(pubkey.Address())] = v.PubKey
	}

	// we only update the changes array if we successfully updated the tree
	app.ValUpdates = append(app.ValUpdates, v)

	return types.ResponseDeliverTx{Code: code.CodeTypeOK}
}

func (app *PersistentUbcApplication) Validators() (validators []types.ValidatorUpdate) {
	itr, err := app.app.state.db.Iterator(nil, nil)
	if err != nil {
		panic(err)
	}
	for ; itr.Valid(); itr.Next() {
		if isValidatorTx(itr.Key()) {
			validator := new(types.ValidatorUpdate)
			err := types.ReadMessage(bytes.NewBuffer(itr.Value()), validator)
			if err != nil {
				panic(err)
			}
			validators = append(validators, *validator)
		}
	}
	return
}

func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), ValidatorSetChangePrefix)
}
