package keeper_test

import (
	"math/big"
	"math/rand"

	"github.com/SigmaGmbH/librustgo"
	"github.com/SigmaGmbH/librustgo/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"

	"swisstronik/tests"
	compliancetypes "swisstronik/x/compliance/types"
	evmkeeper "swisstronik/x/evm/keeper"
)

func insertAccount(
	connector *evmkeeper.Connector,
	address common.Address,
	balance, nonce *big.Int,
) error {
	// Encode request
	request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
		Req: &librustgo.CosmosRequest_InsertAccount{
			InsertAccount: &librustgo.QueryInsertAccount{
				Address: address.Bytes(),
				Balance: balance.Bytes(),
				Nonce:   nonce.Uint64(),
			},
		},
	})

	if encodeErr != nil {
		return encodeErr
	}

	responseBytes, queryErr := connector.Query(request)
	if queryErr != nil {
		return queryErr
	}

	response := &librustgo.QueryInsertAccountResponse{}
	decodingError := proto.Unmarshal(responseBytes, response)
	if decodingError != nil {
		return decodingError
	}

	return nil
}

func (suite *KeeperTestSuite) TestSGXVMConnector() {
	var (
		connector evmkeeper.Connector
	)

	connector = evmkeeper.Connector{
		Context:   suite.ctx,
		EVMKeeper: suite.app.EvmKeeper,
	}

	testCases := []struct {
		name   string
		action func()
	}{
		{
			"Should be able to insert account",
			func() {
				var err error

				addressToSet := common.BigToAddress(big.NewInt(rand.Int63n(100000)))
				balanceToSet := big.NewInt(10000)
				nonceToSet := big.NewInt(1)

				err = insertAccount(&connector, addressToSet, balanceToSet, nonceToSet)
				suite.Require().NoError(err)

				// Check if account was inserted correctly
				account := connector.EVMKeeper.GetAccountOrEmpty(connector.Context, addressToSet)
				suite.Require().Equal(balanceToSet, account.Balance)
				suite.Require().Equal(nonceToSet.Uint64(), account.Nonce)
			},
		},
		{
			"Should be able to check if account exists",
			func() {
				addressToSet := common.BigToAddress(big.NewInt(rand.Int63n(100000)))
				balanceToSet := big.NewInt(10000)
				nonceToSet := big.NewInt(1)

				err := insertAccount(&connector, addressToSet, balanceToSet, nonceToSet)
				suite.Require().NoError(err)

				// Encode request
				request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_ContainsKey{
						ContainsKey: &librustgo.QueryContainsKey{
							Key: addressToSet.Bytes(),
						},
					},
				})
				suite.Require().NoError(encodeErr)

				responseBytes, queryErr := connector.Query(request)
				suite.Require().NoError(queryErr)

				response := &librustgo.QueryContainsKeyResponse{}
				decodingError := proto.Unmarshal(responseBytes, response)
				suite.Require().NoError(decodingError)

				suite.Require().True(response.Contains)
			},
		},
		{
			"Should be able to get account data",
			func() {
				addressToSet := common.BigToAddress(big.NewInt(rand.Int63n(100000)))
				balanceToSet := big.NewInt(1400)
				nonceToSet := big.NewInt(22)

				err := insertAccount(&connector, addressToSet, balanceToSet, nonceToSet)
				suite.Require().NoError(err)

				// Encode request
				request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_GetAccount{
						GetAccount: &librustgo.QueryGetAccount{
							Address: addressToSet.Bytes(),
						},
					},
				})
				suite.Require().NoError(encodeErr)

				responseBytes, queryErr := connector.Query(request)
				suite.Require().NoError(queryErr)

				response := &librustgo.QueryGetAccountResponse{}
				decodingError := proto.Unmarshal(responseBytes, response)
				suite.Require().NoError(decodingError)

				returnedBalance := &big.Int{}
				returnedBalance.SetBytes(response.Balance)
				suite.Require().Equal(balanceToSet, returnedBalance)

				returnedNonce := response.Nonce
				suite.Require().Equal(nonceToSet.Uint64(), returnedNonce)
			},
		},
		{
			"Should be able to set account code",
			func() {
				var err error

				// Arrange
				addressToSet := common.BigToAddress(big.NewInt(rand.Int63n(100000)))
				bytecode := make([]byte, 32)
				rand.Read(bytecode)

				err = insertAccount(&connector, addressToSet, big.NewInt(0), big.NewInt(1))
				suite.Require().NoError(err)

				// Encode request
				request, err := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_InsertAccountCode{
						InsertAccountCode: &librustgo.QueryInsertAccountCode{
							Address: addressToSet.Bytes(),
							Code:    bytecode,
						},
					},
				})
				suite.Require().NoError(err)

				// Make a query
				_, err = connector.Query(request)
				suite.Require().NoError(err)

				// Check if account code was set correctly
				codeHash := crypto.Keccak256(bytecode)
				recoveredCode := connector.EVMKeeper.GetCode(connector.Context, common.BytesToHash(codeHash))
				suite.Require().Equal(bytecode, recoveredCode)

				account := connector.EVMKeeper.GetAccountOrEmpty(connector.Context, addressToSet)
				recoveredCodeHash := account.CodeHash
				suite.Require().Equal(codeHash, recoveredCodeHash)
			},
		},
		{
			"Should be able to set & get account code",
			func() {
				var err error

				addressToSet := common.BigToAddress(big.NewInt(rand.Int63n(100000)))
				bytecode := make([]byte, 128)
				rand.Read(bytecode)

				err = insertAccount(&connector, addressToSet, big.NewInt(0), big.NewInt(1))
				suite.Require().NoError(err)

				//
				// Insert account code
				//
				request, err := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_InsertAccountCode{
						InsertAccountCode: &librustgo.QueryInsertAccountCode{
							Address: addressToSet.Bytes(),
							Code:    bytecode,
						},
					},
				})
				suite.Require().NoError(err)

				responseBytes, err := connector.Query(request)
				suite.Require().NoError(err)

				response := &librustgo.QueryInsertAccountCodeResponse{}
				err = proto.Unmarshal(responseBytes, response)
				suite.Require().NoError(err)

				//
				// Request inserted account codesu
				//
				getRequest, err := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_AccountCode{
						AccountCode: &librustgo.QueryGetAccountCode{
							Address: addressToSet.Bytes(),
						},
					},
				})
				suite.Require().NoError(err)

				responseAccountCodeBytes, queryAccountCodeErr := connector.Query(getRequest)
				suite.Require().NoError(queryAccountCodeErr)

				accountCodeResponse := &librustgo.QueryGetAccountCodeResponse{}
				accCodeDecodingErr := proto.Unmarshal(responseAccountCodeBytes, accountCodeResponse)
				suite.Require().NoError(accCodeDecodingErr)
				suite.Require().Equal(bytecode, accountCodeResponse.Code)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			tc.action()
		})
	}
}

