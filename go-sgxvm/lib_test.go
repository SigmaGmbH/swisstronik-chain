package librustgo

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/SigmaGmbH/librustgo/internal/api"
	"github.com/SigmaGmbH/librustgo/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestCreate(t *testing.T) {
	db := types.CreateMockedDatabase()
	from := common.HexToAddress("0x690b9a9e9aa1c9db991c7721a92d351db4fac990")

	connector := types.MockedConnector{DB: &db}
	value := common.Hex2Bytes("00")
	data := common.Hex2Bytes("60806040526000805534801561001457600080fd5b50610394806100246000396000f3fe608060405234801561001057600080fd5b50600436106100415760003560e01c80634f2be91f1461004657806361bc221a146100505780636deebae31461006e575b600080fd5b61004e610078565b005b610058610103565b60405161006591906101f1565b60405180910390f35b610076610109565b005b60008081548092919061008a9061023b565b91905055507f64a55044d1f2eddebe1b90e8e2853e8e96931cefadbfa0b2ceb34bee360619416000546040516100c091906101f1565b60405180910390a17f938d2ee5be9cfb0f7270ee2eff90507e94b37625d9d2b3a61c97d30a4560b8296000546040516100f991906101f1565b60405180910390a1565b60005481565b60008054116040518060400160405280600f81526020017f434f554e5445525f544f4f5f4c4f57000000000000000000000000000000000081525090610185576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161017c9190610313565b60405180910390fd5b5060008081548092919061019890610335565b91905055507f938d2ee5be9cfb0f7270ee2eff90507e94b37625d9d2b3a61c97d30a4560b8296000546040516101ce91906101f1565b60405180910390a1565b6000819050919050565b6101eb816101d8565b82525050565b600060208201905061020660008301846101e2565b92915050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000610246826101d8565b91507fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036102785761027761020c565b5b600182019050919050565b600081519050919050565b600082825260208201905092915050565b60005b838110156102bd5780820151818401526020810190506102a2565b60008484015250505050565b6000601f19601f8301169050919050565b60006102e582610283565b6102ef818561028e565b93506102ff81856020860161029f565b610308816102c9565b840191505092915050565b6000602082019050818103600083015261032d81846102da565b905092915050565b6000610340826101d8565b9150600082036103535761035261020c565b5b60018203905091905056fea264697066735822122009b7dbde115b8323afdd451cd1b9c02d5e332011af0eb72b9ef71469fe56ab3564736f6c63430008120033")
	gasLimit := uint64(2000000)
	gasPrice := []byte{0,0}
	txContext := types.GetDefaultTxContext()

	// Calculate contract address
	fromAcct, err := db.GetAccountOrEmpty(from)
	if err != nil {
		t.Fatal(err)
	}
	contractAddress := crypto.CreateAddress(from, fromAcct.Nonce)

	// Deploy contract
	_, err = api.Create(
		connector,
		from.Bytes(),
		data,
		value,
		nil,
		gasLimit,
		gasPrice,
		0,
		txContext,
		true,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Check if contract was deployed correctly
	acct, _ := db.GetAccountOrEmpty(contractAddress)
	if len(acct.Code) == 0 {
		t.Fatal("Contract was not deployed")
	}
}

func TestCoinTransfer(t *testing.T) {
	db := types.CreateMockedDatabase()
	from := common.HexToAddress("0x690b9a9e9aa1c9db991c7721a92d351db4fac990")
	to := common.HexToAddress("0xad60cdbe1d3ceb5f67074303f99ac95af082784d")

	connector := types.MockedConnector{DB: &db}
	value := big.NewInt(100)
	gasLimit := uint64(2000000)
	txContext := types.GetDefaultTxContext()

	// Insert account with balance
	balanceToInsert := big.NewInt(100000)
	err := db.InsertAccount(from, balanceToInsert.Bytes(), 0)
	if err != nil {
		t.Fail()
	}
	fromAcct, err := db.GetAccountOrEmpty(from)
	if err != nil {
		t.Fail()
	}
	toAcct, err := db.GetAccountOrEmpty(to)
	if err != nil {
		t.Fail()
	}
	senderBalanceBefore := types.BytesToBig(fromAcct.Balance).Uint64()
	senderNonceBefore := fromAcct.Nonce
	receiverBalanceBefore := types.BytesToBig(toAcct.Balance).Uint64()
	receiverNonceBefore := toAcct.Nonce

	gasPrice := []byte{0,0}

	// Send a transaction to transfer 100 wei
	_, err = api.Call(
		connector,
		from.Bytes(),
		to.Bytes(),
		nil,
		value.Bytes(),
		nil,
		gasLimit,
		gasPrice,
		0,
		txContext,
		true,
		false,
	)
	if err != nil {
		t.Fail()
	}

	fromAcct, err = db.GetAccountOrEmpty(from)
	if err != nil {
		t.Fail()
	}
	toAcct, err = db.GetAccountOrEmpty(to)
	if err != nil {
		t.Fail()
	}
	senderBalanceAfter := types.BytesToBig(fromAcct.Balance).Uint64()
	senderNonceAfter := fromAcct.Nonce
	receiverBalanceAfter := types.BytesToBig(toAcct.Balance).Uint64()
	receiverNonceAfter := toAcct.Nonce

	// Check if nonce was updated correctly
	if senderNonceBefore+1 != senderNonceAfter {
		t.Fail()
	}
	if receiverNonceBefore != receiverNonceAfter {
		// Receiver's nonce should not be updated
		t.Fail()
	}
	// Check if balances were updated correctly
	if senderBalanceBefore-value.Uint64() != senderBalanceAfter {
		t.Fail()
	}
	if receiverBalanceBefore+value.Uint64() != receiverBalanceAfter {
		t.Fail()
	}
}

func TestSeedExchangeDCAP(t *testing.T) {
	if err := api.InitializeEnclave(true); err != nil {
		t.Fail()
	}

	dcapAddress := "localhost:8998"
	err := api.StartAttestationServer(dcapAddress)
	if err != nil {
		t.Fail()
	}

	// Test DCAP Attestation
	dcapHost := "localhost"
	dcapPort := 8998
	if err := api.RequestEpochKeys(dcapHost, dcapPort, true); err != nil {
		t.Fail()
	} else {
		println("DCAP PASSED")
	}
}

func TestNodeInitialized(t *testing.T) {
	res, err := api.IsNodeInitialized()
	if err != nil {
		t.Fail()
	}

	fmt.Println("node initialized: ", res)
}

func TestDumpQuote(t *testing.T) {
	if err := api.DumpDCAPQuote("quote.dat"); err != nil {
		t.Fail()
	}
}

func TestVerifyQuote(t *testing.T) {
	if err := api.VerifyDCAPQuote("quote.dat"); err != nil {
		t.Fail()
	}
}
