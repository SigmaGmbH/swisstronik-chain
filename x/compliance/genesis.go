package compliance

import (
	"cosmossdk.io/errors"
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
		// Operator address must not be empty
		if len(issuerData.Details.Operator) < 1 {
			panic(errors.Wrap(types.ErrInvalidParam, "empty operator of issuer"))
		}
		if err := k.SetIssuerDetails(ctx, issuerAddress, issuerData.Details); err != nil {
			panic(err)
		}
	}

	// Restore verification data
	for _, verificationData := range genState.VerificationDetails {
		// TODO, Check if issuer address is valid or timestamp of issuance/expiration are valid

		// Check if issuer address is valid
		issuerAddress, err := sdk.AccAddressFromBech32(verificationData.Details.IssuerAddress)
		if err != nil {
			panic(err)
		}
		if exists, err := k.IssuerExists(ctx, issuerAddress); !exists || err != nil {
			panic(err)
		}
		// Check the issuance timestamp and proof
		if verificationData.Details.IssuanceTimestamp >= verificationData.Details.ExpirationTimestamp {
			panic(errors.Wrap(types.ErrInvalidParam, "invalid issuance timestamp"))
		}
		if len(verificationData.Details.OriginalData) < 1 {
			panic(errors.Wrap(types.ErrInvalidParam, "empty proof data"))
		}

		if err := k.SetVerificationDetails(ctx, verificationData.Id, verificationData.Details); err != nil {
			panic(err)
		}
	}

	// Restore accounts
	for _, addressData := range genState.AddressDetails {
		address, err := sdk.AccAddressFromBech32(addressData.Address)
		if err != nil {
			panic(err)
		}

		// If address is verified, verification data must exist and issuer must be valid
		if addressData.Details.IsVerified {
			for _, verificationData := range addressData.Details.Verifications {
				// Check if issuer is valid
				issuerAddress, err := sdk.AccAddressFromBech32(verificationData.IssuerAddress)
				if err != nil {
					panic(err)
				}
				if exists, err := k.IssuerExists(ctx, issuerAddress); !exists || err != nil {
					panic(err)
				}
				// Check if verification data exists
				if verificationData.VerificationId == nil {
					panic(errors.Wrap(types.ErrInvalidParam, "verification id is nil"))
				}
				if details, err := k.GetVerificationDetails(ctx, verificationData.VerificationId); details == nil || err != nil {
					panic(err)
				}
			}
		}

		if err := k.SetAddressDetails(ctx, address, addressData.Details); err != nil {
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