func requestAddVerificationDetails(
	connector *evmkeeper.Connector,
	userAddress common.Address,
	issuerAddress common.Address,
	verificationType compliancetypes.VerificationType,
	originChain string,
	issuanceTimestamp uint32,
	expirationTimestamp uint32,
	proofData []byte,
	schema string,
	issuerVerificationId string,
	version uint32,
) ([]byte, error) {
	// Encode request
	request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
		Req: &librustgo.CosmosRequest_AddVerificationDetails{
			AddVerificationDetails: &librustgo.QueryAddVerificationDetails{
				UserAddress:          userAddress.Bytes(),
				IssuerAddress:        issuerAddress.Bytes(),
				OriginChain:          originChain,
				VerificationType:     uint32(verificationType),
				IssuanceTimestamp:    issuanceTimestamp,
				ExpirationTimestamp:  expirationTimestamp,
				ProofData:            proofData,
				Schema:               schema,
				IssuerVerificationId: issuerVerificationId,
				Version:              version,
			},
		},
	})

	if encodeErr != nil {
		return nil, encodeErr
	}

	respBytes, queryErr := connector.Query(request)
	if queryErr != nil {
		return nil, queryErr
	}

	resp := &librustgo.QueryAddVerificationDetailsResponse{}
	decodeErr := proto.Unmarshal(respBytes, resp)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return resp.VerificationId, nil
}

