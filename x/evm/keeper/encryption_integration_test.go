package keeper_test

import "github.com/SigmaGmbH/librustgo"

func (suite *KeeperTestSuite) TestGetNodePublicKey() {
	var v1EpochStartingBlock uint64 = 2
	var v2EpochStartingBlock uint64 = 5

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	// Add 2 epochs
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	err = librustgo.AddEpoch(v2EpochStartingBlock)
	suite.Require().NoError(err)

	updatedEpochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(updatedEpochs), 3, "Should be 3 epochs")

	// Request node public key
	nodePublicKey, err := suite.app.EvmKeeper.GetNodePublicKey(0)
	suite.Require().NoError(err)
	nodePublicKeyV1, err := suite.app.EvmKeeper.GetNodePublicKey(v1EpochStartingBlock)
	suite.Require().NoError(err)
	nodePublicKeyV2, err := suite.app.EvmKeeper.GetNodePublicKey(v2EpochStartingBlock)
	suite.Require().NoError(err)

	suite.Require().NotEqual(nodePublicKey, nodePublicKeyV1)
	suite.Require().NotEqual(nodePublicKeyV1, nodePublicKeyV2)
}

func (suite *KeeperTestSuite) TestAddEpoch() {
	var v1EpochStartingBlock uint64 = 2

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	epochs, err := librustgo.ListEpochs()
	suite.Require().Equal(len(epochs), 1, "Should be only one epoch")
	suite.Require().Equal(epochs[0].EpochNumber, uint32(0), "First epoch should have 0 number")
	suite.Require().Equal(epochs[0].StartingBlock, uint64(0), "First epoch should have 0 starting block")

	// Add new epoch key, which starts from 2nd block
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	// Check updated epochs
	updatedEpochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(updatedEpochs), 2, "Should be two epochs")
	suite.Require().Equal(updatedEpochs[1].EpochNumber, uint32(1), "Second epoch should have 1 number")
	suite.Require().Equal(updatedEpochs[1].StartingBlock, v1EpochStartingBlock, "Incorrect epoch starting block")
}

func (suite *KeeperTestSuite) TestRemoveEpoch() {
	var v1EpochStartingBlock uint64 = 2
	var v2EpochStartingBlock uint64 = 3

	suite.SetupSGXVMTest()

	// Initialize empty key manager with genesis epoch key
	err := librustgo.InitializeEnclave(true)
	suite.Require().NoError(err)

	epochs, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochs), 1, "Should be only one epoch")
	suite.Require().Equal(epochs[0].EpochNumber, uint32(0), "First epoch should have 0 number")
	suite.Require().Equal(epochs[0].StartingBlock, uint64(0), "First epoch should have 0 starting block")

	// Should not be able to remove last epoch
	err = librustgo.RemoveLatestEpoch()
	suite.Require().Error(err)

	// Add another epoch
	err = librustgo.AddEpoch(v1EpochStartingBlock)
	suite.Require().NoError(err)

	// Check updated epochs
	epochsAfterAdd, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochsAfterAdd), 2, "Should be two epochs")
	suite.Require().Equal(epochsAfterAdd[1].EpochNumber, uint32(1), "Second epoch should have 1 number")
	suite.Require().Equal(epochsAfterAdd[1].StartingBlock, v1EpochStartingBlock, "Incorrect epoch starting block")

	// Should be able to remove latest epoch
	err = librustgo.RemoveLatestEpoch()
	suite.Require().NoError(err)

	epochsAfterRemoval, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochsAfterRemoval), 1, "Should be only one epoch")
	suite.Require().Equal(epochsAfterRemoval[0].EpochNumber, uint32(0), "First epoch should have 0 number")
	suite.Require().Equal(epochsAfterRemoval[0].StartingBlock, uint64(0), "First epoch should have 0 starting block")

	// Should be able to add epoch again with changed starting block
	err = librustgo.AddEpoch(v2EpochStartingBlock)
	suite.Require().NoError(err)

	epochsAfterNewAdd, err := librustgo.ListEpochs()
	suite.Require().NoError(err)
	suite.Require().Equal(len(epochsAfterNewAdd), 2, "Should be two epochs")
	suite.Require().Equal(epochsAfterNewAdd[1].EpochNumber, uint32(1), "Second epoch should have 1 number")
	suite.Require().Equal(epochsAfterNewAdd[1].StartingBlock, v2EpochStartingBlock, "Incorrect epoch starting block")
}
