package did

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/did/keeper"
	"swisstronik/x/did/types"
)

// InitGenesis initializes module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState *types.GenesisState) {
	// Set params
	k.SetParams(ctx, genState.Params)

	// Set DID documents
	for _, versionSet := range genState.VersionSets {
		for _, didDoc := range versionSet.DidDocs {
			err := k.SetDIDDocumentVersion(&ctx, didDoc, false)
			if err != nil {
				panic(err)
			}
		}

		err := k.SetLatestDIDDocumentVersion(&ctx, versionSet.DidDocs[0].DidDoc.Id, versionSet.LatestVersion)
		if err != nil {
			panic(err)
		}
	}

	// Set DID resources
	for _, resource := range genState.Resources {
		if err := k.SetResource(&ctx, resource); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	docs, err := k.GetAllDIDDocuments(&ctx)
	if err != nil {
		panic(err)
	}

	resourceList, err := k.GetAllResources(&ctx)
	if err != nil {
		panic(err.Error())
	}

	genesis := types.GenesisState{
		VersionSets: docs,
		Resources:   resourceList,
		Params:      k.GetParams(ctx),
	}

	return &genesis
}
