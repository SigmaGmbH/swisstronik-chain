package keeper

import (
	"errors"
	"math/big"

	"github.com/SigmaGmbH/librustgo"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"

	compliancetypes "swisstronik/x/compliance/types"
)

// Connector allows our VM interact with existing Cosmos application.
// It is passed by pointer into SGX to make it accessible for our VM.
type Connector struct {
	// Keeper used to store and obtain state
	EVMKeeper *Keeper
	// Context used to make Keeper calls available
	Context sdk.Context
}

func (q Connector) Query(req []byte) ([]byte, error) {
	// Decode protobuf
	decodedRequest := &librustgo.CosmosRequest{}
	if err := proto.Unmarshal(req, decodedRequest); err != nil {
		return nil, err
	}

	switch request := decodedRequest.Req.(type) {
	// Handle request for account data such as balance and nonce
	case *librustgo.CosmosRequest_GetAccount:
		return q.GetAccount(request)
	// Handles request if such account exists
	case *librustgo.CosmosRequest_ContainsKey:
		return q.ContainsKey(request)
	// Handles contract code request
	case *librustgo.CosmosRequest_AccountCode:
		return q.GetAccountCode(request)
	// Handles storage cell data request
	case *librustgo.CosmosRequest_StorageCell:
		return q.GetStorageCell(request)
	// Handles inserting storage cell
	case *librustgo.CosmosRequest_InsertStorageCell:
		return q.InsertStorageCell(request)
	// Handles updating contract code
	case *librustgo.CosmosRequest_InsertAccountCode:
		return q.InsertAccountCode(request)
	// Handles remove storage cell request
	case *librustgo.CosmosRequest_RemoveStorageCell:
		return q.RemoveStorageCell(request)
	// Handles removing account storage, account record, etc.
	case *librustgo.CosmosRequest_Remove:
		return q.Remove(request)
	// Returns block hash
	case *librustgo.CosmosRequest_BlockHash:
		return q.BlockHash(request)
	case *librustgo.CosmosRequest_AddVerificationDetails:
		return q.AddVerificationDetails(request)
	case *librustgo.CosmosRequest_HasVerification:
		return q.HasVerification(request)
	case *librustgo.CosmosRequest_GetVerificationData:
		return q.GetVerificationData(request)
	case *librustgo.CosmosRequest_InsertAccountNonce:
		return q.InsertAccountNonce(request)
	case *librustgo.CosmosRequest_InsertAccountBalance:
		return q.InsertAccountBalance(request)
	case *librustgo.CosmosRequest_GetAccountCodeHash:
		return q.GetAccountCodeHash(request)
	case *librustgo.CosmosRequest_GetAccountCodeSize:
		return q.GetAccountCodeSize(request)
	case *librustgo.CosmosRequest_IssuanceTreeRoot:
		return q.GetIssuanceTreeRoot()
	case *librustgo.CosmosRequest_RevocationTreeRoot:
		return q.GetRevocationTreeRoot()
	case *librustgo.CosmosRequest_AddVerificationDetailsV2:
		return q.AddVerificationDetailsV2(request)
	case *librustgo.CosmosRequest_RevokeVerification:
		return q.RevokeVerification(request)
	case *librustgo.CosmosRequest_ConvertCredential:
		return q.ConvertCredential(request)
	}

	return nil, errors.New("wrong query received")
}

// GetAccount handles incoming protobuf-encoded request for account data such as balance and nonce.
// Returns data in protobuf-encoded format
func (q Connector) GetAccount(req *librustgo.CosmosRequest_GetAccount) ([]byte, error) {
	//println("Connector::Query GetAccount invoked")
	ethAddress := common.BytesToAddress(req.GetAccount.Address)
	account := q.EVMKeeper.GetAccountOrEmpty(q.Context, ethAddress)

	return proto.Marshal(&librustgo.QueryGetAccountResponse{
		Balance: account.Balance.Bytes(),
		Nonce:   account.Nonce,
	})
}

// ContainsKey handles incoming protobuf-encoded request to check whether specified address exists
func (q Connector) ContainsKey(req *librustgo.CosmosRequest_ContainsKey) ([]byte, error) {
	//println("Connector::Query ContainsKey invoked")
	ethAddress := common.BytesToAddress(req.ContainsKey.Key)
	account := q.EVMKeeper.GetAccountWithoutBalance(q.Context, ethAddress)
	return proto.Marshal(&librustgo.QueryContainsKeyResponse{Contains: account != nil})
}

