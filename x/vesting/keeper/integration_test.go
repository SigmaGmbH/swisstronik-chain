package keeper_test

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/math"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/suite"

	"swisstronik/app"
	"swisstronik/app/ante"
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

	initialVesting sdk.Coins // initial vesting coins
	gasCoins       sdk.Coins // additional funds to pay gas fees for tx

	vaPrivKey cryptotypes.PrivKey // private key of vesting account
	va        sdk.AccAddress      // vesting account
	funder    sdk.AccAddress      // funder account who initiates monthly vesting
}

func TestVestingTestSuite(t *testing.T) {
	// Run Ginkgo integration tests
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vesting Keeper Suite")

	s := new(VestingTestSuite)
	suite.Run(t, s)
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

func (suite *VestingTestSuite) ExpectFundCoins(address sdk.AccAddress, coins sdk.Coins) {
	err := testutil.FundAccount(suite.ctx, suite.app.BankKeeper, address, coins)
	Expect(err).To(BeNil())
	err = suite.Commit()
	Expect(err).To(BeNil())
}

func (suite *VestingTestSuite) ExpectDelegateSuccess(privKey cryptotypes.PrivKey, delegating sdk.Coin, validator stakingtypes.Validator) {
	res, err := testutil.Delegate(
		suite.ctx,
		suite.app,
		privKey,
		delegating,
		validator,
	)
	Expect(err).To(BeNil())
	Expect(res.IsOK()).To(BeTrue())
	err = suite.Commit()
	Expect(err).To(BeNil())
}

func (suite *VestingTestSuite) ExpectUndelegateSuccess(privKey cryptotypes.PrivKey, undelegating sdk.Coin, validator stakingtypes.Validator) {
	// Undelegate delegated amount and consume gas
	res, err := testutil.Undelegate(
		suite.ctx,
		suite.app,
		privKey,
		undelegating,
		validator,
	)
	Expect(err).To(BeNil())
	Expect(res.IsOK()).To(BeTrue())

	// Wait unbonding time to finish undelegation
	ut := suite.app.StakingKeeper.UnbondingTime(suite.ctx)
	err = suite.CommitAfter(ut)
	Expect(err).To(BeNil())
	err = suite.Commit()
	Expect(err).To(BeNil())
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
	appS, _ := app.SetupSwissApp(nil, chainID)
	suite.app = appS

	address := tests.RandomAccAddress()

	pks := simtestutil.CreateTestPubKeys(1)
	consAddress := sdk.ConsAddress(pks[0].Address())

	header := testutil.NewHeader(
		1, time.Now().UTC(), chainID, consAddress, nil, nil,
	)
	suite.ctx = appS.BaseApp.NewContext(false, header)
	suite.goCtx = sdk.WrapSDKContext(suite.ctx)

	stakingParams := appS.StakingKeeper.GetParams(suite.ctx)
	stakingParams.BondDenom = utils.BaseDenom
	err := appS.StakingKeeper.SetParams(suite.ctx, stakingParams)
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
	validator, err := stakingtypes.NewValidator(valAddr, pks[0], stakingtypes.Description{})
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

func minCoins(a, b sdk.Coins) sdk.Coins {
	if a.IsAllGTE(b) {
		return b
	}
	return a
}

func subCoins(from, sub sdk.Coins) sdk.Coins {
	if sub == nil {
		return from
	}
	if from.IsAllGTE(sub) {
		return from.Sub(sub...)
	}
	return sdk.NewCoins()
}

func addCoins(from, add sdk.Coins) sdk.Coins {
	if add == nil {
		return from
	}
	return from.Add(add...)
}

func validateVestingAccountBalances(
	ctx sdk.Context,
	app *app.App,
	address sdk.AccAddress, // address to monthly vesting account
	prevDelegatedFree, prevDelegatedVesting sdk.Coins,
	delegating, undelegating sdk.Coins,
	expVested, expUnvested, initialBalance, sentAmount, receivedAmount, consumedFee sdk.Coins,
) {
	mva, err := app.VestingKeeper.GetMonthlyVestingAccount(ctx, address)
	Expect(err).To(BeNil())
	Expect(mva).ToNot(BeNil())

	var (
		expDelegatedVesting sdk.Coins = prevDelegatedVesting
		expDelegatedFree    sdk.Coins = prevDelegatedFree
	)
	if delegating != nil {
		X := minCoins(subCoins(expUnvested, prevDelegatedVesting), delegating)
		Y := subCoins(delegating, X)
		expDelegatedVesting = addCoins(expDelegatedVesting, X)
		expDelegatedFree = addCoins(expDelegatedFree, Y)
	}

	if undelegating != nil {
		X := minCoins(prevDelegatedFree, undelegating)
		Y := minCoins(prevDelegatedVesting, subCoins(undelegating, X))
		expDelegatedFree = subCoins(expDelegatedFree, X)
		expDelegatedVesting = subCoins(expDelegatedVesting, Y)
	}

	expLocked := subCoins(expUnvested, addCoins(expDelegatedFree, expDelegatedVesting))
	expBalances := addCoins(
		subCoins(
			subCoins(
				subCoins(
					addCoins(subCoins(initialBalance, sentAmount), receivedAmount),
					consumedFee,
				),
				addCoins(prevDelegatedVesting, prevDelegatedFree),
			),
			delegating,
		),
		undelegating,
	)
	expSpendable := subCoins(expBalances, expLocked)

	balancesX := app.BankKeeper.GetAllBalances(ctx, address)
	Expect(balancesX).To(Equal(expBalances))

	lockedX := app.BankKeeper.LockedCoins(ctx, address)
	if lockedX == nil {
		Expect(sdk.NewCoins()).To(Equal(expLocked))
	} else {
		Expect(lockedX).To(Equal(expLocked))
	}

	spendableX := app.BankKeeper.SpendableCoins(ctx, address)
	if spendableX == nil {
		Expect(sdk.NewCoins()).To(Equal(expSpendable))
	} else {
		Expect(spendableX).To(Equal(expSpendable))
	}

	now := ctx.BlockTime()
	vestedX := mva.GetVestedCoins(now)
	if vestedX == nil {
		Expect(sdk.NewCoins()).To(Equal(expVested))
	} else {
		Expect(vestedX).To(Equal(expVested))
	}
	unvestedX := mva.GetVestingCoins(now)
	if unvestedX == nil {
		Expect(sdk.NewCoins()).To(Equal(expUnvested))
	} else {
		Expect(unvestedX).To(Equal(expUnvested))
	}

	if mva.DelegatedFree == nil {
		Expect(sdk.NewCoins()).To(Equal(expDelegatedFree))
	} else {
		Expect(mva.DelegatedFree).To(Equal(expDelegatedFree))
	}
	if mva.DelegatedVesting == nil {
		Expect(sdk.NewCoins()).To(Equal(expDelegatedVesting))
	} else {
		Expect(mva.DelegatedVesting).To(Equal(expDelegatedVesting))
	}
}

func validateVestingAccountBalancesWithValues(
	ctx sdk.Context,
	app *app.App,
	address sdk.AccAddress, // address to monthly vesting account
	expDelegatedVesting, expDelegatedFree, expLocked, expVested, expUnvested, expBalances, expSpendable sdk.Coins) {
	mva, err := app.VestingKeeper.GetMonthlyVestingAccount(ctx, address)
	Expect(err).To(BeNil())
	Expect(mva).ToNot(BeNil())

	now := ctx.BlockTime()

	if mva.DelegatedVesting == nil {
		Expect(expDelegatedVesting).To(Equal(sdk.NewCoins()))
	} else {
		Expect(mva.DelegatedVesting).To(Equal(expDelegatedVesting))
	}
	if mva.DelegatedFree == nil {
		Expect(expDelegatedFree).To(Equal(sdk.NewCoins()))
	} else {
		Expect(mva.DelegatedFree).To(Equal(expDelegatedFree))
	}
	locked := app.BankKeeper.LockedCoins(ctx, mva.GetAddress())
	Expect(locked).To(Equal(expLocked))
	vested := mva.GetVestedCoins(now)
	if vested == nil {
		Expect(expVested).To(Equal(sdk.NewCoins()))
	} else {
		Expect(vested).To(Equal(expVested))
	}
	unvested := mva.GetVestingCoins(now)
	if unvested == nil {
		Expect(expUnvested).To(Equal(sdk.NewCoins()))
	} else {
		Expect(unvested).To(Equal(expUnvested))
	}
	balances := app.BankKeeper.GetAllBalances(ctx, mva.GetAddress())
	Expect(balances).To(Equal(expBalances))
	spendable := app.BankKeeper.SpendableCoins(ctx, mva.GetAddress())
	Expect(spendable).To(Equal(expSpendable))
}

var _ = Describe("Monthly Vesting Account", Ordered, func() {
	const (
		cliffDays = 30
		months    = 3
	)

	var (
		s *VestingTestSuite

		user sdk.AccAddress

		mva *types.MonthlyVestingAccount

		unvested    sdk.Coins
		vested      sdk.Coins
		expUnvested sdk.Coins
		expVested   sdk.Coins
	)

	BeforeEach(func() {
		var err error

		s = new(VestingTestSuite)
		err = s.SetupTest()
		Expect(err).To(BeNil())

		now := s.ctx.BlockTime()

		var from common.Address
		from, s.vaPrivKey = tests.RandomEthAddressWithPrivateKey()
		s.va = sdk.AccAddress(from.Bytes())
		user = tests.RandomAccAddress()
		s.funder = tests.RandomAccAddress()

		amount := math.NewInt(1e17).Mul(math.NewInt(months))
		s.initialVesting = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount))

		// Fund coins to funder to create monthly vesting
		s.ExpectFundCoins(s.funder, s.initialVesting)

		// Create monthly vesting
		resp, err := s.msgServer.HandleCreateMonthlyVestingAccount(s.goCtx, &types.MsgCreateMonthlyVestingAccount{
			FromAddress: s.funder.String(),
			ToAddress:   s.va.String(),
			CliffDays:   cliffDays, // 30 days
			Months:      months,    // 3 months
			Amount:      s.initialVesting,
		})
		Expect(resp).To(Equal(&types.MsgCreateMonthlyVestingAccountResponse{}))
		err = s.Commit()
		Expect(err).To(BeNil())

		// Query monthly vesting account
		mva, err = s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
		Expect(err).To(BeNil())
		Expect(mva).ToNot(BeNil())

		// Fund again as spendable coins for gas consuming
		s.gasCoins = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e6)))
		s.ExpectFundCoins(mva.GetAddress(), s.gasCoins)

		// Check if all the coins are unvested at beginning
		unvested := mva.GetVestingCoins(now)
		vested := mva.GetVestedCoins(now)
		Expect(unvested).To(Equal(s.initialVesting))
		Expect(vested.IsZero()).To(BeTrue())
	})

	Context("starting cliff days", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			err := s.Commit()
			Expect(err).To(BeNil())

			// Ensure no coins are vested
			now := s.ctx.BlockTime()
			unvested = mva.GetVestingCoins(now)
			vested = mva.GetVestedCoins(now)
			Expect(unvested).To(Equal(s.initialVesting))
			Expect(vested.IsZero()).To(BeTrue())

			expUnvested = s.initialVesting
			expVested = sdk.NewCoins()

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // vesting
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				sdk.NewCoins(),                      // consumed fee
			)
		})
		It("can transfer spendable coins", func() {
			s.ExpectFundCoins(s.va, unvested)

			err := s.app.BankKeeper.SendCoins(s.ctx, s.va, user, unvested)
			Expect(err).To(BeNil())

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(), // prev delegated free
				sdk.NewCoins(), // prev delegated vesting
				nil,            // delegating
				nil,            // undelegating
				expVested,      // vested
				expUnvested,    // unvested
				s.initialVesting.Add(s.gasCoins...).Add(unvested...), // initial balance
				unvested,       // sent amount
				sdk.NewCoins(), // received amount
				sdk.NewCoins(), // consumed fee
			)
		})
		It("cannot transfer unvested coins", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, s.va, user, unvested)
			Expect(err).ToNot(BeNil())
		})
		It("can delegate unvested coins", func() {
			// Delegate unvested coins
			delegating := unvested
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
	})

	Context("after cliff, before vested", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			duration := time.Duration(types.SecondsOfDay*cliffDays) * time.Second
			err := s.CommitAfter(duration)
			Expect(err).To(BeNil())

			// Ensure no coins are vested
			now := s.ctx.BlockTime()
			unvested = mva.GetVestingCoins(now)
			vested = mva.GetVestedCoins(now)
			Expect(unvested).To(Equal(s.initialVesting))
			Expect(vested.IsZero()).To(BeTrue())

			expUnvested = s.initialVesting
			expVested = sdk.NewCoins()

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				sdk.NewCoins(),                      // consumed fee
			)
		})
		It("can transfer spendable coins", func() {
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
			s.ExpectFundCoins(s.va, coins)

			err := s.app.BankKeeper.SendCoins(s.ctx, s.va, user, coins)
			Expect(err).To(BeNil())

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(), // prev delegated free
				sdk.NewCoins(), // prev delegated vesting
				nil,            // delegating
				nil,            // undelegating
				expVested,      // vested
				expUnvested,    // unvested
				s.initialVesting.Add(s.gasCoins...).Add(coins...), // initial balance
				coins,          // sent amount
				sdk.NewCoins(), // received amount
				sdk.NewCoins(), // consumed fee
			)
		})
		It("cannot transfer unvested coins", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, s.va, user, unvested)
			Expect(err).ToNot(BeNil())
		})
		It("can delegate portion of unvested coins", func() {
			// Delegate portion of unvested and consume gas
			delegating := unvested.Sub(s.gasCoins...)
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)
			// Undelegate portion of delegated amount and consume gas
			undelegating := delegating.QuoInt(sdk.NewInt(2))
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can delegate all unvested coins", func() {
			// Delegate coins more than unvested, and consume gas
			res, err := testutil.Delegate(
				s.ctx,
				s.app,
				s.vaPrivKey,
				unvested[0].Add(s.gasCoins[0]),
				s.validator,
			)
			Expect(err).ToNot(BeNil())
			Expect(err).Should(MatchError(ContainSubstring("insufficient funds")))
			Expect(res.IsOK()).To(BeTrue())

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)

			// Delegate unvested coins, consume gas
			delegating := unvested
			res, err = testutil.Delegate(
				s.ctx,
				s.app,
				s.vaPrivKey,
				delegating[0],
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
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)
			// Undelegate all unvested amount and consume gas
			undelegating := delegating
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can perform ethereum tx with spendable coins", func() {
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
			s.ExpectFundCoins(s.va, coins)

			amount := coins.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, s.vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertSuccessEthNative(mva.GetAddress(), user, amount, utils.BaseDenom, s.vaPrivKey, msg)
		})
		It("cannot perform ethereum tx with unvested coins", func() {
			amount := unvested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, s.vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertFailEthNative(s.vaPrivKey, msg)
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
			expVested = s.initialVesting.QuoInt(sdk.NewInt(months))
			expUnvested = s.initialVesting.Sub(s.initialVesting.QuoInt(sdk.NewInt(months))...)
			Expect(vested).To(Equal(expVested))
			Expect(unvested).To(Equal(expUnvested))

			// Check balances of vesting account
			balances, err := s.querier.Balances(s.goCtx, &types.QueryBalancesRequest{Address: mva.Address})
			Expect(err).To(BeNil())
			Expect(balances.Vested).To(Equal(expVested))
			Expect(balances.Unvested).To(Equal(expUnvested))
			// All coins from vesting schedule should be locked
			Expect(balances.Locked).To(Equal(s.initialVesting.Sub(expVested...)))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				sdk.NewCoins(),                      // consumed fee
			)
		})
		It("can delegate portion of vested coins", func() {
			// Verify that the total spendable coins should include vested amount.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(vested.Add(s.gasCoins...)))

			// Delegate the vested coins, consume gas
			delegating := vested.QuoInt(sdk.NewInt(2))
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			// Check spendable coins have not been reduced except gas fee.
			// Delegate unvested coins first and then vested coins.
			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(s.gasCoins...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)
			// Undelegate more than delegated amount and consume gas
			undelegating := vested
			res, err := testutil.Undelegate(
				s.ctx,
				s.app,
				s.vaPrivKey,
				undelegating[0],
				s.validator,
			)
			Expect(err).ToNot(BeNil())
			Expect(err).Should(MatchError(ContainSubstring("invalid shares amount")))
			Expect(res.IsOK()).To(BeTrue())

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)
			// Undelegate more than delegated amount and consume gas
			undelegating = delegating
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can delegate unvested + vested coins", func() {
			// Verify that the total spendable coins should include vested amount.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(vested.Add(s.gasCoins...)))

			// Delegate the vested coins, consume gas
			delegating := s.initialVesting
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			// Delegate unvested coins first and then vested coins.
			// Check spendable coins have been reduced vested coins as well.
			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(s.gasCoins...).Sub(vested...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)
			// Undelegate portion of delegated amount and consume gas
			undelegating := delegating
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can delegate coins from account balance and initial vesting", func() {
			// Funds some coins to delegate
			coinsToDelegate := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e18)))
			s.ExpectFundCoins(s.va, coinsToDelegate)

			// Verify that the total spendable coins should include vested coins and newly funded coins.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(vested.Add(s.gasCoins...).Add(coinsToDelegate...)))

			// Delegate funds not in vesting schedule
			delegating := sdk.NewCoins(coinsToDelegate.Add(s.initialVesting...)...)
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			// Check spendable balance is updated properly
			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(sdk.NewCoins()))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can transfer vested coins", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, vested)
			Expect(err).To(BeNil())

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				vested,                              // sent amount
				sdk.NewCoins(),                      // received amount
				sdk.NewCoins(),                      // consumed fee
			)
		})
		It("cannot transfer unvested coins", func() {
			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, unvested)
			Expect(err).ToNot(BeNil())
		})
		It("can perform ethereum tx with vested coins", func() {
			amount := vested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, s.vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertSuccessEthNative(mva.GetAddress(), user, amount, utils.BaseDenom, s.vaPrivKey, msg)
		})
		It("cannot perform ethereum tx with unvested coins", func() {
			amount := unvested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, s.vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertFailEthNative(s.vaPrivKey, msg)
		})
		It("cannot perform ethereum tx with vested coins if not sufficient coins for gas consuming", func() {
			// Send extra coins to another user to empty extra balance
			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, s.gasCoins)
			Expect(err).To(BeNil())

			// Try to send vested amount, but should be failed because of gas fee
			amount := vested.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, s.vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			err = s.validateEthVestingTransactionDecorator(msg)
			Expect(err).To(BeNil())
			_, err = testutil.DeliverEthTx(s.app, s.vaPrivKey, msg)
			Expect(err).ToNot(BeNil())
			Expect(err).Should(MatchError(ContainSubstring("Insert account balance failed. Empty response")))
		})
	})

	Context("after entire vesting period", func() {
		BeforeEach(func() {
			// Add a commit to instantiate blocks
			duration := time.Duration(types.SecondsOfDay*cliffDays) * time.Second
			duration = duration + time.Duration(types.SecondsOfMonth*months)*time.Second
			err := s.CommitAfter(duration)
			Expect(err).To(BeNil())

			// Check if all the coins of initial vesting were vested
			now := s.ctx.BlockTime()
			vested = mva.GetVestedCoins(now)
			unvested = mva.GetVestingCoins(now)
			expVested = s.initialVesting
			expUnvested = sdk.NewCoins()
			Expect(vested).To(Equal(expVested))
			Expect(unvested).To(Equal(expUnvested))

			// Check balances of vesting account
			balances, err := s.querier.Balances(s.goCtx, &types.QueryBalancesRequest{Address: mva.Address})
			Expect(err).To(BeNil())
			Expect(balances.Vested).To(Equal(vested))
			Expect(balances.Unvested).To(Equal(unvested))
			Expect(balances.Locked).To(Equal(sdk.Coins{})) // no coins were locked

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				sdk.NewCoins(),                      // consumed fee
			)
		})
		It("can send entire initial vesting coins", func() {
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(s.initialVesting.Add(s.gasCoins...)))

			err := s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, s.initialVesting)
			Expect(err).To(BeNil())

			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(s.gasCoins))

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				s.initialVesting,                    // sent amount
				sdk.NewCoins(),                      // received amount
				sdk.NewCoins(),                      // consumed fee
			)
		})
		It("can delegate portion of initial vesting coins", func() {
			// Verify that the total spendable coins should include initial vesting coins.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(s.initialVesting.Add(s.gasCoins...)))

			// Delegate the initial vesting coins
			delegating := s.initialVesting.QuoInt(sdk.NewInt(3))
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(s.gasCoins...).Sub(delegating...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// No vesting coins after vesting period.
			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can delegate initial vesting coins", func() {
			// Verify that the total spendable coins should include initial vesting coins.
			spendablePre := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePre).To(Equal(s.initialVesting.Add(s.gasCoins...)))

			// Delegate the initial vesting coins
			delegating := s.initialVesting
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)

			spendablePost := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
			Expect(spendablePost).To(Equal(spendablePre.Sub(s.gasCoins...).Sub(delegating...)))

			// Check delegation was created successfully
			delegations, err := s.stkQuerier.DelegatorDelegations(
				s.goCtx,
				&stakingtypes.QueryDelegatorDelegationsRequest{
					DelegatorAddr: mva.Address,
				},
			)
			Expect(err).To(BeNil())
			Expect(delegations.DelegationResponses).To(HaveLen(1))
			Expect(delegations.DelegationResponses[0].Balance.Amount).To(Equal(delegating[0].Amount))

			// No vesting coins after vesting period.
			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			// Fund gas coins for next transaction
			s.ExpectFundCoins(s.va, s.gasCoins)
			// Undelegate portion of delegated amount and consume gas
			undelegating := delegating
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)
		})
		It("can perform ethereum tx with initial vesting coins", func() {
			amount := s.initialVesting.AmountOf(utils.BaseDenom)
			msg, err := utiltx.CreateEthTx(s.ctx, s.app, s.vaPrivKey, mva.GetAddress(), user, amount.BigInt(), 0)
			Expect(err).To(BeNil())
			s.assertSuccessEthNative(mva.GetAddress(), user, amount, utils.BaseDenom, s.vaPrivKey, msg)
		})
	})
})

