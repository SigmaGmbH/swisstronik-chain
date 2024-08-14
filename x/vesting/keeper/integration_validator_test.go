package keeper_test

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"swisstronik/tests"
	"swisstronik/testutil"
	"swisstronik/utils"
	"swisstronik/x/vesting/types"
)

func (suite *VestingTestSuite) bootstrapValidators(numVals int) ([]sdk.AccAddress, []sdk.ValAddress, []cryptotypes.PubKey, []stakingtypes.Validator) {
	var (
		addrDels = make([]sdk.AccAddress, numVals)
		addrVals = make([]sdk.ValAddress, numVals)
		pks      = simtestutil.CreateTestPubKeys(numVals)
		vs       = make([]stakingtypes.Validator, numVals)
	)
	for i := 0; i < numVals; i++ {
		addrDels[i] = tests.RandomAccAddress()
		addrVals[i] = addrDels[i].Bytes()

		vs[i], _ = stakingtypes.NewValidator(addrVals[i], pks[i], stakingtypes.Description{})
		vs[i] = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, vs[i], true)
		_ = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, vs[i].GetOperator())
		_ = suite.app.StakingKeeper.SetValidatorByConsAddr(suite.ctx, vs[i])
	}
	return addrDels, addrVals, pks, vs
}

func (suite *VestingTestSuite) initializeValidators(powers []int64) ([]sdk.AccAddress, []sdk.ValAddress, []cryptotypes.PubKey, []stakingtypes.Validator) {
	numVals := len(powers)
	var totalPower int64
	for _, power := range powers {
		totalPower += power
	}

	addrDels, addrVals, pks, vs := suite.bootstrapValidators(numVals)

	amount := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, totalPower)
	totalSupply := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, amount))
	notBondedPool := suite.app.StakingKeeper.GetNotBondedPool(suite.ctx)
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, notBondedPool)
	_ = testutil.FundModuleAccount(suite.ctx, suite.app.BankKeeper, notBondedPool.GetName(), totalSupply)

	var issuedShares = make([]sdk.Dec, numVals)
	for i, power := range powers {
		vs[i], _ = stakingtypes.NewValidator(addrVals[i], pks[i], stakingtypes.Description{})
		tokens := suite.app.StakingKeeper.TokensFromConsensusPower(suite.ctx, power)
		vs[i], issuedShares[i] = vs[i].AddTokensFromDel(tokens)
		vs[i] = stakingkeeper.TestingUpdateValidator(&suite.app.StakingKeeper, suite.ctx, vs[i], true)
		_ = suite.app.StakingKeeper.Hooks().AfterValidatorCreated(suite.ctx, vs[i].GetOperator())

		if !tokens.Equal(math.ZeroInt()) {
			// setup delegation if non-zero shares
			bond := stakingtypes.NewDelegation(addrDels[i], addrVals[i], issuedShares[i])
			suite.app.StakingKeeper.SetDelegation(suite.ctx, bond)
			suite.app.DistrKeeper.SetDelegatorStartingInfo(suite.ctx, vs[i].GetOperator(), addrDels[i], distrtypes.NewDelegatorStartingInfo(0, math.LegacyOneDec(), uint64(suite.ctx.BlockHeight())))
			// set historical reward record for new validator
			decCoins := sdk.NewDecCoins()
			historicalRewards := distrtypes.NewValidatorHistoricalRewards(decCoins, 2)
			suite.app.DistrKeeper.SetValidatorHistoricalRewards(suite.ctx, addrVals[i], 0, historicalRewards)
		}
	}

	_ = suite.Commit()

	return addrDels, addrVals, pks, vs
}

func (suite *VestingTestSuite) TestSetupTestWithBondedValidators() error {
	err := suite.SetupTest()
	if err != nil {
		return err
	}

	powers := []int64{0, 100, 400, 400, 200}
	_, addrVals, _, _ := suite.initializeValidators(powers)

	for i, address := range addrVals {
		val, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, address)
		Expect(ok).To(BeTrue())
		if i == 0 {
			Expect(val.Status).To(Equal(stakingtypes.Unbonded))
		} else {
			Expect(val.Status).To(Equal(stakingtypes.Bonded))
		}
		fmt.Printf("%d: status: %s, tokens: %s\n", i, val.Status.String(), val.Tokens.String())
	}

	return nil
}

