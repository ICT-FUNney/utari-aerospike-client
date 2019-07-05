package utari

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	aero "github.com/aerospike/aerospike-client-go"
)

// blockテーブル用のkeyを取得する関数
func getBlockKey(hash string) (*aero.Key, error) {
	namespace := GetAerospikeNamespace()
	table := GetAerospikeBlockTable()
	key, err := aero.NewKey(namespace, table, hash)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// transactionテーブル用のkeyを取得する関数
func getTransactionKey(hash string) (*aero.Key, error) {
	namespace := GetAerospikeNamespace()
	table := GetAerospikeTxTable()
	key, err := aero.NewKey(namespace, table, hash)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func getBalanceKey(address string) (*aero.Key, error) {
	namespace := GetAerospikeNamespace()
	table := "balance"
	key, err := aero.NewKey(namespace, table, address)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func getUnsettledKey(address string) (*aero.Key, error) {
	namespace := GetAerospikeNamespace()
	table := "unsettled"
	key, err := aero.NewKey(namespace, table, address)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GetHash はkeyとして利用するhash値を取得する関数
func GetHash(v interface{}) string {
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
func blockToBinMap(b Block) aero.BinMap {
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
func transactionToBinMap(t Transaction) aero.BinMap {
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
func binMapToBlock(record *aero.Record) (Block, error) {
	var block Block
	binMap := record.Bins

	// Idの型アサーション
	id, ok := binMap["Id"].(string)
	if !ok {
		return Block{}, fmt.Errorf("failed Id assertion")
	}
	block.Id = id

	// Versionの型アサーション
	version, ok := binMap["Version"].(int)
	if !ok {
		return Block{}, fmt.Errorf("failed Version assertion")
	}
	block.Version = int32(version)

	// Prehashの型アサーション
	prehash, ok := binMap["Prehash"].(string)
	if !ok {
		return Block{}, fmt.Errorf("failed Prehash assertion")
	}
	block.Prehash = prehash

	// Merklerootの型アサーション
	merkleroot, ok := binMap["Merkleroot"].(string)
	if !ok {
		return Block{}, fmt.Errorf("failed Merkleroot assertion")
	}
	block.Merkleroot = merkleroot

	// Timestampの型アサーション
	timestamp, ok := binMap["Timestamp"].(string)
	if !ok {
		return Block{}, fmt.Errorf("failed Timestamp assertion")
	}
	block.Timestamp = timestamp

	// Levelの型アサーション
	level, ok := binMap["Level"].(string)
	if !ok {
		return Block{}, fmt.Errorf("failed Level assertion")
	}
	block.Level = level

	// Nonceの型アサーション
	nonce, ok := binMap["Nonce"].(int)
	if !ok {
		return Block{}, fmt.Errorf("failed Nonce assertion")
	}
	block.Nonce = uint32(nonce)

	// Sizeの型アサーション
	size, ok := binMap["Size"].(int)
	if !ok {
		return Block{}, fmt.Errorf("failed Size assertion")
	}
	block.Size = int64(size)

	// Txcountの型アサーション
	txcount, ok := binMap["Txcount"].(int)
	if !ok {
		return Block{}, fmt.Errorf("failed Txcount assertion")
	}
	block.Txcount = int64(txcount)

	// TxidListの型アサーション
	var txidList []string
	// まずはスライスの型アサーション
	interfaceSlice, ok := binMap["TxidList"].([]interface{})
	if !ok {
		return Block{}, fmt.Errorf("failed TxidList assertion")
	}

	// スライスの中身を型アサーション
	for _, value := range interfaceSlice {
		txid, ok := value.(string)
		if !ok {
			return Block{}, fmt.Errorf("failed TxidList assertion")
		}
		txidList = append(txidList, txid)
	}
	block.TxidList = txidList

	return block, nil
}

// binmap to Transaction構造体
func binMapToTransaction(record *aero.Record) (Transaction, error) {
	var tx Transaction
	binMap := record.Bins

	// txidの型アサーション
	txid, ok := binMap["Txid"].(string)
	if !ok {
		return Transaction{}, fmt.Errorf("failed Txid assertion")
	}
	tx.Txid = txid

	// outputの型アサーション
	output, ok := binMap["Output"].(string)
	if !ok {
		return Transaction{}, fmt.Errorf("failed output assertion")
	}
	tx.Output = output

	// Inputの型アサーション
	input, ok := binMap["Input"].(string)
	if !ok {
		return Transaction{}, fmt.Errorf("failed input assertion")
	}
	tx.Input = input

	// Amountの型アサーション
	amountFloat, ok := binMap["Amount"].(float64)
	if !ok {
		amountInt, ok := binMap["Amount"].(int)
		if !ok {
			return Transaction{}, fmt.Errorf("failed Amount assertion")
		}
		amountFloat = float64(amountInt)
	}
	tx.Amount = amountFloat

	// Timestampの型アサーション
	timestamp, ok := binMap["Timestamp"].(string)
	if !ok {
		return Transaction{}, fmt.Errorf("failed Timestamp assertion")
	}
	tx.Timestamp = timestamp

	// Signの型アサーション
	sign, ok := binMap["Sign"].(string)
	if !ok {
		return Transaction{}, fmt.Errorf("failed sign assertion")
	}
	tx.Sign = sign

	// Pubkeyの型アサーション
	pubkey, ok := binMap["Pubkey"].(string)
	if !ok {
		return Transaction{}, fmt.Errorf("failed pubkey assertion")
	}
	tx.Pubkey = pubkey

	return tx, nil
}

func binMapToBalance(record *aero.Record) (Balance, error) {
	var balance Balance
	binMap := record.Bins

	balanceFloat, ok := binMap["Balance"].(float64)
	if !ok {
		balanceInt, ok := binMap["Balance"].(int)
		if !ok {
			return Balance{}, fmt.Errorf("failed Balance assertion")
		}
		balanceFloat = float64(balanceInt)
	}
	balance.Balance = balanceFloat

	address, ok := binMap["Address"].(string)
	if !ok {
		return Balance{}, fmt.Errorf("failed Address assertion")
	}
	balance.Address = address
	return balance, nil
}

func balanceToBinMap(b Balance) aero.BinMap {
	return aero.BinMap{
		"Address": b.Address,
		"Balance": b.Balance,
	}
}

func binMapToUnsettled(record *aero.Record) (Unsettled, error) {
	var unsettled Unsettled
	binMap := record.Bins

	unsettledFloat, ok := binMap["Unsettled"].(float64)
	if !ok {
		unsettledInt, ok := binMap["Unsettled"].(int)
		if !ok {
			return Unsettled{}, fmt.Errorf("failed Unssetled assertion")
		}
		unsettledFloat = float64(unsettledInt)
	}
	unsettled.Unsettled = unsettledFloat

	address, ok := binMap["Address"].(string)
	if !ok {
		return Unsettled{}, fmt.Errorf("failed Address assertion")
	}
	unsettled.Address = address
	return unsettled, nil
}

func unsettledToBinMap(b Unsettled) aero.BinMap {
	return aero.BinMap{
		"Address":   b.Address,
		"Unsettled": b.Unsettled,
	}
}
