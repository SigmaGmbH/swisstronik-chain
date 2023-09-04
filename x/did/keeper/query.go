package keeper

import (
	"context"

	"swisstronik/x/did/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) DIDDocument(
	goCtx context.Context, 
	req *types.QueryDIDDocumentRequest,
) (*types.QueryDIDDocumentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	req.Id = types.NormalizeDID(req.Id)

	ctx := sdk.UnwrapSDKContext(goCtx)

	didDoc, err := k.GetLatestDIDDocument(&ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryDIDDocumentResponse{Value: &didDoc}, nil
}

func (k Keeper) AllDIDDocumentVersionsMetadata(
	goCtx context.Context, 
	req *types.QueryAllDIDDocumentVersionsMetadataRequest,
) (*types.QueryAllDIDDocumentVersionsMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	req.Id = types.NormalizeDID(req.Id)

	ctx := sdk.UnwrapSDKContext(goCtx)

	versions, err := k.GetAllDIDDocumentVersions(&ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryAllDIDDocumentVersionsMetadataResponse{Versions: versions}, nil
}

func (k Keeper) DIDDocumentVersion(
	goCtx context.Context,
	req *types.QueryDIDDocumentVersionRequest,
) (*types.QueryDIDDocumentVersionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	req.Id = types.NormalizeDID(req.Id)

	ctx := sdk.UnwrapSDKContext(goCtx)

	didDoc, err := k.GetDIDDocumentVersion(&ctx, req.Id, req.Version)
	if err != nil {
		return nil, err
	}

	return &types.QueryDIDDocumentVersionResponse{Value: &didDoc}, nil
}