func (suite *VestingTestSuite) TestSetupWithUnbondedValidators() error {
	err := suite.SetupTest()
	if err != nil {
		return err
	}

	_, addrVals, _, _ := suite.bootstrapValidators(5)
	_ = suite.Commit()

	for i, address := range addrVals {
		val, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, address)
		Expect(ok).To(BeTrue())
		Expect(val.Status).To(Equal(stakingtypes.Unbonded))
		fmt.Printf("%d: status: %s, tokens: %s\n", i, val.Status.String(), val.Tokens.String())
	}

	bootValidators := suite.app.StakingKeeper.GetValidators(suite.ctx, 20)
	for i, val := range bootValidators {
		fmt.Printf("%d: status: %s, tokens: %s\n", i, val.Status.String(), val.Tokens.String())
	}
	return nil
}

func (suite *VestingTestSuite) TestSetupWithJailedValidatorsBySelfUndelegation() error {
	err := suite.SetupTest()
	if err != nil {
		return err
	}

	powers := []int64{0, 100, 400, 400, 200}
	addrDels, addrVals, _, vs := suite.initializeValidators(powers)

	for i, val := range vs {
		fmt.Printf("%d: status: %s, tokens: %s\n", i, val.Status.String(), val.Tokens.String())

		if val.Status == stakingtypes.Bonded {
			// Make active validator jailed
			val.MinSelfDelegation = val.Tokens

			coin := sdk.NewCoin(utils.BaseDenom, val.Tokens)
			undelMsg := stakingtypes.NewMsgUndelegate(addrDels[i], addrVals[i], coin)
			msgServer := stakingkeeper.NewMsgServerImpl(&suite.app.StakingKeeper)
			_, err = msgServer.Undelegate(suite.ctx, undelMsg)
			Expect(err).To(BeNil())

			_ = suite.Commit()

			val2, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, addrVals[i])
			Expect(ok).To(BeTrue())
			Expect(val2.Jailed).To(BeTrue())

			fmt.Printf("> %d: status: %s, tokens: %s\n", i, val2.Status.String(), val2.Tokens.String())
		}
	}

	return nil
}

func (suite *VestingTestSuite) TestSetupWithJailedValidatorsByDowntime() error {
	err := suite.SetupTest()
	if err != nil {
		return err
	}

	powers := []int64{0, 100, 400, 400, 200}
	_, _, pks, _ := suite.initializeValidators(powers)

	for i, pk := range pks {
		consAddr := sdk.ConsAddress(pk.Address())
		val := suite.app.StakingKeeper.ValidatorByConsAddr(suite.ctx, consAddr)
		fmt.Printf("%d: status: %s, tokens: %s\n", i, val.GetStatus().String(), val.GetTokens().String())

		if val.GetStatus() == stakingtypes.Bonded {
			params := suite.app.SlashingKeeper.GetParams(suite.ctx)
			if i%2 == 0 {
				// set slash fraction double sign 5% for odd
				params.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.05") // 5%
				_ = suite.app.SlashingKeeper.SetParams(suite.ctx, params)
			} else {
				// set slash fraction double sign 10% for even
				params.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.1") // 1%
				_ = suite.app.SlashingKeeper.SetParams(suite.ctx, params)
			}

			// Pass blocks of signed blocks window
			signedBlockWindow := suite.app.SlashingKeeper.SignedBlocksWindow(suite.ctx)
			for j := int64(0); j <= signedBlockWindow; j++ {
				_ = suite.Commit()
			}

			signingInfo := slashingtypes.NewValidatorSigningInfo(
				consAddr,
				0,
				int64(0),
				time.Unix(0, 0),
				false,
				int64(signedBlockWindow), // should be over signedBlockWindow - minSignedBlockWindow
			)
			suite.app.SlashingKeeper.SetValidatorSigningInfo(suite.ctx, consAddr, signingInfo)
			suite.app.SlashingKeeper.HandleValidatorSignature(suite.ctx, consAddr.Bytes(), powers[i], false)

			val = suite.app.StakingKeeper.ValidatorByConsAddr(suite.ctx, consAddr)
			Expect(val).To(Not(BeNil()))
			Expect(val.IsJailed()).To(BeTrue())

			_ = suite.Commit()

			fmt.Printf("> %d: slash:%s status: %s, tokens: %s\n", i, params.SlashFractionDowntime.String(), val.GetStatus().String(), val.GetTokens().String())
		}
	}

	return nil
}

