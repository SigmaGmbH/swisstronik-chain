package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"swisstronik/x/compliance/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) VerificationData(goCtx context.Context, req *types.QueryVerificationDataRequest) (*types.QueryVerificationDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	verificationData, err := k.GetAddressInfo(ctx, address)
	if err != nil {
		return nil, err
	}

	return &types.QueryVerificationDataResponse{Data: verificationData}, nil
}
