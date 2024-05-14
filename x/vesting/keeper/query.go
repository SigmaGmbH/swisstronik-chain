package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"swisstronik/x/vesting/types"
)

type Querier struct {
	Keeper
}

var _ types.QueryServer = Querier{}

func (q Querier) Balances(goCtx context.Context, req *types.QueryBalancesRequest) (*types.QueryBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	account, err := q.Keeper.GetMonthlyVestingAccount(ctx, address)
	if err != nil {
		return nil, err
	}

	blockTime := ctx.BlockTime()
	locked := account.LockedCoins(blockTime)
	unvested := account.GetVestingCoins(blockTime)
	vested := account.GetVestedCoins(blockTime)

	return &types.QueryBalancesResponse{
		Locked:   locked,
		Unvested: unvested,
		Vested:   vested,
	}, nil
}
