package main

import (
	"fmt"

	aero "github.com/ICT-FUNney/utari-aerospike-client"
)

var (
	client aero.IAeroSpikeClinet
	err    error
)

func init() {
	client, err = aero.NewAeroSpikeClient("127.0.0.1", 3000)
	if err != nil {
		panic(err)
	}
}

func main() {
	// indexの作成 初回時のみ実行する すでにindexが貼ってある状態でindexを貼ろうとするとエラーになる
	// createIndex()

	// ダミーデータ作成
	block := aero.Block{
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
	tx := aero.Transaction{
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
	blockHash := aero.GetHash(block)

	// レコードの取得
	blockRecv, err := client.GetBlock(blockHash)
	if err != nil {
		panic(err)
	}
	txs, err := client.GetTransactionByInput("testinput")
	if err != nil {
		panic(err)
	}
	fmt.Printf("transactions:%v\n", txs)

	txs, err = client.GetTransactionByOutput("testoutput")
	if err != nil {
		panic(err)
	}
	fmt.Printf("transactions:%v\n", txs)

	txs, err = client.GetTransactionByInput("hoge")
	if err != nil {
		panic(err)
	}
	fmt.Printf("transactions:%v\n", txs)

	// データの確認
	fmt.Printf("block:%v\n", blockRecv)

	// データの削除
	err = client.DeleteBlock(blockHash)
}

func createIndex() {
	//// オプションの定義
	options := []aero.CreateIndexOptions{
		aero.CreateIndexOptions{
			Namespace: aero.GetAerospikeNamespace(),
			Set:       aero.GetAerospikeTxTable(),
			Bin:       "Input",
			IndexName: "input_index",
			IndexType: aero.IndexTypeString,
		},
		aero.CreateIndexOptions{
			Namespace: aero.GetAerospikeNamespace(),
			Set:       aero.GetAerospikeTxTable(),
			Bin:       "Output",
			IndexName: "output_index",
			IndexType: aero.IndexTypeString,
		},
		aero.CreateIndexOptions{
			Namespace: aero.GetAerospikeNamespace(),
			Set:       aero.GetAerospikeTxTable(),
			Bin:       "Amount",
			IndexName: "amount_index",
			IndexType: aero.IndexTypeNumric,
		},
		aero.CreateIndexOptions{
			Namespace: aero.GetAerospikeNamespace(),
			Set:       aero.GetAerospikeTxTable(),
			Bin:       "Timestamp",
			IndexName: "timestamp_index",
			IndexType: aero.IndexTypeString,
		},
	}

	//// indexの作成
	for _, opt := range options {
		err := client.CreateIndex(opt)
		if err != nil {
			panic(err)
		}
	}
}