func (suite *KeeperTestSuite) TestSingleVerificationDetails() {
	connector := evmkeeper.Connector{
		Context:   suite.ctx,
		EVMKeeper: suite.app.EvmKeeper,
	}

	var (
		userAddress          common.Address
		userAccount          sdk.AccAddress
		issuerAddress        common.Address
		issuerAccount        sdk.AccAddress
		illegalIssuerAccount sdk.AccAddress

		verificationType            = compliancetypes.VerificationType_VT_KYC
		expectedVerificationDetails *types.VerificationDetails
	)

	setup := func() {
		userAddress = tests.RandomEthAddress()
		userAccount = sdk.AccAddress(userAddress.Bytes())
		issuerAddress = tests.RandomEthAddress()
		issuerAccount = sdk.AccAddress(issuerAddress.Bytes())
		illegalIssuerAccount = tests.RandomAccAddress()

		// Verify issuer to add verification details which are verified by issuer
		_ = suite.app.ComplianceKeeper.SetIssuerDetails(suite.ctx, issuerAccount, &compliancetypes.IssuerDetails{
			Name: "test issuer",
		})
		_ = suite.app.ComplianceKeeper.SetAddressVerificationStatus(suite.ctx, issuerAccount, true)

		expectedVerificationDetails = &types.VerificationDetails{
			IssuerAddress:        issuerAccount.Bytes(),
			OriginChain:          "samplechain",
			IssuanceTimestamp:    uint32(suite.ctx.BlockTime().Unix()),
			ExpirationTimestamp:  uint32(0),
			OriginalData:         []byte("Proof Data"),
			Schema:               "Schema",
			IssuerVerificationId: "Issuer Verification ID",
			Version:              uint32(0),
		}
	}

	testCases := []struct {
		name   string
		action func(verificationID []byte)
	}{
		{
			"success - check verification from compliance keeper",
			func(verificationID []byte) {
				// Check if verification details exists in compliance keeper
				verificationDetails, err := suite.app.ComplianceKeeper.GetVerificationDetails(connector.Context, verificationID)
				suite.Require().NoError(err)
				// Addresses in compliance keeper are Cosmos Addresses
				// Addresses in Query requests are Ethereum Addresses
				suite.Require().Equal(issuerAccount.String(), verificationDetails.IssuerAddress)
				suite.Require().Equal(expectedVerificationDetails.OriginChain, verificationDetails.OriginChain)
				suite.Require().Equal(expectedVerificationDetails.IssuanceTimestamp, verificationDetails.IssuanceTimestamp)
				suite.Require().Equal(expectedVerificationDetails.ExpirationTimestamp, verificationDetails.ExpirationTimestamp)
				suite.Require().Equal(expectedVerificationDetails.OriginalData, verificationDetails.OriginalData)
				suite.Require().Equal(expectedVerificationDetails.Schema, verificationDetails.Schema)
				suite.Require().Equal(expectedVerificationDetails.IssuerVerificationId, verificationDetails.IssuerVerificationId)
				suite.Require().Equal(expectedVerificationDetails.Version, verificationDetails.Version)

				// Check if user has verification
				addressDetails, err := suite.app.ComplianceKeeper.GetAddressDetails(connector.Context, userAccount)
				suite.Require().Equal(1, len(addressDetails.Verifications))
				suite.Require().Equal(verificationType, addressDetails.Verifications[0].Type)
				suite.Require().Equal(verificationID, addressDetails.Verifications[0].VerificationId)
				suite.Require().Equal(issuerAccount.String(), addressDetails.Verifications[0].IssuerAddress)

				// Check if `hasVerification` with empty issuers returns true
				has, err := connector.EVMKeeper.ComplianceKeeper.HasVerificationOfType(connector.Context, userAccount, verificationType, nil)
				suite.Require().NoError(err)
				suite.Require().True(has)

				has, err = connector.EVMKeeper.ComplianceKeeper.HasVerificationOfType(connector.Context, userAccount, verificationType, []sdk.Address{issuerAccount})
				suite.Require().NoError(err)
				suite.Require().True(has)

				has, err = connector.EVMKeeper.ComplianceKeeper.HasVerificationOfType(connector.Context, userAccount, verificationType, []sdk.Address{illegalIssuerAccount})
				suite.Require().NoError(err)
				suite.Require().False(has)

				// Check if `getVerificationData` returns one verification details that added above
				verifications, verificationData, err := suite.app.ComplianceKeeper.GetVerificationDetailsByIssuer(connector.Context, userAccount, issuerAccount)
				suite.Require().NoError(err)
				suite.Require().Equal(1, len(verificationData))
				suite.Require().Equal(1, len(verifications))
				suite.Require().Equal(verificationDetails, verificationData[0])

				verifications, verificationData, err = suite.app.ComplianceKeeper.GetVerificationDetailsByIssuer(connector.Context, userAccount, illegalIssuerAccount)
				suite.Require().NoError(err)
				suite.Require().Equal(0, len(verifications))
				suite.Require().Equal(0, len(verificationData))
			},
		},
		{
			"success - check verification by HasVerification query",
			func(verificationID []byte) {
				// Encode request
				request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_HasVerification{
						HasVerification: &librustgo.QueryHasVerification{
							UserAddress:         userAddress.Bytes(),
							VerificationType:    uint32(verificationType),
							ExpirationTimestamp: uint32(expectedVerificationDetails.ExpirationTimestamp),
							AllowedIssuers:      nil,
						},
					},
				})
				suite.Require().NoError(encodeErr)

				respBytes, queryErr := connector.Query(request)
				suite.Require().NoError(queryErr)

				resp := &librustgo.QueryHasVerificationResponse{}
				decodeErr := proto.Unmarshal(respBytes, resp)
				suite.Require().NoError(decodeErr)

				suite.Require().True(resp.HasVerification)
			},
		},
		{
			"success - check verification by HasVerification query",
			func(verificationID []byte) {
				// Encode request
				request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_HasVerification{
						HasVerification: &librustgo.QueryHasVerification{
							UserAddress:         userAddress.Bytes(),
							VerificationType:    uint32(verificationType),
							ExpirationTimestamp: uint32(expectedVerificationDetails.ExpirationTimestamp),
							AllowedIssuers:      [][]byte{issuerAccount.Bytes()},
						},
					},
				})
				suite.Require().NoError(encodeErr)

				respBytes, queryErr := connector.Query(request)
				suite.Require().NoError(queryErr)

				resp := &librustgo.QueryHasVerificationResponse{}
				decodeErr := proto.Unmarshal(respBytes, resp)
				suite.Require().NoError(decodeErr)

				suite.Require().True(resp.HasVerification)
			},
		},
		{
			"success - check verification by GetVerificationData query",
			func(verificationID []byte) {
				// Encode request
				request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_GetVerificationData{
						GetVerificationData: &librustgo.QueryGetVerificationData{
							UserAddress:   userAddress.Bytes(),
							IssuerAddress: issuerAccount.Bytes(),
						},
					},
				})
				suite.Require().NoError(encodeErr)

				respBytes, queryErr := connector.Query(request)
				suite.Require().NoError(queryErr)

				resp := &librustgo.QueryGetVerificationDataResponse{}
				decodeErr := proto.Unmarshal(respBytes, resp)
				suite.Require().NoError(decodeErr)

				suite.Require().Equal(1, len(resp.Data))
				suite.Require().Equal(expectedVerificationDetails, resp.Data[0])

				// Should be empty for illegal issuer account
				request, encodeErr = proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_GetVerificationData{
						GetVerificationData: &librustgo.QueryGetVerificationData{
							UserAddress:   userAddress.Bytes(),
							IssuerAddress: illegalIssuerAccount.Bytes(),
						},
					},
				})
				suite.Require().NoError(encodeErr)
				respBytes, queryErr = connector.Query(request)
				suite.Require().NoError(queryErr)

				resp = &librustgo.QueryGetVerificationDataResponse{}
				decodeErr = proto.Unmarshal(respBytes, resp)
				suite.Require().NoError(decodeErr)

				suite.Require().Equal(0, len(resp.Data))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

			verificationID, err := requestAddVerificationDetails(
				&connector,
				userAddress,
				issuerAddress,
				verificationType,
				expectedVerificationDetails.OriginChain,
				expectedVerificationDetails.IssuanceTimestamp,
				expectedVerificationDetails.ExpirationTimestamp,
				expectedVerificationDetails.OriginalData,
				expectedVerificationDetails.Schema,
				expectedVerificationDetails.IssuerVerificationId,
				0,
			)
			suite.Require().NoError(err)
			suite.Require().NotNil(verificationID)

			tc.action(verificationID)
		})
	}
}

