package librustgo

import (
	"swisstronik/go-sgxvm/internal/api"
	"swisstronik/go-sgxvm/types"

	ffi "swisstronik/go-sgxvm/go_protobuf_gen"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Logs returned by EVM
type Log = ffi.Log
type Topic = ffi.Topic

// TransactionContext contains information about block timestamp, coinbase address, block gas limit, etc.
type TransactionContext = ffi.TransactionContext

// TransactionData contains data which is necessary to handle the transaction
type TransactionData = ffi.TransactionData

// Export protobuf messages for FFI
type QueryGetAccount = ffi.QueryGetAccount
type QueryGetAccountResponse = ffi.QueryGetAccountResponse
type CosmosRequest = ffi.CosmosRequest
type QueryInsertAccount = ffi.QueryInsertAccount
type QueryInsertAccountResponse = ffi.QueryInsertAccountResponse
type QueryContainsKey = ffi.QueryContainsKey
type QueryContainsKeyResponse = ffi.QueryContainsKeyResponse
type QueryGetAccountStorageCell = ffi.QueryGetAccountStorageCell
type QueryGetAccountStorageCellResponse = ffi.QueryGetAccountStorageCellResponse
type QueryGetAccountCode = ffi.QueryGetAccountCode
type QueryGetAccountCodeResponse = ffi.QueryGetAccountCodeResponse
type QueryInsertAccountCode = ffi.QueryInsertAccountCode
type QueryInsertAccountCodeResponse = ffi.QueryInsertAccountCodeResponse
type QueryInsertStorageCell = ffi.QueryInsertStorageCell
type QueryInsertStorageCellResponse = ffi.QueryInsertStorageCellResponse
type QueryRemove = ffi.QueryRemove
type QueryRemoveResponse = ffi.QueryRemoveResponse
type QueryRemoveStorageCell = ffi.QueryRemoveStorageCell
type QueryRemoveStorageCellResponse = ffi.QueryRemoveStorageCellResponse
type QueryBlockHash = ffi.QueryBlockHash
type QueryBlockHashResponse = ffi.QueryBlockHashResponse
type VerificationMethod = ffi.VerificationMethod
type QueryVerificationMethods = ffi.QueryVerificationMethods
type QueryVerificationMethodsResponse = ffi.QueryVerificationMethodsResponse

// Storage requests
type CosmosRequest_GetAccount = ffi.CosmosRequest_GetAccount
type CosmosRequest_InsertAccount = ffi.CosmosRequest_InsertAccount
type CosmosRequest_ContainsKey = ffi.CosmosRequest_ContainsKey
type CosmosRequest_AccountCode = ffi.CosmosRequest_AccountCode
type CosmosRequest_StorageCell = ffi.CosmosRequest_StorageCell
type CosmosRequest_InsertAccountCode = ffi.CosmosRequest_InsertAccountCode
type CosmosRequest_InsertStorageCell = ffi.CosmosRequest_InsertStorageCell
type CosmosRequest_Remove = ffi.CosmosRequest_Remove
type CosmosRequest_RemoveStorageCell = ffi.CosmosRequest_RemoveStorageCell
type CosmosRequest_VerificationMethods = ffi.CosmosRequest_VerificationMethods

// Backend requests
type CosmosRequest_BlockHash = ffi.CosmosRequest_BlockHash

type HandleTransactionResponse = ffi.HandleTransactionResponse
type NodePublicKeyRequest = ffi.NodePublicKeyRequest
type NodePublicKeyResponse = ffi.NodePublicKeyResponse

// IsNodeInitialized checks if node was properly initialized and master key was sealed
func IsNodeInitialized() (bool, error) {
	return api.IsNodeInitialized()
}

// Call handles incoming transaction data to transfer value or call some contract
func Call(
	querier types.Connector,
	from, to, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit, nonce uint64,
	txContext *TransactionContext,
	commit bool,
) (*ffi.HandleTransactionResponse, error) {
	executionResult, err := api.Call(querier, from, to, data, value, accessList, gasLimit, nonce, txContext, commit)
	if err != nil {
		return &ffi.HandleTransactionResponse{}, err
	}

	return executionResult, nil
}

// Create handles incoming transaction data and creates a new smart contract
func Create(
	querier types.Connector,
	from, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit, nonce uint64,
	txContext *TransactionContext,
	commit bool,
) (*ffi.HandleTransactionResponse, error) {
	executionResult, err := api.Create(querier, from, data, value, accessList, gasLimit, nonce, txContext, commit)
	if err != nil {
		return &ffi.HandleTransactionResponse{}, err
	}

	return executionResult, nil
}

func InitializeMasterKey(shouldReset bool) error {
	return api.InitializeMasterKey(shouldReset)
}

func StartSeedServer(addr string) error {
	return api.StartSeedServer(addr)
}

// RequestSeed handles requesting seed and passing Remote Attestation.
// Returns error if Remote Attestation was not passed or provided seed server address is not accessible
func RequestSeed(host string, port int) error {
	return api.RequestSeed(host, port)
}

// GetNodePublicKey handles request for node public key
func GetNodePublicKey() (*ffi.NodePublicKeyResponse, error) {
	result, err := api.GetNodePublicKey()
	if err != nil {
		return &ffi.NodePublicKeyResponse{}, err
	}
	return result, nil
}

// Libsgx_wrapperVersion returns the version of the loaded library
// at runtime. This can be used for debugging to verify the loaded version
// matches the expected version.
func Libsgx_wrapperVersion() (string, error) {
	return api.Libsgx_wrapperVersion()
}
