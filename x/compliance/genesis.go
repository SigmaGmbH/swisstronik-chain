package compliance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Restore issuers
	for _, issuerData := range genState.Issuers {
		issuerAddress, err := sdk.AccAddressFromBech32(issuerData.Address)
		if err != nil {
			panic(err)
		}
		if err := k.SetIssuerDetails(ctx, issuerAddress, issuerData.Details); err != nil {
			panic(err)
		}
	}

	// Restore accounts
	for _, addressData := range genState.AddressDetails {
		address, err := sdk.AccAddressFromBech32(addressData.Address)
		if err != nil {
			panic(err)
		}

		if err := k.SetAddressDetails(ctx, address, addressData.Details); err != nil {
			panic(err)
		}
	}

	// Restore verification data
	for _, verificationData := range genState.VerificationDetails {
		if err := k.SetVerificationDetails(ctx, verificationData.Id, verificationData.Details); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	issuers, err := k.ExportIssuerAccounts(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Issuers = issuers

	addressDetails, err := k.ExportAddressDetails(ctx)
	if err != nil {
		panic(err)
	}
	genesis.AddressDetails = addressDetails

	verificationDetails, err := k.ExportVerificationDetails(ctx)
	if err != nil {
		panic(err)
	}
	genesis.VerificationDetails = verificationDetails

	return genesis
}