func (suite *KeeperTestSuite) TestMultipleVerificationDetails() {
	// Add multiple verification details

	var (
		userAddress   = tests.RandomEthAddress()
		issuerAddress = tests.RandomEthAddress()
		issuerAccount = sdk.AccAddress(issuerAddress.Bytes())

		expected []*types.VerificationDetails

		verificationType = compliancetypes.VerificationType_VT_KYC
	)

	// Verify issuer to add verification details which are verified by issuer
	_ = suite.app.ComplianceKeeper.SetIssuerDetails(suite.ctx, issuerAccount, &compliancetypes.IssuerDetails{
		Name: "test issuer",
	})
	_ = suite.app.ComplianceKeeper.SetAddressVerificationStatus(suite.ctx, issuerAccount, true)

	connector := evmkeeper.Connector{
		Context:   suite.ctx,
		EVMKeeper: suite.app.EvmKeeper,
	}

	numOfRetry := 10
	for i := 0; i < numOfRetry; i++ {
		// Addresses before making Query request should be Ethereum Addresses
		verificationDetails := &types.VerificationDetails{
			IssuerAddress:        issuerAddress.Bytes(),
			OriginChain:          "samplechain",
			IssuanceTimestamp:    uint32(suite.ctx.BlockTime().Unix()),
			ExpirationTimestamp:  uint32(0),
			OriginalData:         big.NewInt(rand.Int63n(100000)).Bytes(),
			Schema:               "HelloWorld",
			IssuerVerificationId: "HelloIssuer",
			Version:              uint32(0),
		}
		expected = append(expected, verificationDetails)

		verificationID, err := requestAddVerificationDetails(
			&connector,
			userAddress,
			issuerAddress,
			verificationType,
			verificationDetails.OriginChain,
			verificationDetails.IssuanceTimestamp,
			verificationDetails.ExpirationTimestamp,
			verificationDetails.OriginalData,
			verificationDetails.Schema,
			verificationDetails.IssuerVerificationId,
			0,
		)
		suite.Require().NoError(err)
		suite.Require().NotNil(verificationID)
	}

	request, encodeErr := proto.Marshal(&librustgo.CosmosRequest{
		Req: &librustgo.CosmosRequest_GetVerificationData{
			GetVerificationData: &librustgo.QueryGetVerificationData{
				UserAddress:   userAddress.Bytes(),
				IssuerAddress: issuerAccount.Bytes(),
			},
		},
	})
	suite.Require().NoError(encodeErr)

	respBytes, queryErr := connector.Query(request)
	suite.Require().NoError(queryErr)

	resp := &librustgo.QueryGetVerificationDataResponse{}
	decodeErr := proto.Unmarshal(respBytes, resp)
	suite.Require().NoError(decodeErr)
	suite.Require().Equal(numOfRetry, len(resp.Data))
	for i, details := range resp.Data {
		suite.Require().Equal(expected[i].IssuerAddress, details.IssuerAddress)
		suite.Require().Equal(expected[i].OriginChain, details.OriginChain)
		suite.Require().Equal(expected[i].IssuanceTimestamp, details.IssuanceTimestamp)
		suite.Require().Equal(expected[i].ExpirationTimestamp, details.ExpirationTimestamp)
		suite.Require().Equal(expected[i].OriginalData, details.OriginalData)
		suite.Require().Equal(expected[i].Schema, details.Schema)
		suite.Require().Equal(expected[i].IssuerVerificationId, details.IssuerVerificationId)
		suite.Require().Equal(expected[i].Version, details.Version)
	}
}