var _ = Describe("Additional tests for Monthly Vesting Account", Ordered, func() {
	var (
		s *VestingTestSuite

		user sdk.AccAddress

		err error

		expUnvested sdk.Coins
		expVested   sdk.Coins
	)

	BeforeEach(func() {
		s = new(VestingTestSuite)
		err = s.SetupTest()
		Expect(err).To(BeNil())

		var from common.Address
		from, s.vaPrivKey = tests.RandomEthAddressWithPrivateKey()
		s.va = sdk.AccAddress(from.Bytes())
		user = tests.RandomAccAddress()
		s.funder = tests.RandomAccAddress()

		s.initialVesting = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1000, 18)))
		s.gasCoins = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e6)))

		// Fund coins to funder to create monthly vesting account
		s.ExpectFundCoins(s.funder, s.initialVesting)
	})

	Context("Create vesting account cliff days 10, month 12", func() {
		var delegating sdk.Coins

		BeforeEach(func() {
			// Create monthly vesting account
			resp, err := s.msgServer.HandleCreateMonthlyVestingAccount(s.goCtx, &types.MsgCreateMonthlyVestingAccount{
				FromAddress: s.funder.String(),
				ToAddress:   s.va.String(),
				CliffDays:   10,
				Months:      12,
				Amount:      s.initialVesting,
			})
			Expect(resp).To(Equal(&types.MsgCreateMonthlyVestingAccountResponse{}))
			err = s.Commit()
			Expect(err).To(BeNil())

			// Query monthly vesting account
			mva, err := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())

			// Fund coins to mva for gas consuming
			s.ExpectFundCoins(s.va, s.gasCoins)

			expUnvested = s.initialVesting
			expVested = sdk.NewCoins()

			// stake 500 swtr
			delegating = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)
		})
		It("validate", func() {
			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				delegating,                             // delegated vesting = 500 swtr
				sdk.NewCoins(),                         // delegated free = 0 swtr
				delegating,                             // locked = 500 swtr
				sdk.NewCoins(),                         // vested = 0 swtr
				s.initialVesting,                       // unvested = 1000 swtr
				subCoins(s.initialVesting, delegating), // balances = 500 swtr
				sdk.NewCoins(),                         // spendable balances = 0 swtr
			)
		})
		It("undelegate after cliff", func() {
			mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)

			// Pass after cliff
			duration := time.Duration(mva.CliffTime-s.ctx.BlockTime().Unix()) * time.Second
			err = s.CommitAfter(duration)
			Expect(err).To(BeNil())

			// Fund coins to mva for gas consuming
			s.ExpectFundCoins(s.va, s.gasCoins)

			undelegating := delegating
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			expUnvested = s.initialVesting
			expVested = sdk.NewCoins()

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues( // todo
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),   // delegated vesting = 0 swtr
				sdk.NewCoins(),   // delegated free = 0 swtr
				s.initialVesting, // locked = 1000 swtr
				sdk.NewCoins(),   // vested = 0 swtr
				s.initialVesting, // unvested = 1000 swtr
				s.initialVesting, // balances = 1000 swtr
				sdk.NewCoins(),   // spendable balances = 0 swtr
			)
		})
		It("wait until the end of vesting period", func() {
			mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)

			duration := time.Duration(mva.EndTime-s.ctx.BlockTime().Unix()) * time.Second
			err = s.CommitAfter(duration)
			Expect(err).To(BeNil())

			expVested = s.initialVesting
			expUnvested = sdk.NewCoins()

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				delegating,                             // delegated vesting = 500 swtr
				sdk.NewCoins(),                         // delegated free = 0 swtr
				sdk.NewCoins(),                         // locked = 0 swtr
				s.initialVesting,                       // vested = 1000 swtr
				sdk.NewCoins(),                         // unvested = 0 swtr
				subCoins(s.initialVesting, delegating), // balances = 500 swtr
				subCoins(s.initialVesting, delegating), // spendable balances = 500 swtr
			)
		})
		It("receive coins", func() {
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(300, 18)))
			s.ExpectFundCoins(s.va, coins)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				coins,                               // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				delegating,       // delegated vesting = 500 swtr
				sdk.NewCoins(),   // delegated free = 0 swtr
				delegating,       // locked = 500 swtr
				sdk.NewCoins(),   // vested = 0 swtr
				s.initialVesting, // unvested = 1000 swtr
				addCoins(subCoins(s.initialVesting, delegating), coins), // balances = 500 swtr + 300 swtr
				coins, // spendable balances = 300 swtr
			)
		})
		It("failed in sending coins", func() {
			mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)

			// Try to send 300 coins from monthly vesting account
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(300, 18)))
			err = s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, coins)
			Expect(err).ToNot(BeNil())
		})
	})

	Context("Create vesting account cliff days 0, month 12", func() {
		var delegating sdk.Coins

		BeforeEach(func() {
			// Create monthly vesting account
			resp, err := s.msgServer.HandleCreateMonthlyVestingAccount(s.goCtx, &types.MsgCreateMonthlyVestingAccount{
				FromAddress: s.funder.String(),
				ToAddress:   s.va.String(),
				CliffDays:   0,
				Months:      12,
				Amount:      s.initialVesting,
			})
			Expect(resp).To(Equal(&types.MsgCreateMonthlyVestingAccountResponse{}))
			err = s.Commit()
			Expect(err).To(BeNil())

			// Query monthly vesting account
			mva, err := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())

			// Fund coins to mva for gas consuming
			s.ExpectFundCoins(s.va, s.gasCoins)

			expUnvested = s.initialVesting
			expVested = sdk.NewCoins()

			// stake 500 swtr
			delegating = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
			s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], s.validator)
		})
		It("validate", func() {
			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				delegating,                             // delegated vesting = 500 swtr
				sdk.NewCoins(),                         // delegated free = 0 swtr
				delegating,                             // locked = 500 swtr
				sdk.NewCoins(),                         // vested = 0 swtr
				s.initialVesting,                       // unvested = 1000 swtr
				subCoins(s.initialVesting, delegating), // balances = 500 swtr
				sdk.NewCoins(),                         // spendable balances = 0 swtr
			)
		})
		It("undelegate after cliff", func() {
			// No need to pass, cliff 0

			// Fund coins to mva for gas consuming
			s.ExpectFundCoins(s.va, s.gasCoins)

			undelegating := delegating
			s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], s.validator)

			expUnvested = s.initialVesting
			expVested = sdk.NewCoins()

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				undelegating,                        // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),   // delegated vesting = 0 swtr
				sdk.NewCoins(),   // delegated free = 0 swtr
				s.initialVesting, // locked = 1000 swtr
				sdk.NewCoins(),   // vested = 0 swtr
				s.initialVesting, // unvested = 1000 swtr
				s.initialVesting, // balances = 1000 swtr
				sdk.NewCoins(),   // spendable balances = 0 swtr
			)
		})
		It("wait until the end of vesting period", func() {
			mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)

			duration := time.Duration(mva.EndTime-s.ctx.BlockTime().Unix()) * time.Second
			//duration := time.Duration(mva.EndTime-mva.StartTime) * time.Second
			err = s.CommitAfter(duration)
			Expect(err).To(BeNil())

			expVested = s.initialVesting
			expUnvested = sdk.NewCoins()

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				delegating,                          // prev delegated vesting
				nil,                                 // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				sdk.NewCoins(),                      // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				delegating,                             // delegated vesting = 500 swtr
				sdk.NewCoins(),                         // delegated free = 0 swtr
				sdk.NewCoins(),                         // locked = 0 swtr
				s.initialVesting,                       // vested = 1000 swtr
				sdk.NewCoins(),                         // unvested = 0 swtr
				subCoins(s.initialVesting, delegating), // balances = 500 swtr
				subCoins(s.initialVesting, delegating), // spendable balances = 500 swtr
			)
		})
		It("receive coins", func() {
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(300, 18)))
			s.ExpectFundCoins(s.va, coins)

			// Check all the balances of vesting account
			validateVestingAccountBalances(
				s.ctx,
				s.app,
				s.va,
				sdk.NewCoins(),                      // prev delegated free
				sdk.NewCoins(),                      // prev delegated vesting
				delegating,                          // delegating
				nil,                                 // undelegating
				expVested,                           // vested
				expUnvested,                         // unvested
				s.initialVesting.Add(s.gasCoins...), // initial balance
				sdk.NewCoins(),                      // sent amount
				coins,                               // received amount
				s.gasCoins,                          // consumed fee
			)

			validateVestingAccountBalancesWithValues(
				s.ctx,
				s.app,
				s.va,
				delegating,       // delegated vesting = 500 swtr
				sdk.NewCoins(),   // delegated free = 0 swtr
				delegating,       // locked = 500 swtr
				sdk.NewCoins(),   // vested = 0 swtr
				s.initialVesting, // unvested = 1000 swtr
				addCoins(subCoins(s.initialVesting, delegating), coins), // balances = 500 swtr + 300 swtr
				coins, // spendable balances = 300 swtr
			)
		})
		It("failed in sending coins", func() {
			mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)

			// Try to send 300 coins from monthly vesting account
			coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(300, 18)))
			err = s.app.BankKeeper.SendCoins(s.ctx, mva.GetAddress(), user, coins)
			Expect(err).ToNot(BeNil())
		})
	})
})
