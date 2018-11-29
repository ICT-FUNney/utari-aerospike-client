package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	aero "github.com/aerospike/aerospike-client-go"
	"github.com/youtangai/aerospike-sample/model"
)

const (
	AEROSPIKE_HOST        = "127.0.0.1"
	AEROSPIKE_PORT        = 3000
	AEROSPIKE_NAMESPACE   = "test"
	AEROSPIKE_TX_TABLE    = "TxTable"
	AEROSPIKE_BLOCL_TABLE = "BlockTable"
)

type aeroSpikeClient struct {
	client *aero.Client
}

type IAeroSpikeClinet interface {
	PutBlock(model.Block) error
	PutTransaction(model.Transaction) error
	GetBlock(string) (model.Block, error)
	GetTransaction(string) (model.Transaction, error)
	DeleteBlock(string) error
	DeleteTransaction(string) error
}

func NewAeroSpikeClient(host string, port int) (IAeroSpikeClinet, error) {
	cli, err := aero.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return aeroSpikeClient{
		client: cli,
	}, nil
}

func main() {
	// クライアントを取得する
	client, err := NewAeroSpikeClient(AEROSPIKE_HOST, AEROSPIKE_PORT)
	if err != nil {
		panic(err)
	}

	// ダミーデータ作成
	block := model.Block{
		Id:         "testid",
		Version:    12,
		Prehash:    "testprehash",
		Merkleroot: "testmerkleroot",
		Timestamp:  "test_timestamp",
		Level:      "test_level",
		Nonce:      123,
		Size:       1234,
		Txcount:    12345,
		TxidList:   []string{"testid1", "testid2"},
	}
	tx := model.Transaction{
		Txid:      "testtxid",
		Output:    "testoutput",
		Input:     "testinput",
		Amount:    12.34,
		Timestamp: "test_timestamp",
		Sign:      "test_sign",
		Pubkey:    "test_pubkey",
	}

	// データの格納
	err = client.PutBlock(block)
	if err != nil {
		panic(err)
	}
	err = client.PutTransaction(tx)
	if err != nil {
		panic(err)
	}

	// keyとして必要なハッシュ値を取得
	blockHash := getHash(block)
	txHash := getHash(tx)

	// レコードの取得
	blockRecv, err := client.GetBlock(blockHash)
	if err != nil {
		panic(err)
	}
	txRecv, err := client.GetTransaction(txHash)
	if err != nil {
		panic(err)
	}

	// データの確認
	fmt.Printf("block:%v\n", blockRecv)
	fmt.Printf("transaction:%v\n", txRecv)

	// データの削除
	err = client.DeleteBlock(blockHash)
	err = client.DeleteTransaction(txHash)
}

func (a aeroSpikeClient) PutBlock(block model.Block) error {
	// hash値の取得
	hash := getHash(block)

	// aerospike用のkey構造体を取得
	key, err := getBlockKey(hash)
	if err != nil {
		return err
	}

	// dataをbinmap(aerospikeに挿入可能な形)へ変換
	data := blockToBinMap(block)

	// データの格納
	err = a.client.Put(nil, key, data)
	if err != nil {
		return err
	}
	return nil
}

func (a aeroSpikeClient) PutTransaction(tx model.Transaction) error {
	// hash値の取得
	hash := getHash(tx)

	// aerospike用のkey構造体を取得
	key, err := getTransactionKey(hash)
	if err != nil {
		return err
	}

	// dataをbinmap(aerospikeに挿入可能な形)へ変換
	data := transactionToBinMap(tx)

	// データの格納
	err = a.client.Put(nil, key, data)
	if err != nil {
		return err
	}
	return nil
}

func (a aeroSpikeClient) GetBlock(hash string) (model.Block, error) {
	key, err := getBlockKey(hash)
	if err != nil {
		return model.Block{}, err
	}
	// レコードの取得
	record, err := a.client.Get(nil, key)
	if err != nil {
		return model.Block{}, err
	}

	// binmap to block
	block, err := binMapToBlock(record)
	if err != nil {
		return model.Block{}, err
	}

	return block, nil
}

func (a aeroSpikeClient) GetTransaction(hash string) (model.Transaction, error) {
	key, err := getTransactionKey(hash)
	if err != nil {
		return model.Transaction{}, err
	}

	// レコードの取得
	record, err := a.client.Get(nil, key)
	if err != nil {
		return model.Transaction{}, err
	}

	// binmap to tx
	tx, err := binMapToTransaction(record)
	if err != nil {
		return model.Transaction{}, err
	}

	return tx, nil
}

func (a aeroSpikeClient) DeleteBlock(hash string) error {
	key, err := getBlockKey(hash)
	_, err = a.client.Delete(nil, key)
	if err != nil {
		return err
	}
	return nil
}

func (a aeroSpikeClient) DeleteTransaction(hash string) error {
	key, err := getTransactionKey(hash)
	_, err = a.client.Delete(nil, key)
	if err != nil {
		return err
	}
	return nil
}

