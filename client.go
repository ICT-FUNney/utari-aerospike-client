package utari

import (
	aero "github.com/aerospike/aerospike-client-go"
)

type aeroSpikeClient struct {
	client *aero.Client
}

const (
	// IndexTypeNumric は int型のbinにインデックスを貼るときに使います
	IndexTypeNumric = aero.NUMERIC
	// IndexTypeString は string型のbinにインデックスを貼るときに使います
	IndexTypeString = aero.STRING
)

// IAeroSpikeClinet は AeroSpikeClientの振る舞いを定義
type IAeroSpikeClinet interface {
	PutBlock(Block) error
	PutTransaction(Transaction) error
	PutBalance(string, float64) error
	PutUnsettled(string, float64) error
	GetBlock(string) (Block, error)
	GetTransactionByInput(string) ([]Transaction, error)
	GetTransactionByOutput(string) ([]Transaction, error)
	GetBalanceByAddress(string) (float64, error)
	GetUnsettled(string) (float64, error)
	DeleteBlock(string) error
	DeleteTransaction(string) error
	CreateIndex(CreateIndexOptions) error
	Close()
}

// CreateIndexOptions は インデックスを作成する際のオプション構造体です
type CreateIndexOptions struct {
	Namespace string
	Set       string
	Bin       string
	IndexName string
	IndexType aero.IndexType
}

// NewAeroSpikeClient は 新しいaerospikeクライアントを取得する関数です
func NewAeroSpikeClient(host string, port int) (IAeroSpikeClinet, error) {
	cli, err := aero.NewClient(host, port)
	if err != nil {
		return nil, err
	}
	return aeroSpikeClient{
		client: cli,
	}, nil
}

// PutBlock は 引数のblockをaerospikeに格納するメソッドです
func (a aeroSpikeClient) PutBlock(block Block) error {
	// hash値の取得
	hash := GetHash(block)

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

// PutTransaction は 引数Transactionをaerospikeに格納するメソッドです
func (a aeroSpikeClient) PutTransaction(tx Transaction) error {
	// hash値の取得
	hash := GetHash(tx)

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

// GetBlockは 引数であるhashをキーに Blockを取得するメソッドです
func (a aeroSpikeClient) GetBlock(hash string) (Block, error) {
	key, err := getBlockKey(hash)
	if err != nil {
		return Block{}, err
	}
	// レコードの取得
	record, err := a.client.Get(nil, key)
	if err != nil {
		return Block{}, err
	}

	// binmap to block
	block, err := binMapToBlock(record)
	if err != nil {
		return Block{}, err
	}

	return block, nil
}

// GetTransactionByInputは 引数であるinputとレコードのInputが一致する Transactionを1つ以上取得するメソッドです
func (a aeroSpikeClient) GetTransactionByInput(input string) ([]Transaction, error) {
	// select * from namespace.tableを表す
	stmt := aero.NewStatement(
		GetAerospikeNamespace(),
		GetAerospikeTxTable(),
		"Txid",
		"Output",
		"Input",
		"Amount",
		"Timestamp",
		"Sign",
		"Pubkey",
	)

	// where Input = input を表す
	stmt.SetFilter(aero.NewEqualFilter("Input", input))

	// レコードの取得
	recordSet, err := a.client.Query(nil, stmt)
	if err != nil {
		return nil, err
	}

	var transactions []Transaction
	for result := range recordSet.Results() {
		if result.Err != nil {
			return nil, result.Err
		}

		// binmap to tx
		tx, err := binMapToTransaction(result.Record)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetTransactionByOutputは 引数であるOutputとレコードのOutputが一致する Transactionを1つ以上取得するメソッドです
func (a aeroSpikeClient) GetTransactionByOutput(output string) ([]Transaction, error) {
	// select * from namespace.tableを表す
	stmt := aero.NewStatement(
		GetAerospikeNamespace(),
		GetAerospikeTxTable(),
		"Txid",
		"Output",
		"Input",
		"Amount",
		"Timestamp",
		"Sign",
		"Pubkey",
	)

	// where Input = input を表す
	stmt.SetFilter(aero.NewEqualFilter("Output", output))

	// レコードの取得
	recordSet, err := a.client.Query(nil, stmt)
	if err != nil {
		return nil, err
	}

	var transactions []Transaction
	for result := range recordSet.Results() {
		if result.Err != nil {
			return nil, result.Err
		}

		// binmap to tx
		tx, err := binMapToTransaction(result.Record)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// DeleteBlock は hashをキーに blockを削除するメソッドです
func (a aeroSpikeClient) DeleteBlock(hash string) error {
	key, err := getBlockKey(hash)
	_, err = a.client.Delete(nil, key)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTransaction は hashをキーに transactionを削除するメソッドです
func (a aeroSpikeClient) DeleteTransaction(hash string) error {
	key, err := getTransactionKey(hash)
	_, err = a.client.Delete(nil, key)
	if err != nil {
		return err
	}
	return nil
}

// CreateIndex は CreateIndexOptionに基づいてインデックスを作成するメソッドです
func (a aeroSpikeClient) CreateIndex(option CreateIndexOptions) error {
	task, err := a.client.CreateIndex(nil, option.Namespace, option.Set, option.IndexName, option.Bin, option.IndexType)
	if err != nil {
		return err
	}
	err = <-task.OnComplete()
	if err != nil {
		return err
	}
	return nil
}

func (a aeroSpikeClient) PutBalance(address string, balance float64) error {
	key, err := getBalanceKey(address)
	bal := Balance{Address: address, Balance: balance}
	data := balanceToBinMap(bal)
	err = a.client.Put(nil, key, data)
	if err != nil {
		return err
	}
	return nil
}

func (a aeroSpikeClient) GetBalanceByAddress(address string) (float64, error) {
	key, err := getBalanceKey(address)
	record, err := a.client.Get(nil, key)
	if err != nil {
		return -1.0, err
	}
	if record == nil {
		return 0.0, nil
	}

	bal, err := binMapToBalance(record)
	if err != nil {
		return -1.0, err
	}

	return bal.Balance, nil
}

func (a aeroSpikeClient) PutUnsettled(address string, unsettled float64) error {
	key, err := getUnsettledKey(address)
	bal := Unsettled{Address: address, Unsettled: unsettled}
	data := unsettledToBinMap(bal)
	err = a.client.Put(nil, key, data)
	if err != nil {
		return err
	}
	return nil
}

func (a aeroSpikeClient) GetUnsettled(address string) (float64, error) {
	key, err := getUnsettledKey(address)
	record, err := a.client.Get(nil, key)
	if err != nil {
		return -1.0, err
	}
	if record == nil {
		return 0.0, nil
	}

	uns, err := binMapToUnsettled(record)
	if err != nil {
		return -1.0, err
	}

	return uns.Unsettled, nil
}

func (a aeroSpikeClient) Close() {
	a.client.Close()
}
