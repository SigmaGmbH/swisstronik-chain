package librustgo

import (
	"github.com/SigmaGmbH/librustgo/internal/api"
	"github.com/SigmaGmbH/librustgo/types"
	"math/big"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Logs returned by EVM
type Log = types.Log
type Topic = types.Topic

// TransactionContext contains information about block timestamp, coinbase address, block gas limit, etc.
type TransactionContext = types.TransactionContext

// TransactionData contains data which is necessary to handle the transaction
type TransactionData = types.TransactionData

// Export protobuf messages for FFI
type QueryGetAccount = types.QueryGetAccount
type QueryGetAccountResponse = types.QueryGetAccountResponse
type CosmosRequest = types.CosmosRequest
type QueryContainsKey = types.QueryContainsKey
type QueryContainsKeyResponse = types.QueryContainsKeyResponse
type QueryGetAccountStorageCell = types.QueryGetAccountStorageCell
type QueryGetAccountStorageCellResponse = types.QueryGetAccountStorageCellResponse
type QueryGetAccountCode = types.QueryGetAccountCode
type QueryGetAccountCodeResponse = types.QueryGetAccountCodeResponse
type QueryInsertAccountCode = types.QueryInsertAccountCode
type QueryInsertAccountCodeResponse = types.QueryInsertAccountCodeResponse
type QueryInsertStorageCell = types.QueryInsertStorageCell
type QueryInsertStorageCellResponse = types.QueryInsertStorageCellResponse
type QueryRemove = types.QueryRemove
type QueryRemoveResponse = types.QueryRemoveResponse
type QueryRemoveStorageCell = types.QueryRemoveStorageCell
type QueryRemoveStorageCellResponse = types.QueryRemoveStorageCellResponse
type QueryBlockHash = types.QueryBlockHash
type QueryBlockHashResponse = types.QueryBlockHashResponse
type QueryAddVerificationDetails = types.QueryAddVerificationDetails
type QueryAddVerificationDetailsResponse = types.QueryAddVerificationDetailsResponse
type QueryHasVerification = types.QueryHasVerification
type QueryHasVerificationResponse = types.QueryHasVerificationResponse
type QueryGetVerificationData = types.QueryGetVerificationData
type VerificationDetails = types.VerificationDetails
type QueryGetVerificationDataResponse = types.QueryGetVerificationDataResponse
type QueryAccountCodeSize = types.QueryGetAccountCodeSize
type QueryAccountCodeSizeResponse = types.QueryGetAccountCodeSizeResponse
type QueryAccountCodeHash = types.QueryGetAccountCodeHash
type QueryAccountCodeHashResponse = types.QueryGetAccountCodeHashResponse
type QueryInsertAccountNonce = types.QueryInsertAccountNonce
type QueryInsertAccountNonceResponse = types.QueryInsertAccountNonceResponse
type QueryInsertAccountBalance = types.QueryInsertAccountBalance
type QueryInsertAccountBalanceResponse = types.QueryInsertAccountBalanceResponse
type QueryIssuanceTreeRoot = types.QueryIssuanceTreeRoot
type QueryIssuanceTreeRootResponse = types.QueryIssuanceTreeRootResponse
type QueryRevocationTreeRoot = types.QueryRevocationTreeRoot
type QueryRevocationTreeRootResponse = types.QueryRevocationTreeRootResponse
type QueryAddVerificationDetailsV2 = types.QueryAddVerificationDetailsV2
type QueryAddVerificationDetailsV2Response = types.QueryAddVerificationDetailsV2Response
type QueryRevokeVerification = types.QueryRevokeVerification
type QueryRevokeVerificationResponse = types.QueryRevokeVerificationResponse
type QueryConvertCredential = types.QueryConvertCredential
type QueryConvertCredentialResponse = types.QueryConvertCredentialResponse

// Storage requests
type CosmosRequest_GetAccount = types.CosmosRequest_GetAccount
type CosmosRequest_ContainsKey = types.CosmosRequest_ContainsKey
type CosmosRequest_AccountCode = types.CosmosRequest_AccountCode
type CosmosRequest_StorageCell = types.CosmosRequest_StorageCell
type CosmosRequest_InsertAccountCode = types.CosmosRequest_InsertAccountCode
type CosmosRequest_InsertStorageCell = types.CosmosRequest_InsertStorageCell
type CosmosRequest_Remove = types.CosmosRequest_Remove
type CosmosRequest_RemoveStorageCell = types.CosmosRequest_RemoveStorageCell
type CosmosRequest_AddVerificationDetails = types.CosmosRequest_AddVerificationDetails
type CosmosRequest_HasVerification = types.CosmosRequest_HasVerification
type CosmosRequest_GetVerificationData = types.CosmosRequest_GetVerificationData
type CosmosRequest_GetAccountCodeSize = types.CosmosRequest_CodeSize
type CosmosRequest_GetAccountCodeHash = types.CosmosRequest_CodeHash
type CosmosRequest_InsertAccountBalance = types.CosmosRequest_InsertAccountBalance
type CosmosRequest_InsertAccountNonce = types.CosmosRequest_InsertAccountNonce
type CosmosRequest_IssuanceTreeRoot = types.CosmosRequest_IssuanceTreeRoot
type CosmosRequest_RevocationTreeRoot = types.CosmosRequest_RevocationTreeRoot
type CosmosRequest_AddVerificationDetailsV2 = types.CosmosRequest_AddVerificationDetailsV2
type CosmosRequest_RevokeVerification = types.CosmosRequest_RevokeVerification
type CosmosRequest_ConvertCredential = types.CosmosRequest_ConvertCredential

// Backend requests
type CosmosRequest_BlockHash = types.CosmosRequest_BlockHash

type HandleTransactionResponse = types.HandleTransactionResponse
type NodePublicKeyRequest = types.NodePublicKeyRequest
type NodePublicKeyResponse = types.NodePublicKeyResponse

// CheckNodeStatus checks if SGX requirements are met
func CheckNodeStatus() error {
	return api.CheckNodeStatus()
}

// IsNodeInitialized checks if node was properly initialized and key manager state was sealed
func IsNodeInitialized() (bool, error) {
	return api.IsNodeInitialized()
}

// Call handles incoming transaction data to transfer value or call some contract
func Call(
	querier types.Connector,
	from, to, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit uint64,
	gasPrice *big.Int,
	nonce uint64,
	txContext *TransactionContext,
	commit bool,
	isUnencrypted bool,
	transactionSignature []byte,
	maxFeePerGas *big.Int,
	maxPriorityFeePerGas *big.Int,
	txType uint8,
) (*types.HandleTransactionResponse, error) {
	executionResult, err := api.Call(querier, from, to, data, value, accessList, gasLimit, gasPrice, nonce, txContext, commit, isUnencrypted, transactionSignature, maxFeePerGas, maxPriorityFeePerGas, txType)
	if err != nil {
		return &types.HandleTransactionResponse{}, err
	}

	return executionResult, nil
}

func EstimateGas(
	querier types.Connector,
	from, to, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit uint64,
	gasPrice *big.Int,
	nonce uint64,
	txContext *TransactionContext,
	isUnencrypted bool,
	maxFeePerGas *big.Int,
	maxPriorityFeePerGas *big.Int,
	txType uint8,
) (*types.HandleTransactionResponse, error) {
	executionResult, err := api.EstimateGas(querier, from, to, data, value, accessList, gasLimit, gasPrice, nonce, txContext, isUnencrypted, maxFeePerGas, maxPriorityFeePerGas, txType)
	if err != nil {
		return &types.HandleTransactionResponse{}, err
	}

	return executionResult, nil
}

// Create handles incoming transaction data and creates a new smart contract
func Create(
	querier types.Connector,
	from, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit uint64,
	gasPrice *big.Int,
	nonce uint64,
	txContext *TransactionContext,
	commit bool,
	transactionSignature []byte,
	maxFeePerGas *big.Int,
	maxPriorityFeePerGas *big.Int,
	txType uint8,
) (*types.HandleTransactionResponse, error) {
	executionResult, err := api.Create(querier, from, data, value, accessList, gasLimit, gasPrice, nonce, txContext, commit, transactionSignature, maxFeePerGas, maxPriorityFeePerGas, txType)
	if err != nil {
		return &types.HandleTransactionResponse{}, err
	}

	return executionResult, nil
}

func InitializeEnclave(shouldReset bool) error {
	return api.InitializeEnclave(shouldReset)
}

// StartAttestationServer handles incoming request for starting attestation server
// to share epoch keys with new nodes which passed Remote Attestation.
func StartAttestationServer(dcapAddress string) error {
	return api.StartAttestationServer(dcapAddress)
}

// RequestEpochKeys handles requesting seed and passing Remote Attestation.
// Returns error if Remote Attestation was not passed or provided seed server address is not accessible
func RequestEpochKeys(host string, port int) error {
	return api.RequestEpochKeys(host, port)
}

// GetNodePublicKey handles request for node public key
func GetNodePublicKey(blockNumber uint64) (*types.NodePublicKeyResponse, error) {
	result, err := api.GetNodePublicKey(blockNumber)
	if err != nil {
		return &types.NodePublicKeyResponse{}, err
	}
	return result, nil
}

// Libsgx_wrapperVersion returns the version of the loaded library
// at runtime. This can be used for debugging to verify the loaded version
// matches the expected version.
func Libsgx_wrapperVersion() (string, error) {
	return api.Libsgx_wrapperVersion()
}

func AddEpoch(startingBlock uint64) error {
	return api.AddEpoch(startingBlock)
}

func RemoveLatestEpoch() error {
	return api.RemoveLatestEpoch()
}

func ListEpochs() ([]*types.EpochData, error) {
	return api.ListEpochs()
}
