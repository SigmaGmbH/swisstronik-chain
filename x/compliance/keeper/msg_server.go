package keeper

import (
	"context"
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

	// TODO: Double check operator address

	issuerAddress, err := sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return nil, err
	}

	if err := k.SetIssuerDetails(ctx, issuerAddress, msg.Details); err != nil {
		return nil, err
	}

	return &types.MsgSetIssuerDetailsResponse{}, nil
}

func (k msgServer) HandleUpdateIssuerDetails(goCtx context.Context, msg *types.MsgUpdateIssuerDetails) (*types.MsgUpdateIssuerDetailsResponse, error) {
	// TODO: check if issuer exists
	// TODO: check if operator is correct
	// TODO: if issuer is verified, revoke verification
	return &types.MsgUpdateIssuerDetailsResponse{}, nil
}

func (k msgServer) HandleRemoveIssuer(goCtx context.Context, msg *types.MsgRemoveIssuer) (*types.MsgRemoveIssuerResponse, error) {
	// TODO: check if issuer exists
	// TODO: check if operator is correct
	// TODO: remove issuer
	return &types.MsgRemoveIssuerResponse{}, nil
}
