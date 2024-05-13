package keeper_test

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/app/ante"
	"swisstronik/crypto/ethsecp256k1"
	"swisstronik/tests"
	"swisstronik/testutil"
	utiltx "swisstronik/testutil/tx"
	"swisstronik/utils"
	feemarkettypes "swisstronik/x/feemarket/types"
	"swisstronik/x/vesting/keeper"
	"swisstronik/x/vesting/types"
)

type VestingTestSuite struct {
	suite.Suite

	ctx        sdk.Context
	goCtx      context.Context
	validator  stakingtypes.Validator
	app        *app.App
	querier    keeper.Querier
	msgServer  types.MsgServer
	stkQuerier stakingkeeper.Querier
}

func TestVestingTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Compliance Keeper Suite")
}

// Commit commits and starts a new block with an updated context.
func (suite *VestingTestSuite) Commit() error {
	return suite.CommitAfter(time.Second * 0)
}

// Commit commits a block at a given time.
func (suite *VestingTestSuite) CommitAfter(t time.Duration) error {
	var err error
	suite.ctx, err = testutil.CommitAndCreateNewCtx(suite.ctx, suite.app, t)
	suite.goCtx = sdk.WrapSDKContext(suite.ctx)
	return err
}

func (suite *VestingTestSuite) validateEthVestingTransactionDecorator(msgs ...sdk.Msg) error {
	dec := ante.NewEthVestingTransactionDecorator(suite.app.AccountKeeper, suite.app.BankKeeper, suite.app.EvmKeeper)
	return testutil.ValidateAnteForMsgs(suite.ctx, dec, msgs...)
}

func (suite *VestingTestSuite) assertFailEthNative(fromPrivKey cryptotypes.PrivKey, msgs ...sdk.Msg) {
	const insufficient = "insufficient"

	// Validate ante handler
	err := suite.validateEthVestingTransactionDecorator(msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficient))

	_, err = testutil.DeliverEthTx(suite.app, fromPrivKey, msgs...)
	Expect(err).ToNot(BeNil())
	Expect(strings.Contains(err.Error(), insufficient))
}

func (suite *VestingTestSuite) assertSuccessEthNative(src, dest sdk.AccAddress, amount math.Int, denom string, fromPrivKey cryptotypes.PrivKey, msgs ...sdk.Msg) {
	srcEthAddr := common.BytesToAddress(src.Bytes())
	destEthAddr := common.BytesToAddress(dest.Bytes())

	srcEvmBalanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, srcEthAddr)
	Expect(srcEvmBalanceBefore.Cmp(big.NewInt(0))).Should(BeNumerically(">", 0))
	destEvmBalanceBefore := suite.app.EvmKeeper.GetBalance(suite.ctx, destEthAddr)
	srcBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, src, denom)
	destBalanceBefore := suite.app.BankKeeper.GetBalance(suite.ctx, dest, denom)

	// Validate ante handler
	err := suite.validateEthVestingTransactionDecorator(msgs...)
	Expect(err).To(BeNil())

	_, err = testutil.DeliverEthTx(suite.app, fromPrivKey, msgs...)
	Expect(err).To(BeNil())

	srcEvmBalanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, srcEthAddr)
	dstEvmBalanceAfter := suite.app.EvmKeeper.GetBalance(suite.ctx, destEthAddr)
	srcBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, src, denom)
	destBalanceAfter := suite.app.BankKeeper.GetBalance(suite.ctx, dest, denom)
	Expect(new(big.Int).Sub(dstEvmBalanceAfter, destEvmBalanceBefore).Uint64()).Should(BeNumerically("==", amount.Uint64()))
	Expect(destBalanceBefore.AddAmount(amount).Amount.Uint64()).Should(BeNumerically("==", destBalanceAfter.Amount.Uint64()))
	// Reduced balance should be greater or equal to amount, because gas fee is non-recoverable
	Expect(new(big.Int).Sub(srcEvmBalanceBefore, srcEvmBalanceAfter).Uint64()).Should(BeNumerically(">=", amount.Uint64()))
	Expect(srcBalanceBefore.SubAmount(amount).Amount.Uint64()).Should(BeNumerically(">=", srcBalanceAfter.Amount.Uint64()))
}