// InsertAccountCode handles incoming protobuf-encoded request for adding or modifying existing account code
// It will insert account code only if account exists, otherwise it returns an error
func (q Connector) InsertAccountCode(req *librustgo.CosmosRequest_InsertAccountCode) ([]byte, error) {
	//println("Connector::Query InsertAccountCode invoked")
	ethAddress := common.BytesToAddress(req.InsertAccountCode.Address)
	if err := q.EVMKeeper.SetAccountCode(q.Context, ethAddress, req.InsertAccountCode.Code); err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryInsertAccountCodeResponse{})
}

// RemoveStorageCell handles incoming protobuf-encoded request for removing contract storage cell for given key (index)
func (q Connector) RemoveStorageCell(req *librustgo.CosmosRequest_RemoveStorageCell) ([]byte, error) {
	//println("Connector::Query RemoveStorageCell invoked")
	address := common.BytesToAddress(req.RemoveStorageCell.Address)
	index := common.BytesToHash(req.RemoveStorageCell.Index)

	q.EVMKeeper.SetState(q.Context, address, index, common.Hash{}.Bytes())

	return proto.Marshal(&librustgo.QueryRemoveStorageCellResponse{})
}

// Remove handles incoming protobuf-encoded request for removing smart contract (selfdestruct)
func (q Connector) Remove(req *librustgo.CosmosRequest_Remove) ([]byte, error) {
	//println("Connector::Query Remove invoked")
	ethAddress := common.BytesToAddress(req.Remove.Address)
	if err := q.EVMKeeper.DeleteAccount(q.Context, ethAddress); err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryRemoveResponse{})
}

// BlockHash handles incoming protobuf-encoded request for getting block hash
func (q Connector) BlockHash(req *librustgo.CosmosRequest_BlockHash) ([]byte, error) {
	//println("Connector::Query BlockHash invoked")

	blockNumber := &big.Int{}
	blockNumber.SetBytes(req.BlockHash.Number)
	blockHash := q.EVMKeeper.GetHashFn(q.Context)(blockNumber.Uint64())

	return proto.Marshal(&librustgo.QueryBlockHashResponse{Hash: blockHash.Bytes()})
}

// InsertStorageCell handles incoming protobuf-encoded request for updating state of storage cell
func (q Connector) InsertStorageCell(req *librustgo.CosmosRequest_InsertStorageCell) ([]byte, error) {
	ethAddress := common.BytesToAddress(req.InsertStorageCell.Address)
	index := common.BytesToHash(req.InsertStorageCell.Index)

	q.EVMKeeper.SetState(q.Context, ethAddress, index, req.InsertStorageCell.Value)
	return proto.Marshal(&librustgo.QueryInsertStorageCellResponse{})
}

// GetStorageCell handles incoming protobuf-encoded request of storage cell value
func (q Connector) GetStorageCell(req *librustgo.CosmosRequest_StorageCell) ([]byte, error) {
	//println("Connector::Query Request value of storage cell")
	ethAddress := common.BytesToAddress(req.StorageCell.Address)
	index := common.BytesToHash(req.StorageCell.Index)
	value := q.EVMKeeper.GetState(q.Context, ethAddress, index)

	return proto.Marshal(&librustgo.QueryGetAccountStorageCellResponse{Value: value})
}

// GetAccountCode handles incoming protobuf-encoded request and returns bytecode associated
// with given account. If account does not exist, it returns empty response
func (q Connector) GetAccountCode(req *librustgo.CosmosRequest_AccountCode) ([]byte, error) {
	//println("Connector::Query Request account code")
	ethAddress := common.BytesToAddress(req.AccountCode.Address)
	account := q.EVMKeeper.GetAccountWithoutBalance(q.Context, ethAddress)
	if account == nil {
		return proto.Marshal(&librustgo.QueryGetAccountCodeResponse{
			Code: nil,
		})
	}

	code := q.EVMKeeper.GetCode(q.Context, common.BytesToHash(account.CodeHash))
	return proto.Marshal(&librustgo.QueryGetAccountCodeResponse{
		Code: code,
	})
}

func (q Connector) InsertAccountNonce(req *librustgo.CosmosRequest_InsertAccountNonce) ([]byte, error) {
	ethAddress := common.BytesToAddress(req.InsertAccountNonce.Address)
	nonce := req.InsertAccountNonce.Nonce

	if err := q.EVMKeeper.SetNonce(q.Context, ethAddress, nonce); err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryInsertAccountNonceResponse{})
}

