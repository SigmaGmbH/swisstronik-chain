package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/iden3/go-merkletree-sql"
	"math/big"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

func (suite *KeeperTestSuite) TestSetSMTRoot() {
	suite.Setup(suite.T())

	ctx := sdk.WrapSDKContext(suite.ctx)
	storage := keeper.NewTreeStorage(suite.ctx, &suite.keeper, types.KeyPrefixRevocationTree)

	r, err := merkletree.NewHashFromBigInt(big.NewInt(1))
	suite.Require().NoError(err)

	err = storage.SetRoot(ctx, r)
	suite.Require().NoError(err)

	root, err := storage.GetRoot(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(r, root)
}

func (suite *KeeperTestSuite) TestPutAndGet() {
	suite.Setup(suite.T())

	ctx := sdk.WrapSDKContext(suite.ctx)
	storage := keeper.NewTreeStorage(suite.ctx, &suite.keeper, types.KeyPrefixRevocationTree)

	key, err := merkletree.NewHashFromBigInt(big.NewInt(1))
	suite.Require().NoError(err)

	value, err := merkletree.NewHashFromBigInt(big.NewInt(1))
	suite.Require().NoError(err)

	node := merkletree.NewNodeLeaf(key, value)

	err = storage.Put(ctx, key.BigInt().Bytes(), node)
	suite.Require().NoError(err)

	res, err := storage.Get(ctx, key.BigInt().Bytes())
	suite.Require().NoError(err)

	suite.Require().Equal(node, res)
}

func (suite *KeeperTestSuite) TestList() {
	suite.Setup(suite.T())

	ctx := sdk.WrapSDKContext(suite.ctx)
	storage := keeper.NewTreeStorage(suite.ctx, &suite.keeper, types.KeyPrefixIssuanceTree)

	key, err := merkletree.NewHashFromBigInt(big.NewInt(1))
	suite.Require().NoError(err)
	value, err := merkletree.NewHashFromBigInt(big.NewInt(1))
	suite.Require().NoError(err)

	n := merkletree.NewNodeLeaf(key, value)

	err = storage.Put(ctx, key.BigInt().Bytes(), n)
	suite.Require().NoError(err)

	// put another node
	key2, err := merkletree.NewHashFromBigInt(big.NewInt(2))
	suite.Require().NoError(err)

	value2, err := merkletree.NewHashFromBigInt(big.NewInt(2))
	suite.Require().NoError(err)

	n2 := merkletree.NewNodeLeaf(key2, value2)
	err = storage.Put(ctx, key2.BigInt().Bytes(), n2)
	suite.Require().NoError(err)

	got, err := storage.List(ctx, 0)
	suite.Require().NoError(err)

	exp := []merkletree.KV{{K: key.BigInt().Bytes(), V: *n}, {K: key2.BigInt().Bytes(), V: *n2}}
	suite.Require().Equal(exp, got)
}
