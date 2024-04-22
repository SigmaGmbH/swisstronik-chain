package testutil

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"swisstronik/app"
)

func SubmitProposal(
	ctx sdk.Context,
	app *app.App,
	priv cryptotypes.PrivKey,
	content govv1beta1.Content,
) (event *abci.Event, err error) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())
	stakeDenom := stakingtypes.DefaultParams().BondDenom

	deposit := sdk.NewCoins(sdk.NewCoin(stakeDenom, math.NewInt(100000000)))
	msg, err := govv1beta1.NewMsgSubmitProposal(content, deposit, accountAddress)
	if err != nil {
		return nil, err
	}
	res, err := DeliverTx(ctx, app, priv, nil, msg)
	if err != nil {
		return nil, err
	}

	events := res.GetEvents()
	for _, event := range events {
		if event.Type == "submit_proposal" && event.Attributes[0].Key == "proposal_id" {
			return &event, nil
		}
	}
	return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "SubmitProposal failed")
}

// Delegate delivers a delegate tx
func Delegate(
	ctx sdk.Context,
	app *app.App,
	priv cryptotypes.PrivKey,
	delegateAmount sdk.Coin,
	validator stakingtypes.Validator,
) (abci.ResponseDeliverTx, error) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	val, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	delegateMsg := stakingtypes.NewMsgDelegate(accountAddress, val, delegateAmount)
	return DeliverTx(ctx, app, priv, nil, delegateMsg)
}

// Vote delivers a vote tx with the VoteOption "yes" or "no"
func Vote(
	ctx sdk.Context,
	app *app.App,
	priv cryptotypes.PrivKey,
	proposalID uint64,
	voteOption govv1beta1.VoteOption,
) (abci.ResponseDeliverTx, error) {
	accountAddress := sdk.AccAddress(priv.PubKey().Address().Bytes())

	voteMsg := govv1beta1.NewMsgVote(accountAddress, proposalID, voteOption)
	return DeliverTx(ctx, app, priv, nil, voteMsg)
}
