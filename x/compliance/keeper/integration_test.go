package keeper_test

import (
	"strconv"

	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"swisstronik/tests"
	"swisstronik/testutil"
	"swisstronik/utils"
	"swisstronik/x/compliance/types"
)

var _ = Describe("VerifyIssuer", Ordered, func() {
	issuerCreator := tests.RandomAccAddress()

	from, validIssuerPrivKey := tests.RandomEthAddressWithPrivateKey()
	validIssuer := sdk.AccAddress(from.Bytes())

	from, invalidIssuerPrivKey := tests.RandomEthAddressWithPrivateKey()
	invalidIssuer := sdk.AccAddress(from.Bytes())

	from, accountPrivKey := tests.RandomEthAddressWithPrivateKey()
	account := sdk.AccAddress(from.Bytes())

	from, emptyPrivKey := tests.RandomEthAddressWithPrivateKey()
	empty := sdk.AccAddress(from.Bytes())

	BeforeEach(func() {
		govParams := s.app.GovKeeper.GetParams(s.ctx)
		govParams.Quorum = "0.0000000001"
		err := s.app.GovKeeper.SetParams(s.ctx, govParams)
		Expect(err).To(BeNil())
	})

	Describe("Submitting a verifying issuer proposal through governance", func() {
		BeforeEach(func() {
			// Mint coins to pay gas fee, gov deposit and registering coins in Bank keeper
			amount, ok := sdk.NewIntFromString("10000000000000000000")
			s.Require().True(ok)
			coins := sdk.NewCoins(
				sdk.NewCoin(utils.BaseDenom, amount),
				sdk.NewCoin(stakingtypes.DefaultParams().BondDenom, amount),
			)
			err := testutil.FundAccount(s.ctx, s.app.BankKeeper, validIssuer, coins)
			s.Require().NoError(err)
			s.Commit()

			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, invalidIssuer, coins)
			s.Require().NoError(err)
			s.Commit()

			err = testutil.FundAccount(s.ctx, s.app.BankKeeper, account, coins)
			s.Require().NoError(err)
			s.Commit()
		})

		Describe("valid issuer has been verified", func() {
			BeforeEach(func() {
				// Set issuer details(not verified)
				issuerDetails := &types.IssuerDetails{Name: "test issuer"}
				_ = s.keeper.SetIssuerDetails(s.ctx, issuerCreator, validIssuer, issuerDetails)

				// Submit proposal with sufficient deposit
				content := types.NewVerifyIssuerProposal("test title", "test description", validIssuer.String())
				event, err := testutil.SubmitProposal(s.ctx, s.app, validIssuerPrivKey, content)
				s.Require().NoError(err)

				proposalID, err := strconv.ParseUint(event.Attributes[0].Value, 10, 64)
				s.Require().NoError(err)

				// Make sure proposal has been created
				proposal, found := s.app.GovKeeper.GetProposal(s.ctx, proposalID)
				s.Require().True(found)

				_, err = testutil.Delegate(s.ctx, s.app, accountPrivKey, sdk.NewCoin(utils.BaseDenom, math.NewInt(500000000000000000)), s.validator)
				s.Require().NoError(err)

				_, err = testutil.Vote(s.ctx, s.app, accountPrivKey, proposalID, govv1beta1.OptionYes)
				s.Require().NoError(err)

				// Make proposal pass in EndBlocker
				duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
				s.CommitAfter(duration)
				s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
				s.Commit()
			})
			It("Issuer should be verified", func() {
				verified, err := s.keeper.IsAddressVerified(s.ctx, validIssuer)
				s.Require().NoError(err)
				s.Require().True(verified)
			})
		})

		Describe("invalid issuer has not been verified", func() {
			BeforeEach(func() {
				// Set issuer details(not verified)
				issuerDetails := &types.IssuerDetails{Name: "test issuer"}
				_ = s.keeper.SetIssuerDetails(s.ctx, issuerCreator, invalidIssuer, issuerDetails)

				// Submit proposal with sufficient deposit
				content := types.NewVerifyIssuerProposal("test title", "test description", invalidIssuer.String())
				event, err := testutil.SubmitProposal(s.ctx, s.app, invalidIssuerPrivKey, content)
				s.Require().NoError(err)

				proposalID, err := strconv.ParseUint(event.Attributes[0].Value, 10, 64)
				s.Require().NoError(err)

				proposal, found := s.app.GovKeeper.GetProposal(s.ctx, proposalID)
				s.Require().True(found)

				// There's no delegate, vote to make proposal failed

				// Make proposal pass in EndBlocker
				duration := proposal.VotingEndTime.Sub(s.ctx.BlockTime()) + 1
				s.CommitAfter(duration)
				s.app.EndBlocker(s.ctx, abci.RequestEndBlock{Height: s.ctx.BlockHeight()})
				s.Commit()
			})
			It("Issuer should not be verified", func() {
				verified, err := s.keeper.IsAddressVerified(s.ctx, invalidIssuer)
				s.Require().NoError(err)
				s.Require().False(verified)
			})
		})

		Describe("should not create a proposal for verified issuer", func() {
			BeforeEach(func() {
				// Set issuer details(verified)
				issuerDetails := &types.IssuerDetails{Name: "test issuer"}
				_ = s.keeper.SetIssuerDetails(s.ctx, issuerCreator, validIssuer, issuerDetails)
				_ = s.keeper.SetAddressVerificationStatus(s.ctx, validIssuer, true)
			})
			It("should fail in submitting proposal", func() {
				// Submit proposal with sufficient deposit
				content := types.NewVerifyIssuerProposal("test title", "test description", validIssuer.String())
				_, err := testutil.SubmitProposal(s.ctx, s.app, validIssuerPrivKey, content)
				s.Require().Error(err)
			})
		})

		Describe("should not create a proposal for empty issuer that doesn't exist", func() {
			It("should fail in submitting proposal", func() {
				content := types.NewVerifyIssuerProposal("test title", "test description", empty.String())
				_, err := testutil.SubmitProposal(s.ctx, s.app, emptyPrivKey, content)
				s.Require().Error(err)
			})
		})
	})
})
