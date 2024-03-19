package keeper

import (
	"context"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/compliance/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) HandleSetIssuerDetails(goCtx context.Context, msg *types.MsgSetIssuerDetails) (*types.MsgSetIssuerDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Verify operator address
	operatorAddress, err := sdk.AccAddressFromBech32(msg.Details.Operator)
	if err != nil {
		return nil, err
	}

	signerAddress, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	if !operatorAddress.Equals(signerAddress) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "operator and signer address mismatch %s, %s", operatorAddress.String(), signerAddress.String())
	}

	// Check if there is no such issuer
	issuerAddress, err := sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return nil, err
	}

	issuerExists, err := k.IssuerExists(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}
	if issuerExists {
		return nil, errors.Wrap(types.ErrInvalidParam, "issuer already exists")
	}

	if err := k.SetIssuerDetails(ctx, issuerAddress, msg.Details); err != nil {
		return nil, err
	}

	return &types.MsgSetIssuerDetailsResponse{}, nil
}

func (k msgServer) HandleUpdateIssuerDetails(goCtx context.Context, msg *types.MsgUpdateIssuerDetails) (*types.MsgUpdateIssuerDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if issuer exists
	issuerAddress, err := sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return nil, err
	}

	issuerExists, err := k.IssuerExists(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}
	if !issuerExists {
		return nil, errors.Wrap(types.ErrInvalidParam, "issuer not found")
	}

	// Check if signer is previous issuer operator
	issuerDetails, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}

	signerAddress, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	prevOperatorAddress, err := sdk.AccAddressFromBech32(issuerDetails.Operator)
	if err != nil {
		return nil, err
	}

	if !signerAddress.Equals(prevOperatorAddress) {
		return nil, errors.Wrap(types.ErrInvalidSignature, "invalid signer")
	}

	// Revoke verification if address was verified
	if err := k.SetAddressVerificationStatus(ctx, issuerAddress, false); err != nil {
		return nil, err
	}

	if err := k.SetIssuerDetails(ctx, issuerAddress, msg.Details); err != nil {
		return nil, err
	}
	return &types.MsgUpdateIssuerDetailsResponse{}, nil
}

func (k msgServer) HandleRemoveIssuer(goCtx context.Context, msg *types.MsgRemoveIssuer) (*types.MsgRemoveIssuerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if issuer exists
	issuerAddress, err := sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return nil, err
	}

	issuerExists, err := k.IssuerExists(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}
	if !issuerExists {
		return nil, errors.Wrap(types.ErrInvalidParam, "issuer not found")
	}

	// Check if there is a correct operator
	issuerDetails, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}
	msgSignerAddr, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}
	issuerOperatorAddr, err := sdk.AccAddressFromBech32(issuerDetails.Operator)
	if err != nil {
		return nil, err
	}
	if !issuerOperatorAddr.Equals(msgSignerAddr) {
		return nil, errors.Wrap(types.ErrInvalidParam, "operator and signer address mismatch")
	}

	k.RemoveIssuer(ctx, issuerAddress)
	return &types.MsgRemoveIssuerResponse{}, nil
}
