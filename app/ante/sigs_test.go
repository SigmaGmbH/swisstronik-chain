package ante_test

import (
	"math/big"

	"swisstronik/tests"
	evmtypes "swisstronik/x/evm/types"
)

func (suite *AnteTestSuite) TestSignatures() {
	suite.enableFeemarket = false
	suite.SetupTest() // reset

	addr, privKey := tests.RandomEthAddressWithPrivateKey()
	to := tests.RandomEthAddress()

	acc := evmtypes.NewEmptyAccount()
	acc.Nonce = 1
	acc.Balance = big.NewInt(10000000000)

	_ = suite.app.EvmKeeper.SetAccount(suite.ctx, addr, *acc)
	msgHandleTx := evmtypes.NewTx(suite.app.EvmKeeper.ChainID(), 1, &to, big.NewInt(10), 100000, big.NewInt(1), nil, nil, nil, nil, nil, nil)
	msgHandleTx.From = addr.Hex()

	// CreateTestTx will sign the msgEthereumTx but not sign the cosmos tx since we have signCosmosTx as false
	tx := suite.CreateTestTx(msgHandleTx, privKey, 1, false)
	sigs, err := tx.GetSignaturesV2()
	suite.Require().NoError(err)

	// signatures of cosmos tx should be empty
	suite.Require().Equal(len(sigs), 0)

	txData, err := evmtypes.UnpackTxData(msgHandleTx.Data)
	suite.Require().NoError(err)

	msgV, msgR, msgS := txData.GetRawSignatureValues()

	ethTx := msgHandleTx.AsTransaction()
	ethV, ethR, ethS := ethTx.RawSignatureValues()

	// The signatures of MsgehtereumTx should be the same with the corresponding eth tx
	suite.Require().Equal(msgV, ethV)
	suite.Require().Equal(msgR, ethR)
	suite.Require().Equal(msgS, ethS)
}
