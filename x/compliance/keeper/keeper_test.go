package keeper_test

import (
	"context"
	"swisstronik/tests"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/utils"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

var s *KeeperTestSuite

type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	goCtx  context.Context
	keeper keeper.Keeper
	app    *app.App
}

func TestKeeperTestSuite(t *testing.T) {
	s = new(KeeperTestSuite)
	s.Setup(t)
	suite.Run(t, s)

	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compliance Keeper Suite")
}

func (suite *KeeperTestSuite) Setup(t *testing.T) {
	chainID := utils.TestnetChainID + "-1"
	app, _ := app.SetupSwissApp(false, nil, chainID)
	s.ctx = app.BaseApp.NewContext(false, tmproto.Header{ChainID: chainID})
	s.goCtx = sdk.WrapSDKContext(s.ctx)
	s.keeper = app.ComplianceKeeper
}

func (suite *KeeperTestSuite) TestCreateSimpleAndFetchSimpleIssuer() {

	details := &types.IssuerDetails{Name: "testIssuer"}
	from, _ := tests.RandomEthAddressWithPrivateKey()
	err := suite.keeper.SetIssuerDetails(suite.ctx,sdk.AccAddress(from.Bytes()), details)
	suite.Require().NoError(err)
}