func (suite *VestingTestSuite) SetupTest() error {
	chainID := utils.TestnetChainID + "-1"
	appS, _ := app.SetupSwissApp(false, nil, chainID)
	suite.app = appS

	address := tests.RandomAccAddress()

	privCons, err := ethsecp256k1.GenerateKey()
	if err != nil {
		return err
	}
	consAddress := sdk.ConsAddress(privCons.PubKey().Address())

	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, consAddress, nil, nil,
	)
	suite.ctx = appS.BaseApp.NewContext(false, header)
	suite.goCtx = sdk.WrapSDKContext(suite.ctx)

	stakingParams := appS.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	err = appS.StakingKeeper.SetParams(suite.ctx, stakingParams)
	if err != nil {
		return err
	}

	feeParams := feemarkettypes.DefaultParams()
	feeParams.MinGasPrice = sdk.NewDec(1)
	err = appS.FeeMarketKeeper.SetParams(suite.ctx, feeParams)
	if err != nil {
		return err
	}
	appS.FeeMarketKeeper.SetBaseFee(suite.ctx, sdk.ZeroInt().BigInt())

	// Set Validator
	valAddr := sdk.ValAddress(address.Bytes())
	validator, err := stakingtypes.NewValidator(valAddr, privCons.PubKey(), stakingtypes.Description{})
	if err != nil {
		return err
	}
	validator = stakingkeeper.TestingUpdateValidator(&appS.StakingKeeper, suite.ctx, validator, true)
	err = appS.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, validator.GetOperator())
	if err != nil {
		return err
	}
	err = appS.StakingKeeper.SetValidatorByConsAddr(suite.ctx, validator)
	if err != nil {
		return err
	}

	suite.querier = keeper.Querier{Keeper: appS.VestingKeeper}
	suite.msgServer = keeper.NewMsgServerImpl(appS.VestingKeeper)
	suite.stkQuerier = stakingkeeper.Querier{Keeper: &appS.StakingKeeper}

	validators := appS.StakingKeeper.GetValidators(suite.ctx, 2)
	// Set a bonded validator that takes part in consensus
	if validators[0].Status == stakingtypes.Bonded {
		suite.validator = validators[0]
	} else {
		suite.validator = validators[1]
	}
	return nil
}

