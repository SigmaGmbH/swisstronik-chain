package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	testkeeper "swisstronik/testutil/keeper"
	"swisstronik/x/compliance/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ComplianceKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)
	require.EqualValues(t, params, k.GetParams(ctx))
}
