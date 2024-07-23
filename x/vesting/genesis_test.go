package vesting_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "swisstronik/testutil/keeper"
	"swisstronik/testutil/nullify"
	"swisstronik/x/vesting"
	"swisstronik/x/vesting/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, _, _, ctx := keepertest.VestingKeeper(t)
	vesting.InitGenesis(ctx, *k, genesisState)
	got := vesting.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
