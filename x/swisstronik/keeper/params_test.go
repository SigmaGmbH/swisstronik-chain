package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/swisstronik/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.SwisstronikKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
