package keeper

import (
	"context"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"swisstronik/x/compliance/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) VerificationData(goCtx context.Context, req *types.QueryVerificationDataRequest) (*types.QueryVerificationDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	cfg := sdk.GetConfig()

	// Convert address in bytes format. Used to accept both bech32 and hex addresses
	var userAddress []byte
	switch {
	case common.IsHexAddress(req.Address):
		userAddress = common.HexToAddress(req.Address).Bytes()
	case strings.HasPrefix(req.Address, cfg.GetBech32AccountAddrPrefix()):
		userAddress, _ = sdk.AccAddressFromBech32(req.Address)
	default:
		return nil, fmt.Errorf("expected a valid hex or bech32 address (acc prefix %s), got '%s'", cfg.GetBech32AccountAddrPrefix(), req.Address)
	}

	verificationData, err := k.GetAddressInfo(ctx, common.BytesToAddress(userAddress))
	if err != nil {
		return nil, err
	}

	return &types.QueryVerificationDataResponse{Data: verificationData}, nil
}
