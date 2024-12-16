package keeper_test

import (
	"github.com/iden3/go-merkletree-sql"
	"github.com/iden3/go-merkletree-sql/db/memory"
	"math/big"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

func (suite *KeeperTestSuite) TestProofGeneration() {
	suite.Setup(suite.T())

	storage := keeper.NewTreeStorage(suite.ctx, &suite.keeper, types.KeyPrefixRevocationTree)

	// construct storage-based merkle tree
	tree1, err := merkletree.NewMerkleTree(suite.ctx, &storage, 32)
	suite.Require().NoError(err)

	// construct memory-based merkle tree
	memoryStorage := memory.NewMemoryStorage()
	tree2, err := merkletree.NewMerkleTree(suite.ctx, memoryStorage, 32)
	suite.Require().NoError(err)

	k := big.NewInt(1)
	v := big.NewInt(2)

	// Add same leaf to both trees
	err = tree1.Add(suite.ctx, k, v)
	suite.Require().NoError(err)

	err = tree2.Add(suite.ctx, k, v)
	suite.Require().NoError(err)

	// Both trees should return the same proof
	proof1, _, err := tree1.GenerateProof(suite.ctx, k, nil)
	suite.Require().NoError(err)

	proof2, _, err := tree2.GenerateProof(suite.ctx, k, nil)
	suite.Require().NoError(err)

	suite.Require().Equal(proof1, proof2)
}