func (q Connector) InsertAccountBalance(req *librustgo.CosmosRequest_InsertAccountBalance) ([]byte, error) {
	ethAddress := common.BytesToAddress(req.InsertAccountBalance.Address)
	balance := &big.Int{}
	balance.SetBytes(req.InsertAccountBalance.Balance)

	if err := q.EVMKeeper.SetBalance(q.Context, ethAddress, balance); err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryInsertAccountBalanceResponse{})
}

func (q Connector) GetAccountCodeHash(req *librustgo.CosmosRequest_GetAccountCodeHash) ([]byte, error) {
	ethAddress := common.BytesToAddress(req.CodeHash.Address)
	account := q.EVMKeeper.GetAccountOrEmpty(q.Context, ethAddress)

	return proto.Marshal(&librustgo.QueryAccountCodeHashResponse{Hash: account.CodeHash})
}

func (q Connector) GetAccountCodeSize(req *librustgo.CosmosRequest_GetAccountCodeSize) ([]byte, error) {
	ethAddress := common.BytesToAddress(req.CodeSize.Address)
	account := q.EVMKeeper.GetAccountWithoutBalance(q.Context, ethAddress)
	if account == nil {
		return proto.Marshal(&librustgo.QueryAccountCodeSizeResponse{
			Size: 0,
		})
	}

	code := q.EVMKeeper.GetCode(q.Context, common.BytesToHash(account.CodeHash))
	return proto.Marshal(&librustgo.QueryAccountCodeSizeResponse{Size: uint32(len(code))})
}

// AddVerificationDetails writes provided verification details to x/compliance module
func (q Connector) AddVerificationDetails(req *librustgo.CosmosRequest_AddVerificationDetails) ([]byte, error) {
	userAddress := sdk.AccAddress(req.AddVerificationDetails.UserAddress)
	issuerAddress := sdk.AccAddress(req.AddVerificationDetails.IssuerAddress).String()
	verificationType := compliancetypes.VerificationType(req.AddVerificationDetails.VerificationType)

	// Addresses in keeper are Cosmos Addresses
	verificationDetails := &compliancetypes.VerificationDetails{
		IssuerAddress:        issuerAddress,
		OriginChain:          req.AddVerificationDetails.OriginChain,
		IssuanceTimestamp:    req.AddVerificationDetails.IssuanceTimestamp,
		ExpirationTimestamp:  req.AddVerificationDetails.ExpirationTimestamp,
		OriginalData:         req.AddVerificationDetails.ProofData,
		Schema:               string(req.AddVerificationDetails.Schema),
		IssuerVerificationId: string(req.AddVerificationDetails.IssuerVerificationId),
		Version:              req.AddVerificationDetails.Version,
	}

	verificationID, err := q.EVMKeeper.ComplianceKeeper.AddVerificationDetails(q.Context, userAddress, verificationType, verificationDetails)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryAddVerificationDetailsResponse{
		VerificationId: verificationID,
	})
}

// AddVerificationDetailsV2 writes provided verification details to x/compliance module
func (q Connector) AddVerificationDetailsV2(req *librustgo.CosmosRequest_AddVerificationDetailsV2) ([]byte, error) {
	userAddress := sdk.AccAddress(req.AddVerificationDetailsV2.UserAddress)
	issuerAddress := sdk.AccAddress(req.AddVerificationDetailsV2.IssuerAddress).String()
	verificationType := compliancetypes.VerificationType(req.AddVerificationDetailsV2.VerificationType)

	// Addresses in keeper are Cosmos Addresses
	verificationDetails := &compliancetypes.VerificationDetails{
		IssuerAddress:        issuerAddress,
		OriginChain:          req.AddVerificationDetailsV2.OriginChain,
		IssuanceTimestamp:    req.AddVerificationDetailsV2.IssuanceTimestamp,
		ExpirationTimestamp:  req.AddVerificationDetailsV2.ExpirationTimestamp,
		OriginalData:         req.AddVerificationDetailsV2.ProofData,
		Schema:               req.AddVerificationDetailsV2.Schema,
		IssuerVerificationId: req.AddVerificationDetailsV2.IssuerVerificationId,
		Version:              req.AddVerificationDetailsV2.Version,
	}

	verificationID, err := q.EVMKeeper.ComplianceKeeper.AddVerificationDetailsV2(q.Context, userAddress, verificationType, verificationDetails, req.AddVerificationDetailsV2.UserPublicKey)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryAddVerificationDetailsResponse{
		VerificationId: verificationID,
	})
}