var _ = Describe("Monthly Vesting Account", Ordered, func() {
	const (
		cliffDays = 30
		months    = 3
	)

	var (
		s *VestingTestSuite

		vaPrivKey cryptotypes.PrivKey // private key of vesting account
		va        sdk.AccAddress      // vesting account
		funder    sdk.AccAddress      // funder account who initiates monthly vesting
		user      sdk.AccAddress

		initialVesting sdk.Coins // initial vesting coins
		extraCoins     sdk.Coins // additional funds to pay gas fees for tx
		mva            *types.MonthlyVestingAccount

		unvested sdk.Coins
		vested   sdk.Coins
	)

	BeforeEach(func() {
		var err error

		s = new(VestingTestSuite)
		err = s.SetupTest()
		Expect(err).To(BeNil())

		now := s.ctx.BlockTime()

		var from common.Address
		from, vaPrivKey = tests.RandomEthAddressWithPrivateKey()
		va = sdk.AccAddress(from.Bytes())
		user = tests.RandomAccAddress()
		funder = tests.RandomAccAddress()

		amount := math.NewInt(1e17).Mul(math.NewInt(months))
		initialVesting = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount))

		// Fund coins to funder to create monthly vesting
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, funder, initialVesting)
		Expect(err).To(BeNil())
		err = s.Commit()

		// Create monthly vesting
		resp, err := s.msgServer.HandleCreateMonthlyVestingAccount(s.goCtx, &types.MsgCreateMonthlyVestingAccount{
			FromAddress: funder.String(),
			ToAddress:   va.String(),
			CliffDays:   cliffDays, // 30 days
			Months:      months,    // 3 months
			Amount:      initialVesting,
		})
		Expect(resp).To(Equal(&types.MsgCreateMonthlyVestingAccountResponse{}))
		err = s.Commit()
		Expect(err).To(BeNil())

		// Query monthly vesting account
		mva, err = s.querier.GetMonthlyVestingAccount(s.ctx, va)
		Expect(err).To(BeNil())
		Expect(mva).ToNot(BeNil())

		// Fund again as spendable coins for gas consuming
		extraCoins = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e6)))
		err = testutil.FundAccount(s.ctx, s.app.BankKeeper, mva.GetAddress(), extraCoins)
		Expect(err).To(BeNil())
		err = s.Commit()
		Expect(err).To(BeNil())

		// Check if all the tokens are unvested at beginning
		unvested := mva.GetVestingCoins(now)
		vested := mva.GetVestedCoins(now)
		Expect(unvested).To(Equal(initialVesting))
		Expect(vested.IsZero()).To(BeTrue())
	})

	Context("starting cliff days", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			err := s.Commit()
			Expect(err).To(BeNil())

			// Ensure no tokens are vested
			now := s.ctx.BlockTime()
			unvested = mva.GetVestingCoins(now)
			vested = mva.GetVestedCoins(now)
			Expect(unvested).To(Equal(initialVesting))
			Expect(vested.IsZero()).To(BeTrue())
		})
		It("can transfer spendable coins", func() {
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, mva.GetAddress(), unvested)
			Expect(err).To(BeNil())
			err = s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, unvested)
			Expect(err).To(BeNil())

			spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendable).To(Equal(extraCoins))
		})
		It("cannot transfer unvested coins", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, va, user, unvested)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("after cliff, before vested", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			duration := time.Duration(types.SecondsOfDay*cliffDays) * time.Second
			err := s.CommitAfter(duration)
			Expect(err).To(BeNil())

			// Ensure no tokens are vested
			now := s.ctx.BlockTime()
			unvested = mva.GetVestingCoins(now)
			vested = mva.GetVestedCoins(now)
			Expect(unvested).To(Equal(initialVesting))
			Expect(vested.IsZero()).To(BeTrue())
		})
		It("can transfer spendable coins", func() {
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, mva.GetAddress(), coins)
			Expect(err).To(BeNil())
			err = s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, coins)
			Expect(err).To(BeNil())

			spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendable).To(Equal(extraCoins))
		})
		It("cannot transfer unvested coins", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, va, user, unvested)
			Expect(err).ToNot(BeNil())
		})
		It("can delegate unvested coins", func() {
			// Delegate unvested tokens
			res, err := testutil.Delegate(
				s.ctx,
				s.app,
				vaPrivKey,
				unvested[0],
				s.validator,
			)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(unvested[0].Amount))

			// Check delegation was tracked as delegated vesting
			mva, err = s.app.VestingKeeper.GetMonthlyVestingAccount(s.ctx, mva.GetAddress())
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())
			Expect(mva.DelegatedVesting).To(Equal(unvested))
			Expect(mva.DelegatedFree).To(BeNil())

			// Fund more tokens for gas consuming
			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, mva.GetAddress(), extraCoins)
			Expect(err).To(BeNil())
		})
		It("can perform ethereum tx with spendable coins", func() {
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, mva.GetAddress(), coins)
			Expect(err).To(BeNil())

			amount := coins.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertSuccessEthNative(mva.GetAddress(), user, amount, utils.BaseDenom, vaPrivKey, msg)
		})
		It("cannot perform ethereum tx with unvested coins", func() {
			amount := unvested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertFailEthNative(vaPrivKey, msg)
		})
	})

	Context("after first vesting period", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			duration := time.Duration(types.SecondsOfDay*cliffDays) * time.Second
			duration = duration + time.Duration(types.SecondsOfMonth)*time.Second
			err := s.CommitAfter(duration)
			Expect(err).To(BeNil())

			// Check if 1/3 of initial vesting were vested
			now := s.ctx.BlockTime()
			vested = mva.GetVestedCoins(now)
			unvested = mva.GetVestingCoins(now)
			expVested := initialVesting.QuoInt(sdk.NewInt(months))
			Expect(vested).To(Equal(expVested))
			Expect(vested).ToNot(Equal(initialVesting))
			expUnvested := initialVesting.Sub(initialVesting.QuoInt(math.NewInt(3))...)
			Expect(unvested).To(Equal(expUnvested))

			// Check balances of vesting account
			balances, err := s.querier.Balances(s.goCtx, &types.QueryBalancesRequest{Address: mva.Address})
			Expect(err).To(BeNil())
			Expect(balances.Vested).To(Equal(vested))
			Expect(balances.Unvested).To(Equal(initialVesting.Sub(expVested...)))
			// All coins from vesting schedule should be locked
			Expect(balances.Locked).To(Equal(initialVesting.Sub(expVested...)))
		})
		It("can delegate vested tokens", func() {
			// Verify that the total spendable coins should include vested amount.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(vested.Add(extraCoins...)))

			// Delegate the vested coins
			_, err := testutil.Delegate(
				s.ctx,
				s.app,
				vaPrivKey,
				vested[0],
				s.validator,
			)
			Expect(err).To(BeNil())

			// Check spendable coins have not been reduced except gas fee.
			// Delegate unvested tokens first and then vested tokens.
			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(extraCoins...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(vested[0].Amount))

			// Check delegation was tracked as delegated vesting
			mva, err = s.app.VestingKeeper.GetMonthlyVestingAccount(s.ctx, mva.GetAddress())
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())
			Expect(mva.DelegatedVesting).To(Equal(vested))
			Expect(mva.DelegatedFree).To(BeNil())
		})
		It("can delegate unvested + vested tokens", func() {
			// Verify that the total spendable coins should include vested amount.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(vested.Add(extraCoins...)))

			// Delegate the vested coins
			_, err := testutil.Delegate(
				s.ctx,
				s.app,
				vaPrivKey,
				initialVesting[0],
				s.validator,
			)
			Expect(err).To(BeNil())

			// Delegate unvested tokens first and then vested tokens.
			// Check spendable coins have been reduced vested tokens as well.
			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(extraCoins...).Sub(vested...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(initialVesting[0].Amount))

			// Check delegation was tracked as delegated vesting/free
			mva, err = s.app.VestingKeeper.GetMonthlyVestingAccount(s.ctx, mva.GetAddress())
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())
			Expect(mva.DelegatedVesting).To(Equal(unvested))
			Expect(mva.DelegatedFree).To(Equal(vested))
		})
		It("can delegate tokens from account balance and initial vesting", func() {
			// Funds some coins to delegate
			coinsToDelegate := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, mva.GetAddress(), coinsToDelegate)
			Expect(err).To(BeNil())

			// Verify that the total spendable coins should include vested coins and newly funded coins.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(vested.Add(extraCoins...).Add(coinsToDelegate...)))

			// Delegate funds not in vesting schedule
			res, err := testutil.Delegate(
				s.ctx,
				s.app,
				vaPrivKey,
				coinsToDelegate[0].Add(initialVesting[0]),
				s.validator,
			)
			Expect(err).To(BeNil())
			Expect(res.IsOK()).To(BeTrue())

			// Check spendable balance is updated properly
			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(sdk.Coins{}))
		})
		It("can transfer vested tokens", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, vested)
			Expect(err).To(BeNil())
		})
		It("cannot transfer unvested tokens", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, unvested)
			Expect(err).ToNot(BeNil())
		})
		It("can perform ethereum tx with vested coins", func() {
			amount := vested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertSuccessEthNative(mva.GetAddress(), user, amount, utils.BaseDenom, vaPrivKey, msg)
		})
		It("cannot perform ethereum tx with unvested coins", func() {
			amount := unvested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertFailEthNative(vaPrivKey, msg)
		})
		It("cannot perform ethereum tx with vested coins if not sufficient coins for gas consuming", func() {
			// Send extra coins to another user to empty extra balance
			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, extraCoins)
			Expect(err).To(BeNil())

			// Try to send vested amount, but should be failed because of gas fee
			amount := vested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			err = s.validateEthVestingTransactionDecorator(msg)
			Expect(err).To(BeNil())
			_, err = testutil.DeliverEthTx(s.app, vaPrivKey, msg)
			Expect(err).ToNot(BeNil())
			Expect(err).Should(MatchError(ContainSubstring("Insert account failed. Empty response")))
		})
	})

	Context("after entire vesting period", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			duration := time.Duration(types.SecondsOfDay*cliffDays) * time.Second
			duration = duration + time.Duration(types.SecondsOfMonth*months)*time.Second
			err := s.CommitAfter(duration)
			Expect(err).To(BeNil())

			// Check if all the tokens of initial vesting were vested
			now := s.ctx.BlockTime()
			vested = mva.GetVestedCoins(now)
			unvested = mva.GetVestingCoins(now)
			Expect(vested).To(Equal(initialVesting))
			Expect(unvested).To(Equal(sdk.Coins{}))

			// Check balances of vesting account
			balances, err := s.querier.Balances(s.goCtx, &types.QueryBalancesRequest{Address: mva.Address})
			Expect(err).To(BeNil())
			Expect(balances.Vested).To(Equal(vested))
			Expect(balances.Unvested).To(Equal(unvested))
			Expect(balances.Locked).To(Equal(sdk.Coins{})) // no tokens were locked
		})
		It("can send entire initial vesting tokens", func() {
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(initialVesting.Add(extraCoins...)))

			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, initialVesting)
			Expect(err).To(BeNil())

			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(extraCoins))
		})
		It("can delegate initial vesting tokens", func() {
			// Verify that the total spendable coins should include initial vesting tokens.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(initialVesting.Add(extraCoins...)))

			// Delegate the initial vesting tokens
			_, err := testutil.Delegate(
				s.ctx,
				s.app,
				vaPrivKey,
				initialVesting[0],
				s.validator,
			)
			Expect(err).To(BeNil())

			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(extraCoins...).Sub(initialVesting...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(initialVesting[0].Amount))

			// No vesting tokens after vesting period.
			// Check delegation was tracked as delegated free
			mva, err = s.app.VestingKeeper.GetMonthlyVestingAccount(s.ctx, mva.GetAddress())
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())
			Expect(mva.DelegatedFree).To(Equal(initialVesting))
			Expect(mva.DelegatedVesting).To(BeNil())
		})
		It("can perform ethereum tx with initial vesting coins", func() {
			amount := initialVesting.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertSuccessEthNative(mva.GetAddress(), user, amount, utils.BaseDenom, vaPrivKey, msg)
		})
	})
})
