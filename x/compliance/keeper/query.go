package keeper

import (
	"context"
	"encoding/base64"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"swisstronik/x/compliance/types"
)

// Querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over keeper
type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (k Querier) OperatorDetails(goCtx context.Context, req *types.QueryOperatorDetailsRequest) (*types.QueryOperatorDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.OperatorAddress)
	if err != nil {
		return nil, err
	}

	details, err := k.GetOperatorDetails(ctx, address)
	if err != nil {
		return nil, err
	}

	return &types.QueryOperatorDetailsResponse{Details: details}, nil
}

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

func (k Querier) AddressesDetails(goCtx context.Context, req *types.QueryAddressesDetailsRequest) (*types.QueryAddressesDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var addresses []types.QueryAddressesDetailsResponse_AddressDetailsWithKey
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixAddressDetails)

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		var addressDetails types.AddressDetails
		if err := proto.Unmarshal(value, &addressDetails); err != nil {
			return err
		}
		addresses = append(addresses, types.QueryAddressesDetailsResponse_AddressDetailsWithKey{
			Address:        types.AccAddressFromKey(key).String(),
			AddressDetails: &addressDetails,
		})
		// NOTE, DO NOT FILTER VERIFICATIONS BY ISSUERS' EXISTENCE BECAUSE OF QUERY PERFORMANCE
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAddressesDetailsResponse{
		Addresses:  addresses,
		Pagination: pageRes,
	}, nil
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

func (k Querier) IssuersDetails(goCtx context.Context, req *types.QueryIssuersDetailsRequest) (*types.QueryIssuersDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var issuers []types.QueryIssuersDetailsResponse_IssuerDetailsWithKey
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		var issuerDetails types.IssuerDetails
		if err := proto.Unmarshal(value, &issuerDetails); err != nil {
			return err
		}
		issuers = append(issuers, types.QueryIssuersDetailsResponse_IssuerDetailsWithKey{
			IssuerAddress: types.AccAddressFromKey(key).String(),
			IssuerDetails: &issuerDetails,
		})
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryIssuersDetailsResponse{
		Issuers:    issuers,
		Pagination: pageRes,
	}, nil
}

func (k Querier) VerificationDetails(goCtx context.Context, req *types.QueryVerificationDetailsRequest) (*types.QueryVerificationDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	id, err := base64.StdEncoding.DecodeString(req.VerificationID)
	if err != nil {
		return nil, err
	}
	details, err := k.GetVerificationDetails(ctx, id)
	if err != nil {
		return nil, err
	}

	if details == nil {
		return &types.QueryVerificationDetailsResponse{}, nil
	}

	return &types.QueryVerificationDetailsResponse{Details: details}, nil
}

func (k Querier) VerificationsDetails(goCtx context.Context, req *types.QueryVerificationsDetailsRequest) (*types.QueryVerificationsDetailsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var verifications []types.QueryVerificationsDetailsResponse_VerificationDetailsWithKey
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)

	pageRes, err := query.Paginate(store, req.Pagination, func(key []byte, value []byte) error {
		var verificationDetails types.VerificationDetails
		if err := proto.Unmarshal(value, &verificationDetails); err != nil {
			return err
		}
		verifications = append(verifications, types.QueryVerificationsDetailsResponse_VerificationDetailsWithKey{
			VerificationID:      types.VerificationIdFromKey(key),
			VerificationDetails: &verificationDetails,
		})
		// NOTE, DO NOT FILTER VERIFICATIONS BY ISSUERS' EXISTENCE BECAUSE OF QUERY PERFORMANCE
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryVerificationsDetailsResponse{
		Verifications: verifications,
		Pagination:    pageRes,
	}, nil
}
