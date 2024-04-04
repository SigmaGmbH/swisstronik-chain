package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"swisstronik/x/compliance/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (k Querier) AddressDetails(goCtx context.Context, req *types.QueryAddressDetailsRequest) (*types.QueryAddressDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	details, err := k.GetAddressDetails(ctx, address)
	if err != nil {
		return nil, err
	}

	return &types.QueryAddressDetailsResponse{Data: details}, nil
}

func (k Querier) IssuerDetails(goCtx context.Context, req *types.QueryIssuerDetailsRequest) (*types.QueryIssuerDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	issuerAddress, err := sdk.AccAddressFromBech32(req.IssuerAddress)
	if err != nil {
		return nil, err
	}

	issuerDetails, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return nil, err
	}

	return &types.QueryIssuerDetailsResponse{Details: issuerDetails}, nil
}

func (k Querier) VerificationDetails(goCtx context.Context, req *types.QueryVerificationDetailsRequest) (*types.QueryVerificationDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	details, err := k.GetVerificationDetails(ctx, []byte(req.VerificationID))
	if err != nil {
		return nil, err
	}

	if details == nil {
		return &types.QueryVerificationDetailsResponse{}, nil
	}

	return &types.QueryVerificationDetailsResponse{Details: details}, nil
}