func (suite *VestingTestSuite) TestSetupWithTombstonedValidators() error {
	err := suite.SetupTest()
	if err != nil {
		return err
	}

	powers := []int64{0, 100, 400, 400, 200}
	_, _, pks, _ := suite.initializeValidators(powers)

	for i, pk := range pks {
		consAddr := sdk.ConsAddress(pk.Address())
		val := suite.app.StakingKeeper.ValidatorByConsAddr(suite.ctx, consAddr)
		fmt.Printf("%d: status: %s, tokens: %s\n", i, val.GetStatus().String(), val.GetTokens().String())

		if val.GetStatus() == stakingtypes.Bonded {
			params := suite.app.SlashingKeeper.GetParams(suite.ctx)
			if i%2 == 0 {
				// set slash fraction double sign 5% for odd
				params.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.05") // 5%
				_ = suite.app.SlashingKeeper.SetParams(suite.ctx, params)
			} else {
				// set slash fraction double sign 10% for even
				params.SlashFractionDoubleSign = sdk.MustNewDecFromStr("0.1") // 1%
				_ = suite.app.SlashingKeeper.SetParams(suite.ctx, params)
			}

			signingInfo := slashingtypes.NewValidatorSigningInfo(
				consAddr,
				suite.ctx.BlockHeight(),
				int64(0),
				time.Unix(0, 0),
				false,
				int64(0),
			)
			suite.app.SlashingKeeper.SetValidatorSigningInfo(suite.ctx, consAddr, signingInfo)

			// BeginBlocker in evidence keeper with misbehavior of duplicated vote
			tmEvidence := abci.Misbehavior{
				Type: abci.MisbehaviorType_DUPLICATE_VOTE,
				Validator: abci.Validator{
					Address: consAddr.Bytes(),
					Power:   powers[i],
				},
				Height:           suite.ctx.BlockHeight(),
				Time:             time.Now(),
				TotalVotingPower: powers[i],
			}
			evidence := evidencetypes.FromABCIEvidence(tmEvidence)
			suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence.(*evidencetypes.Equivocation))

			_ = suite.Commit()

			isTombstoned := suite.app.SlashingKeeper.IsTombstoned(suite.ctx, consAddr)
			Expect(isTombstoned).To(BeTrue())

			fmt.Printf("> tombstoned: %t\n", isTombstoned)

			val = suite.app.StakingKeeper.ValidatorByConsAddr(suite.ctx, consAddr)
			Expect(val).To(Not(BeNil()))
			Expect(val.IsJailed()).To(BeTrue())

			fmt.Printf("> %d: slash:%s status: %s, tokens: %s\n", i, params.SlashFractionDoubleSign.String(), val.GetStatus().String(), val.GetTokens().String())
		}
	}

	return nil
}

func (suite *VestingTestSuite) ExpectJailValidatorSuccess(addrVal sdk.ValAddress, slashFraction sdk.Dec) stakingtypes.ValidatorI {
	params := suite.app.SlashingKeeper.GetParams(suite.ctx)
	params.SlashFractionDowntime = slashFraction
	_ = suite.app.SlashingKeeper.SetParams(suite.ctx, params)

	validator, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, addrVal)
	Expect(ok).To(BeTrue())
	Expect(validator.Status).To(Equal(stakingtypes.Bonded))

	signedBlockWindow := suite.app.SlashingKeeper.SignedBlocksWindow(suite.ctx)
	for j := int64(0); j <= signedBlockWindow; j++ {
		_ = suite.Commit()
	}

	pk, _ := validator.ConsPubKey()
	consAddr := sdk.ConsAddress(pk.Address())

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		0,
		int64(0),
		time.Unix(0, 0),
		false,
		int64(signedBlockWindow), // should be over signedBlockWindow - minSignedBlockWindow
	)
	suite.app.SlashingKeeper.SetValidatorSigningInfo(suite.ctx, consAddr, signingInfo)
	power := validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	suite.app.SlashingKeeper.HandleValidatorSignature(suite.ctx, consAddr.Bytes(), power, false)

	_ = suite.Commit() // not necessary

	val := suite.app.StakingKeeper.ValidatorByConsAddr(suite.ctx, consAddr)
	Expect(val).To(Not(BeNil()))
	Expect(val.IsJailed()).To(BeTrue())

	return val
}

