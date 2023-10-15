package keeper_test

import (
	"math/big"
	"math/rand"
	evmkeeper "swisstronik/x/evm/keeper"
	"time"

	"swisstronik/go-sgxvm"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang/protobuf/proto"

	didtypes "swisstronik/x/did/types"
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
		{
			"Should be able to verify JWT VC",
			func() {
				var err error

				// Create DID Document for issuer
				metadata := didtypes.Metadata{
					Created:   time.Now(),
					VersionId: "123e4567-e89b-12d3-a456-426655440000",
				}
				didUrl := "did:swtr:2MGhkRKWKi7ztnBFa8DpQ3"
				verificationMethods := []*didtypes.VerificationMethod{{
					Id:                     "did:swtr:2MGhkRKWKi7ztnBFa8DpQ3#6c1527f2f57601ea2f481a0ab1e605d196f3d952689299491638925cd6f26a7e-1",
					VerificationMethodType: "Ed25519VerificationKey2020",
					Controller:             didUrl,
					VerificationMaterial:   "z6MkmjAncvMFDiqyquFLt2G3CGYaqLfDjqKuzJnmUX3y68JZ",
				}}
				document := didtypes.DIDDocument{
					Id:                 didUrl,
					Controller:         []string{didUrl},
					VerificationMethod: verificationMethods,
					Authentication:     []string{"did:swtr:2MGhkRKWKi7ztnBFa8DpQ3#6c1527f2f57601ea2f481a0ab1e605d196f3d952689299491638925cd6f26a7e-1"},
				}
				didDocument := didtypes.DIDDocumentWithMetadata{
					Metadata: &metadata,
					DidDoc:   &document,
				}
				err = connector.EVMKeeper.DIDKeeper.AddNewDIDDocumentVersion(connector.Context, &didDocument)
				suite.Require().NoError(err)

				// Encode request for DID
				request, err := proto.Marshal(&librustgo.CosmosRequest{
					Req: &librustgo.CosmosRequest_VerificationMethods{
						VerificationMethods: &librustgo.QueryVerificationMethods{
							Did: didUrl,
						},
					},
				})
				suite.Require().NoError(err)

				response, err := connector.Query(request)
				suite.Require().NoError(err)

				vmResponse := &librustgo.QueryVerificationMethodsResponse{}
				err = proto.Unmarshal(response, vmResponse)
				suite.Require().NoError(err)

				// Extract verification methods
				expected := []*librustgo.VerificationMethod{}
				for _, method := range verificationMethods {
					ffiMethod := librustgo.VerificationMethod{
						VerificationMethodType: method.VerificationMethodType,
						VerificationMaterial:   method.VerificationMaterial,
					}
					expected = append(expected, &ffiMethod)
				}

				suite.Require().Equal(expected, vmResponse.Vm)
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
