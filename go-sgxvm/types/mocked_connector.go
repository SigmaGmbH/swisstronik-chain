package types

import (
	"errors"
	ffi "swisstronik/go-sgxvm/go_protobuf_gen"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"
)

type MockedConnector struct {
	DB *MockedDB
}

var _ Connector = MockedConnector{}

func (c MockedConnector) Query(request []byte) ([]byte, error) {
	// Decode protobuf
	decodedRequest := &ffi.CosmosRequest{}
	if err := proto.Unmarshal(request, decodedRequest); err != nil {
		return nil, err
	}

	switch request := decodedRequest.Req.(type) {
	case *ffi.CosmosRequest_BlockHash:
		println("[Go:Query] Block hash")
		blockHash := make([]byte, 32)
		return proto.Marshal(&ffi.QueryBlockHashResponse{Hash: blockHash})
	case *ffi.CosmosRequest_GetAccount:
		ethAddress := common.BytesToAddress(request.GetAccount.Address)
		println("[Go:Query] get account data for address: ", ethAddress.String())
		acct, err := c.DB.GetAccountOrEmpty(ethAddress)
		if err != nil {
			return nil, err
		}

		return proto.Marshal(&ffi.QueryGetAccountResponse{
			Balance: acct.Balance,
			Nonce:   acct.Nonce,
		})
	case *ffi.CosmosRequest_InsertAccount:
		data := request.InsertAccount
		ethAddress := common.BytesToAddress(request.InsertAccount.Address)
		println("[Go:Query] Insert/Update account: ", ethAddress.String())
		if err := c.DB.InsertAccount(ethAddress, data.Balance, data.Nonce); err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryInsertAccountResponse{})
	case *ffi.CosmosRequest_ContainsKey:
		ethAddress := common.BytesToAddress(request.ContainsKey.Key)
		println("[Go:Query] Contains key for: ", ethAddress.String())
		contains, err := c.DB.Contains(ethAddress)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryContainsKeyResponse{Contains: contains})
	case *ffi.CosmosRequest_AccountCode:
		ethAddress := common.BytesToAddress(request.AccountCode.Address)
		acct, err := c.DB.GetAccountOrEmpty(ethAddress)
		if err != nil {
			return nil, err
		}
		println("[Go:Query] Account code: ", ethAddress.String(), "code len: ", len(acct.Code))
		return proto.Marshal(&ffi.QueryGetAccountCodeResponse{Code: acct.Code})
	case *ffi.CosmosRequest_StorageCell:
		ethAddress := common.BytesToAddress(request.StorageCell.Address)
		println("[Go:Query] Get storage cell: ", ethAddress.String())
		value, err := c.DB.GetStorageCell(ethAddress, request.StorageCell.Index)
		if err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryGetAccountStorageCellResponse{Value: value})
	case *ffi.CosmosRequest_InsertAccountCode:
		ethAddress := common.BytesToAddress(request.InsertAccountCode.Address)
		println("[Go:Query] Insert account code: ", ethAddress.String(), "len: ", len(request.InsertAccountCode.Code))
		if err := c.DB.InsertContractCode(ethAddress, request.InsertAccountCode.Code); err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryInsertAccountCodeResponse{})
	case *ffi.CosmosRequest_InsertStorageCell:
		data := request.InsertStorageCell
		ethAddress := common.BytesToAddress(request.InsertStorageCell.Address)
		println("[Go:Query] Insert storage cell: ", ethAddress.String())
		if err := c.DB.InsertStorageCell(ethAddress, data.Index, data.Value); err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryInsertStorageCellResponse{})
	case *ffi.CosmosRequest_Remove:
		ethAddress := common.BytesToAddress(request.Remove.Address)
		println("[Go:Query] Remove account: ", ethAddress.String())
		if err := c.DB.Delete(ethAddress); err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryRemoveResponse{})
	case *ffi.CosmosRequest_RemoveStorageCell:
		ethAddress := common.BytesToAddress(request.RemoveStorageCell.Address)
		println("[Go:Query] Remove storage cell: ", ethAddress.String())
		if err := c.DB.RemoveStorageCell(ethAddress, request.RemoveStorageCell.Index); err != nil {
			return nil, err
		}
		return proto.Marshal(&ffi.QueryRemoveStorageCellResponse{})
	}

	return nil, errors.New("wrong query")
}

func GetDefaultTxContext() *ffi.TransactionContext {
	return &ffi.TransactionContext{
		BlockCoinbase:      common.Address{}.Bytes(),
		BlockNumber:        0,
		BlockBaseFeePerGas: make([]byte, 32),
		Timestamp:          0,
		BlockGasLimit:      100000000000,
		ChainId:            1,
		GasPrice:           make([]byte, 32),
	}
}