func getBlockKey(hash string) (*aero.Key, error) {
	key, err := aero.NewKey(AEROSPIKE_NAMESPACE, AEROSPIKE_BLOCL_TABLE, hash)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func getTransactionKey(hash string) (*aero.Key, error) {
	key, err := aero.NewKey(AEROSPIKE_NAMESPACE, AEROSPIKE_TX_TABLE, hash)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// keyとして利用するhash値を取得する関数
func getHash(v interface{}) string {
	// 構造体を[]byteに変換
	byteData, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	// ハッシュ関数にかける
	checksum := sha256.Sum256(byteData)
	// 文字列として取得する
	hash := fmt.Sprintf("%x", checksum)
	return hash
}

// Block構造体 to binmap
func blockToBinMap(b model.Block) aero.BinMap {
	return aero.BinMap{
		"Id":         b.Id,
		"Version":    b.Version,
		"Prehash":    b.Prehash,
		"Merkleroot": b.Merkleroot,
		"Timestamp":  b.Timestamp,
		"Level":      b.Level,
		"Nonce":      b.Nonce,
		"Size":       b.Size,
		"Txcount":    b.Txcount,
		"TxidList":   b.TxidList,
	}
}

// Transaction構造体 to binmap
func transactionToBinMap(t model.Transaction) aero.BinMap {
	return aero.BinMap{
		"Txid":      t.Txid,
		"Output":    t.Output,
		"Input":     t.Input,
		"Amount":    t.Amount,
		"Timestamp": t.Timestamp,
		"Sign":      t.Sign,
		"Pubkey":    t.Pubkey,
	}
}

// binmap to Block構造体
func binMapToBlock(record *aero.Record) (model.Block, error) {
	var block model.Block
	binMap := record.Bins

	// Idの型アサーション
	id, ok := binMap["Id"].(string)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Id assertion")
	}
	block.Id = id

	// Versionの型アサーション
	version, ok := binMap["Version"].(int)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Version assertion")
	}
	block.Version = int32(version)

	// Prehashの型アサーション
	prehash, ok := binMap["Prehash"].(string)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Prehash assertion")
	}
	block.Prehash = prehash

	// Merklerootの型アサーション
	merkleroot, ok := binMap["Merkleroot"].(string)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Merkleroot assertion")
	}
	block.Merkleroot = merkleroot

	// Timestampの型アサーション
	timestamp, ok := binMap["Timestamp"].(string)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Timestamp assertion")
	}
	block.Timestamp = timestamp

	// Levelの型アサーション
	level, ok := binMap["Level"].(string)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Level assertion")
	}
	block.Level = level

	// Nonceの型アサーション
	nonce, ok := binMap["Nonce"].(int)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Nonce assertion")
	}
	block.Nonce = uint32(nonce)

	// Sizeの型アサーション
	size, ok := binMap["Size"].(int)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Size assertion")
	}
	block.Size = int64(size)

	// Txcountの型アサーション
	txcount, ok := binMap["Txcount"].(int)
	if !ok {
		return model.Block{}, fmt.Errorf("failed Txcount assertion")
	}
	block.Txcount = int64(txcount)

	// TxidListの型アサーション
	var txidList []string
	// まずはスライスの型アサーション
	interfaceSlice, ok := binMap["TxidList"].([]interface{})
	if !ok {
		return model.Block{}, fmt.Errorf("failed TxidList assertion")
	}

	// スライスの中身を型アサーション
	for _, value := range interfaceSlice {
		txid, ok := value.(string)
		if !ok {
			return model.Block{}, fmt.Errorf("failed TxidList assertion")
		}
		txidList = append(txidList, txid)
	}
	block.TxidList = txidList

	return block, nil
}

// binmap to Transaction構造体
func binMapToTransaction(record *aero.Record) (model.Transaction, error) {
	var tx model.Transaction
	binMap := record.Bins

	// txidの型アサーション
	txid, ok := binMap["Txid"].(string)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed Txid assertion")
	}
	tx.Txid = txid

	// outputの型アサーション
	output, ok := binMap["Output"].(string)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed output assertion")
	}
	tx.Output = output

	// Inputの型アサーション
	input, ok := binMap["Input"].(string)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed input assertion")
	}
	tx.Input = input

	// Amountの型アサーション
	amount, ok := binMap["Amount"].(float64)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed Amount assertion")
	}
	tx.Amount = amount

	// Timestampの型アサーション
	timestamp, ok := binMap["Timestamp"].(string)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed Timestamp assertion")
	}
	tx.Timestamp = timestamp

	// Signの型アサーション
	sign, ok := binMap["Sign"].(string)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed sign assertion")
	}
	tx.Sign = sign

	// Pubkeyの型アサーション
	pubkey, ok := binMap["Pubkey"].(string)
	if !ok {
		return model.Transaction{}, fmt.Errorf("failed pubkey assertion")
	}
	tx.Pubkey = pubkey

	return tx, nil
}
