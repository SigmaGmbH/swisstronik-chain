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

func (k msgServer) SetAddressInfo(goCtx context.Context, msg *types.MsgSetAddressInfo) (*types.MsgSetAddressInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(msg.UserAddress)
	if err != nil {
		return nil, err
	}

	if err := k.SetAddressInfoRaw(ctx, address, msg.Data); err != nil {
		return nil, err
	}

	return &types.MsgSetAddressInfoResponse{}, nil
}

func (k msgServer) SetIssuerDetails(goCtx context.Context, msg *types.MsgSetIssuerDetails) (*types.MsgSetIssuerDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	issuerAddress, err := sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return nil, err
	}

	if err := k.SetIssuerDetailsInner(ctx, issuerAddress, msg.Details.IssuerAlias); err != nil {
		return nil, err
	}

	return &types.MsgSetIssuerDetailsResponse{}, nil
}