func (suite *VestingTestSuite) ExpectTombstoneValidtorSuccess(addrVal sdk.ValAddress, slashFraction sdk.Dec) stakingtypes.ValidatorI {
	params := suite.app.SlashingKeeper.GetParams(suite.ctx)
	params.SlashFractionDoubleSign = slashFraction
	_ = suite.app.SlashingKeeper.SetParams(suite.ctx, params)

	validator, ok := suite.app.StakingKeeper.GetValidator(suite.ctx, addrVal)
	Expect(ok).To(BeTrue())
	Expect(validator.Status).To(Equal(stakingtypes.Bonded))

	pk, _ := validator.ConsPubKey()
	consAddr := sdk.ConsAddress(pk.Address())

	signingInfo := slashingtypes.NewValidatorSigningInfo(
		consAddr,
		suite.ctx.BlockHeight(),
		int64(0),
		time.Unix(0, 0),
		false,
		int64(0),
	)
	suite.app.SlashingKeeper.SetValidatorSigningInfo(suite.ctx, consAddr, signingInfo)

	power := validator.GetConsensusPower(suite.app.StakingKeeper.PowerReduction(suite.ctx))
	// BeginBlocker in evidence keeper with misbehavior of duplicated vote
	tmEvidence := abci.Misbehavior{
		Type: abci.MisbehaviorType_DUPLICATE_VOTE,
		Validator: abci.Validator{
			Address: consAddr.Bytes(),
			Power:   power,
		},
		Height:           suite.ctx.BlockHeight(),
		Time:             time.Now(),
		TotalVotingPower: power,
	}
	evidence := evidencetypes.FromABCIEvidence(tmEvidence)
	suite.app.EvidenceKeeper.HandleEquivocationEvidence(suite.ctx, evidence.(*evidencetypes.Equivocation))

	_ = suite.Commit() // not necessary

	isTombstoned := suite.app.SlashingKeeper.IsTombstoned(suite.ctx, consAddr)
	Expect(isTombstoned).To(BeTrue())

	val := suite.app.StakingKeeper.ValidatorByConsAddr(suite.ctx, consAddr)
	Expect(val).To(Not(BeNil()))
	Expect(val.IsJailed()).To(BeTrue())

	return val
}