// HasVerification returns if user has verification of provided type from x/compliance module
func (q Connector) HasVerification(req *librustgo.CosmosRequest_HasVerification) ([]byte, error) {
	userAddress := sdk.AccAddress(req.HasVerification.UserAddress)
	verificationType := compliancetypes.VerificationType(req.HasVerification.VerificationType)
	expirationTimestamp := req.HasVerification.ExpirationTimestamp

	var allowedIssuers []sdk.AccAddress
	for _, issuer := range req.HasVerification.AllowedIssuers {
		allowedIssuers = append(allowedIssuers, sdk.AccAddress(issuer))
	}

	hasVerification, err := q.EVMKeeper.ComplianceKeeper.HasVerificationOfType(q.Context, userAddress, verificationType, expirationTimestamp, allowedIssuers)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryHasVerificationResponse{
		HasVerification: hasVerification,
	})
}

func (q Connector) GetVerificationData(req *librustgo.CosmosRequest_GetVerificationData) ([]byte, error) {
	userAddress := sdk.AccAddress(req.GetVerificationData.UserAddress)
	issuerAddress := sdk.AccAddress(req.GetVerificationData.IssuerAddress)

	verifications, verificationsDetails, err := q.EVMKeeper.ComplianceKeeper.GetVerificationDetailsByIssuer(q.Context, userAddress, issuerAddress)
	if err != nil {
		return nil, err
	}
	if len(verifications) != len(verificationsDetails) {
		return nil, errors.New("invalid verification details")
	}

	var resData []*librustgo.VerificationDetails
	for i, v := range verifications {
		details := verificationsDetails[i]

		isVerificationRevoked, err := q.EVMKeeper.ComplianceKeeper.IsVerificationRevoked(q.Context, verifications[i].VerificationId)
		if err != nil {
			return nil, err
		}

		if !isVerificationRevoked {
			issuerAccount, err := sdk.AccAddressFromBech32(v.IssuerAddress)
			if err != nil {
				return nil, err
			}

			// Addresses from Query requests are Ethereum Addresses
			resData = append(resData, &librustgo.VerificationDetails{
				VerificationType:     uint32(v.Type),
				VerificationID:       v.VerificationId,
				IssuerAddress:        common.Address(issuerAccount.Bytes()).Bytes(),
				OriginChain:          details.OriginChain,
				IssuanceTimestamp:    details.IssuanceTimestamp,
				ExpirationTimestamp:  details.ExpirationTimestamp,
				OriginalData:         details.OriginalData,
				Schema:               details.Schema,
				IssuerVerificationId: details.IssuerVerificationId,
				Version:              details.Version,
			})
		}
	}
	return proto.Marshal(&librustgo.QueryGetVerificationDataResponse{
		Data: resData,
	})
}

func (q Connector) GetIssuanceTreeRoot() ([]byte, error) {
	root, err := q.EVMKeeper.ComplianceKeeper.GetIssuanceTreeRoot(q.Context)
	if err != nil {
		return nil, err
	}

	if root == nil {
		return nil, errors.New("issuance root not found")
	}

	return proto.Marshal(&librustgo.QueryIssuanceTreeRootResponse{
		Root: root.Bytes(),
	})
}

func (q Connector) GetRevocationTreeRoot() ([]byte, error) {
	root, err := q.EVMKeeper.ComplianceKeeper.GetRevocationTreeRoot(q.Context)
	if err != nil {
		return nil, err
	}

	if root == nil {
		return nil, errors.New("revocation root not found")
	}

	return proto.Marshal(&librustgo.QueryRevocationTreeRootResponse{
		Root: root.Bytes(),
	})
}

func (q Connector) RevokeVerification(req *librustgo.CosmosRequest_RevokeVerification) ([]byte, error) {
	issuerAddress := sdk.AccAddress(req.RevokeVerification.Issuer)

	err := q.EVMKeeper.ComplianceKeeper.RevokeVerification(q.Context, req.RevokeVerification.VerificationId, issuerAddress)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryRevokeVerificationResponse{})
}

func (q Connector) ConvertCredential(req *librustgo.CosmosRequest_ConvertCredential) ([]byte, error) {
	caller := sdk.AccAddress(req.ConvertCredential.Caller)

	if req.ConvertCredential.VerificationId == nil {
		return nil, errors.New("invalid verification id")
	}

	if req.ConvertCredential.HolderPublicKey == nil {
		return nil, errors.New("invalid holder public key")
	}

	err := q.EVMKeeper.ComplianceKeeper.ConvertCredential(
		q.Context,
		req.ConvertCredential.VerificationId,
		req.ConvertCredential.HolderPublicKey,
		caller,
	)
	if err != nil {
		return nil, err
	}

	return proto.Marshal(&librustgo.QueryConvertCredentialResponse{})
}
