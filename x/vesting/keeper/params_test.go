package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/vesting/types"
)

func TestGetParams(t *testing.T) {
	k, _, _, ctx := testkeeper.VestingKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
