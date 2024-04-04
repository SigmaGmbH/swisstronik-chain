package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
	"testing"
)

func TestQueryParams(t *testing.T) {
	k, ctx := testkeeper.ComplianceKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)
	q := keeper.Querier{Keeper: *k}
	resp, err := q.Params(sdk.WrapSDKContext(ctx), &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, resp.Params, params)
}
