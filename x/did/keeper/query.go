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

	didDoc, err := k.GetLatestDIDDocument(ctx, req.Id)
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

	versions, err := k.GetAllDIDDocumentVersions(ctx, req.Id)
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

	didDoc, err := k.GetDIDDocumentVersion(ctx, req.Id, req.Version)
	if err != nil {
		return nil, err
	}

	return &types.QueryDIDDocumentVersionResponse{Value: &didDoc}, nil
}

func (k Keeper) Resource(
	goCtx context.Context, 
	req *types.QueryResourceRequest,
) (*types.QueryResourceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	req.Normalize()

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate corresponding DID Document exists
	did := types.JoinDID(types.DIDMethod, req.CollectionId)
	if !k.HasDIDDocument(ctx, did) {
		return nil, types.ErrDIDDocumentNotFound.Wrap(did)
	}

	resource, err := k.GetResource(ctx, req.CollectionId, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryResourceResponse{
		Resource: &resource,
	}, nil
}

func (k Keeper)	ResourceMetadata(
	goCtx context.Context, 
	req *types.QueryResourceMetadataRequest,
) (*types.QueryResourceMetadataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	req.Normalize()

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Validate corresponding DID Document exists
	did := types.JoinDID(types.DIDMethod, req.CollectionId)
	if !k.HasDIDDocument(ctx, did) {
		return nil, types.ErrDIDDocumentNotFound.Wrap(did)
	}

	metadata, err := k.GetResourceMetadata(ctx, req.CollectionId, req.Id)
	if err != nil {
		return nil, err
	}

	return &types.QueryResourceMetadataResponse{
		Resource: &metadata,
	}, nil
}

func (k Keeper) CollectionResources(
	goCtx context.Context, 
	req *types.QueryCollectionResourcesRequest,
) (*types.QueryCollectionResourcesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	req.Normalize()

	// Validate corresponding DID Document exists
	did := types.JoinDID(types.DIDMethod, req.CollectionId)
	if !k.HasDIDDocument(ctx, did) {
		return nil, types.ErrDIDDocumentNotFound.Wrap(did)
	}

	// Get all resources
	resources := k.GetResourceCollection(ctx, req.CollectionId)

	return &types.QueryCollectionResourcesResponse{
		Resources: resources,
	}, nil
}
