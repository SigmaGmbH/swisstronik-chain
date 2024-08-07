package simulation_test

import (
	"encoding/json"
	"math/rand"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authvestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/stretchr/testify/require"

	"swisstronik/x/vesting/simulation"
	vestingtypes "swisstronik/x/vesting/types"
)

func TestRandomizedGenState(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	authtypes.RegisterInterfaces(registry)
	vestingtypes.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	s := rand.NewSource(1)
	r := rand.New(s)

	simState := module.SimulationState{
		AppParams:    make(simtypes.AppParams),
		Cdc:          cdc,
		Rand:         r,
		NumBonded:    3,
		Accounts:     simtypes.RandomAccounts(r, 300),
		InitialStake: sdkmath.NewInt(1000),
		GenState:     make(map[string]json.RawMessage),
	}

	authsim.RandomizedGenState(&simState, simulation.RandomGenesisAccounts)

	var authGenesis authtypes.GenesisState
	simState.Cdc.MustUnmarshalJSON(simState.GenState[authtypes.ModuleName], &authGenesis)

	require.Equal(t, uint64(0xa0), authGenesis.Params.GetMaxMemoCharacters())
	require.Equal(t, uint64(0x353), authGenesis.Params.GetSigVerifyCostED25519())
	require.Equal(t, uint64(0x29a), authGenesis.Params.GetSigVerifyCostSecp256k1())
	require.Equal(t, uint64(7), authGenesis.Params.GetTxSigLimit())
	require.Equal(t, uint64(5), authGenesis.Params.GetTxSizeCostPerByte())

	genAccounts, err := authtypes.UnpackAccounts(authGenesis.Accounts)
	require.NoError(t, err)
	require.Len(t, genAccounts, 300)
	require.Equal(t, "cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", genAccounts[2].GetAddress().String())
	require.Equal(t, uint64(0), genAccounts[2].GetAccountNumber())
	require.Equal(t, uint64(0), genAccounts[2].GetSequence())

	var (
		base       = 0
		continuous = 0
		delayed    = 0
		monthly    = 0
	)
	for _, acc := range genAccounts {
		require.NoError(t, acc.Validate())
		if _, ok := acc.(*vestingtypes.MonthlyVestingAccount); ok {
			monthly++
		} else if _, ok := acc.(*authvestingtypes.ContinuousVestingAccount); ok {
			continuous++
		} else if _, ok := acc.(*authvestingtypes.DelayedVestingAccount); ok {
			delayed++
		} else if _, ok := acc.(*authtypes.BaseAccount); ok {
			base++
		}
	}
	require.Greater(t, base, 0)
	require.Greater(t, continuous, 0)
	require.Greater(t, delayed, 0)
	require.Greater(t, monthly, 0)
}