var _ = Describe("Additional tests with multiple validators for Monthly Vesting Account", func() {
	var (
		s *VestingTestSuite

		err error

		// Different types of validators
		powers []int64
		//addrDels []sdk.AccAddress
		addrVals []sdk.ValAddress
		//pks      []cryptotypes.PubKey
	)

	BeforeEach(func() {
		s = new(VestingTestSuite)
		err = s.SetupTest()
		Expect(err).To(BeNil())

		var from common.Address
		from, s.vaPrivKey = tests.RandomEthAddressWithPrivateKey()
		s.va = sdk.AccAddress(from.Bytes())
		s.funder = tests.RandomAccAddress()

		s.initialVesting = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1000, 18)))
		s.gasCoins = sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewInt(1e6)))

		// Fund coins to funder to create monthly vesting account
		s.ExpectFundCoins(s.funder, s.initialVesting)

		// Setup multiple validators
		powers = []int64{0, 50, 100, 0, 200}
		//addrDels, addrVals, pks, _ = s.initializeValidators(powers)
		_, addrVals, _, _ = s.initializeValidators(powers)
	})

	Context("Create vesting account with cliff 0 day, 2 month, amount 1000", func() {
		BeforeEach(func() {
			// Create monthly vesting account
			resp, err := s.msgServer.HandleCreateMonthlyVestingAccount(s.goCtx, &types.MsgCreateMonthlyVestingAccount{
				FromAddress: s.funder.String(),
				ToAddress:   s.va.String(),
				CliffDays:   0,
				Months:      2,
				Amount:      s.initialVesting,
			})
			Expect(resp).To(Equal(&types.MsgCreateMonthlyVestingAccountResponse{}))
			err = s.Commit()
			Expect(err).To(BeNil())

			// Query monthly vesting account
			mva, err := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
			Expect(err).To(BeNil())
			Expect(mva).ToNot(BeNil())
		})
		It("slashing-tombstoned validator", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 1000 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1000, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1500 SWTR during vesting time
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unlock 50 % of vesting
					duration := time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// “Slashing” during vesting period for tombstoned validator - 10%
					valI := s.ExpectTombstoneValidtorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Expected result: staked amount = (delegated vesting + free vesting) - 10%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expAmount := mva.DelegatedVesting.Add(mva.DelegatedFree...).AmountOf(utils.BaseDenom)
					expAmount = expAmount.Sub(expAmount.MulRaw(10).QuoRaw(100))
					Expect(valI.GetTokens()).To(Equal(expAmount))

					break
				}
			}

		})
		It("slashing-jailed validator", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Stake 1000 SWTR during vesting time
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := s.initialVesting
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unlock 50 % of vesting
					duration := time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// Unlock 100% of vesting
					duration = time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// “Slashing” for jailed validator - 5%
					valI := s.ExpectJailValidatorSuccess(address, sdk.MustNewDecFromStr("0.05"))
					Expect(valI).To(Not(BeNil()))

					// Expected result: delegation vesting 950
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expAmount := mva.DelegatedVesting.Add(mva.DelegatedFree...).AmountOf(utils.BaseDenom)
					expAmount = expAmount.Sub(expAmount.MulRaw(5).QuoRaw(100))
					Expect(valI.GetTokens()).To(Equal(expAmount))
					expAmount = math.NewIntWithDecimal(950, 18)
					Expect(valI.GetTokens()).To(Equal(expAmount))

					break
				}
			}
		})
		It("slashing-tombstoned validator-2", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Stake 1000 SWTR during vesting time
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := s.initialVesting
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unlock 50 % of vesting
					duration := time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// Unlock 100% of vesting
					duration = time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// “Slashing” after vesting time for tombtoned validator - 10%
					valI := s.ExpectTombstoneValidtorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Expected result: delegation vesting 900
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expAmount := mva.DelegatedVesting.Add(mva.DelegatedFree...).AmountOf(utils.BaseDenom)
					expAmount = expAmount.Sub(expAmount.MulRaw(10).QuoRaw(100))
					Expect(valI.GetTokens()).To(Equal(expAmount))

					break
				}
			}
		})
		It("slashing-jailed validator-2", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 1000 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1000, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1500 SWTR during vesting time
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unlock 50 % of vesting
					duration := time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// “Slashing” during vesting period for jailed validator - 5%
					valI := s.ExpectJailValidatorSuccess(address, sdk.MustNewDecFromStr("0.05"))
					Expect(valI).To(Not(BeNil()))

					// Unlock 100% of vesting
					duration = time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// Expected result:staked amount = (delegated vesting + free vesting) - 5%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expAmount := mva.DelegatedVesting.Add(mva.DelegatedFree...).AmountOf(utils.BaseDenom)
					expAmount = expAmount.Sub(expAmount.MulRaw(5).QuoRaw(100))
					Expect(valI.GetTokens()).To(Equal(expAmount))
					expAmount = math.NewIntWithDecimal(1425, 18)
					Expect(valI.GetTokens()).To(Equal(expAmount))

					break
				}
			}

		})
		It("slashing-tombstoned validator-3", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 1000 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1000, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1500 SWTR during vesting time
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unlock 50 % of vesting
					duration := time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// Unlock 100% of vesting
					duration = time.Duration(types.SecondsOfMonth) * time.Second
					err = s.CommitAfter(duration)
					Expect(err).To(BeNil())

					// “Slashing” during vesting period for tombstoned validator - 10%
					valI := s.ExpectTombstoneValidtorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Expected result: staked amount = (delegated vesting + free vesting) - 10%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expAmount := mva.DelegatedVesting.Add(mva.DelegatedFree...).AmountOf(utils.BaseDenom)
					expAmount = expAmount.Sub(expAmount.MulRaw(10).QuoRaw(100))

					Expect(valI.GetTokens()).To(Equal(expAmount))
					break
				}
			}
		})
	})
	Context("Create vesting account with cliff 50 days, 12 months, amount 1000 swtr", func() {
		BeforeEach(func() {
			// Create monthly vesting account
			resp, err := s.msgServer.HandleCreateMonthlyVestingAccount(s.goCtx, &types.MsgCreateMonthlyVestingAccount{
				FromAddress: s.funder.String(),
				ToAddress:   s.va.String(),
				CliffDays:   50,
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
		})
		It("active validator", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Bonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1200 SWTR
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1200, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unstake during cliff 1200 SWTR (Active validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 0, all balance 1500
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(BeNil())
					expBalance := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1500, 18)))
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance).To(Equal(expBalance))

					break
				}
			}
		})
		It("jailed validator", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1200 SWTR
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1200, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Make validator jailed
					valI := s.ExpectJailValidatorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Unstake during cliff 1200 SWTR (Jailed validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating.MulInt(sdk.NewInt(90)).QuoInt(sdk.NewInt(100))
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 120(=10%), all balance (1500-1200) + 1200 * 90%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(Equal(delegating.QuoInt(sdk.NewInt(10))))
					expBalance := math.NewIntWithDecimal(1380, 18)
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance.AmountOf(utils.BaseDenom)).To(Equal(expBalance))

					break
				}
			}
		})
		It("unbonded validator", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1200 SWTR
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1200, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unstake during cliff time 1200 SWTR  (Unbonded validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 0, all balance 1500
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(BeNil())
					expBalance := math.NewIntWithDecimal(1500, 18)
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance.AmountOf(utils.BaseDenom)).To(Equal(expBalance))

					break
				}
			}

		})
		It("tombstoned validator", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 1200 SWTR
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1200, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Make validator tombstoned
					valI := s.ExpectTombstoneValidtorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Unstake during cliff time 1200*90% SWTR  (Tombstoned validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating.MulInt(sdk.NewInt(90)).QuoInt(sdk.NewInt(100))
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 120(=10%), all balance (1500-1200) + 1200 * 90%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(Equal(delegating.QuoInt(sdk.NewInt(10))))
					expBalance := math.NewIntWithDecimal(1380, 18)
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance.AmountOf(utils.BaseDenom)).To(Equal(expBalance))

					break
				}
			}
		})
		It("active validator-2", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Bonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 500 SWTR (vesting amount)
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unstake during cliff 500 SWTR (Active validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 0 , all balance 1500
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(BeNil())
					expBalance := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(1500, 18)))
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance).To(Equal(expBalance))

					break
				}
			}
		})
		It("jailed validator-2", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 500 SWTR (vesting amount)
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Make validator jailed
					valI := s.ExpectJailValidatorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Unstake during cliff 500 SWTR (Jailed validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating.MulInt(sdk.NewInt(90)).QuoInt(sdk.NewInt(100))
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 50(=10%), all balance 1000+500*90%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(Equal(delegating.QuoInt(sdk.NewInt(10))))
					expBalance := math.NewIntWithDecimal(1450, 18)
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance.AmountOf(utils.BaseDenom)).To(Equal(expBalance))

					break
				}
			}
		})
		It("unbonded validator-2", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 500 SWTR
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Unstake during cliff time 500 SWTR  (Unbonded validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 0 , all balance 1500
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(BeNil())
					expBalance := math.NewIntWithDecimal(1500, 18)
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance.AmountOf(utils.BaseDenom)).To(Equal(expBalance))

					break
				}
			}
		})
		It("tombstoned validator-2", func() {
			for _, address := range addrVals {
				val, ok := s.app.StakingKeeper.GetValidator(s.ctx, address)
				Expect(ok).To(BeTrue())

				if val.GetStatus() == stakingtypes.Unbonded {
					// Receive 500 SWTR
					coins := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectFundCoins(s.va, coins)

					// Stake 500 SWTR
					s.ExpectFundCoins(s.va, s.gasCoins)
					delegating := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					s.ExpectDelegateSuccess(s.vaPrivKey, delegating[0], val)

					// Make validator tombstoned
					valI := s.ExpectTombstoneValidtorSuccess(address, sdk.MustNewDecFromStr("0.1"))
					Expect(valI).To(Not(BeNil()))

					// Unstake  during cliff time 500 SWTR  (Tombstoned validator)
					s.ExpectFundCoins(s.va, s.gasCoins)
					undelegating := delegating.MulInt(sdk.NewInt(90)).QuoInt(sdk.NewInt(100))
					s.ExpectUndelegateSuccess(s.vaPrivKey, undelegating[0], val)

					// Expected result: spendable balance 500, delegated vesting 50(=10%), all balance 1000+500*90%
					mva, _ := s.querier.GetMonthlyVestingAccount(s.ctx, s.va)
					expSpendable := sdk.NewCoins(sdk.NewCoin(utils.BaseDenom, math.NewIntWithDecimal(500, 18)))
					spendable := s.app.BankKeeper.SpendableCoins(s.ctx, mva.GetAddress())
					Expect(spendable).To(Equal(expSpendable))
					Expect(mva.DelegatedVesting).To(Equal(delegating.QuoInt(sdk.NewInt(10))))
					expBalance := math.NewIntWithDecimal(1450, 18)
					balance := s.app.BankKeeper.GetAllBalances(s.ctx, mva.GetAddress())
					Expect(balance.AmountOf(utils.BaseDenom)).To(Equal(expBalance))

					break
				}
			}
		})
	})
})
