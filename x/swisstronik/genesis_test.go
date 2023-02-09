package swisstronik_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	keepertest "swisstronik/testutil/keeper"
	"swisstronik/testutil/nullify"
	"swisstronik/x/swisstronik"
	"swisstronik/x/swisstronik/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.SwisstronikKeeper(t)
	swisstronik.InitGenesis(ctx, *k, genesisState)
	got := swisstronik.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
