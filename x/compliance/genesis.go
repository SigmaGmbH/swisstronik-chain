package compliance

import (
	"bytes"
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Restore initial operators
	for _, operatorData := range genState.Operators {
		address, err := sdk.AccAddressFromBech32(operatorData.Operator)
		if err != nil {
			panic(err)
		}
		if operatorData.OperatorType <= types.OperatorType_OT_UNSPECIFIED || operatorData.OperatorType > types.OperatorType_OT_REGULAR {
			panic(errors.Wrap(types.ErrInvalidParam, "operator type is undefined"))
		}
		if err = k.AddOperator(ctx, address, operatorData.OperatorType); err != nil {
			panic(err)
		}
	}

	// Restore issuers
	for _, issuerData := range genState.IssuerDetails {
		address, err := sdk.AccAddressFromBech32(issuerData.Address)
		if err != nil {
			panic(err)
		}
		_, err = sdk.AccAddressFromBech32(issuerData.Details.Creator)
		if err != nil {
			panic(err)
		}
		if err = k.SetIssuerDetails(ctx, address, issuerData.Details); err != nil {
			panic(err)
		}
	}

	// Restore linked public keys to verification id
	for _, verificationToPublicKeyData := range genState.LinksToPublicKey {
		if verificationToPublicKeyData.Id == nil {
			panic(errors.Wrap(types.ErrInvalidParam, "given verification id is empty"))
		}
		if verificationToPublicKeyData.PublicKey == nil {
			panic(errors.Wrap(types.ErrInvalidParam, "given public key is nil"))
		}

		if err := k.LinkVerificationIdToPubKey(ctx, verificationToPublicKeyData.PublicKey, verificationToPublicKeyData.Id); err != nil {
			panic(err)
		}
	}

	// Restore holders public keys
	for _, publicKeyData := range genState.PublicKeys {
		address, err := sdk.AccAddressFromBech32(publicKeyData.Address)
		if err != nil {
			panic(err)
		}

		if publicKeyData.PublicKey == nil {
			panic(errors.Wrap(types.ErrInvalidParam, "public key is nil"))
		}

		k.SetHolderPublicKeyBytes(ctx, address, publicKeyData.PublicKey)
	}

	// Restore verification data
	for _, verificationData := range genState.VerificationDetails {
		// Check if issuer address is valid
		address, err := sdk.AccAddressFromBech32(verificationData.Details.IssuerAddress)
		if err != nil {
			panic(err)
		}
		if exists, err := k.IssuerExists(ctx, address); !exists || err != nil {
			panic(err)
		}
		// Check the issuance timestamp and proof
		if verificationData.Details.IssuanceTimestamp < 1 ||
			(verificationData.Details.ExpirationTimestamp > 0 && verificationData.Details.IssuanceTimestamp >= verificationData.Details.ExpirationTimestamp) {
			panic(errors.Wrap(types.ErrInvalidParam, "invalid issuance timestamp"))
		}
		if len(verificationData.Details.OriginalData) < 1 {
			panic(errors.Wrap(types.ErrInvalidParam, "empty proof data"))
		}

		// Not the most efficient implementation, but it will not destroy genesis state
		var userAddress sdk.AccAddress
		for _, addressData := range genState.AddressDetails {
			for _, addressVerification := range addressData.Details.Verifications {
				if bytes.Equal(verificationData.Id, addressVerification.VerificationId) {
					userAddress, err = sdk.AccAddressFromBech32(addressData.Address)
					if err != nil {
						panic(err)
					}
					break
				}
			}
		}

		if err = k.SetVerificationDetails(ctx, userAddress, verificationData.Id, verificationData.Details); err != nil {
			panic(err)
		}

		if verificationData.Details.IsRevoked {
			if err = k.MarkVerificationDetailsAsRevoked(ctx, verificationData.Id); err != nil {
				panic(err)
			}
		}
	}

	// Restore accounts
	for _, addressData := range genState.AddressDetails {
		address, err := sdk.AccAddressFromBech32(addressData.Address)
		if err != nil {
			panic(err)
		}

		// If address is verified, verification data must exist and issuer must be valid
		for _, verificationData := range addressData.Details.Verifications {
			// Check if issuer is valid
			issuer, err := sdk.AccAddressFromBech32(verificationData.IssuerAddress)
			if err != nil {
				panic(err)
			}
			if exists, err := k.IssuerExists(ctx, issuer); !exists || err != nil {
				panic(err)
			}
			// Check if verification data exists
			if verificationData.VerificationId == nil {
				panic(errors.Wrap(types.ErrInvalidParam, "verification id is nil"))
			}
			if verificationData.Type <= types.VerificationType_VT_UNSPECIFIED || verificationData.Type > types.VerificationType_VT_CREDIT_SCORE {
				panic(errors.Wrap(types.ErrInvalidParam, "verification type is undefined"))
			}
			if _, err = k.GetVerificationDetails(ctx, verificationData.VerificationId); err != nil {
				panic(err)
			}
		}

		if err = k.SetAddressDetails(ctx, address, addressData.Details); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	operators, err := k.ExportOperators(ctx)
	if err != nil {
		panic(err)
	}
	genesis.Operators = operators

	issuerDetails, err := k.ExportIssuerDetails(ctx)
	if err != nil {
		panic(err)
	}
	genesis.IssuerDetails = issuerDetails

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

	holderPublicKeys, err := k.ExportHolderPublicKeys(ctx)
	if err != nil {
		panic(err)
	}
	genesis.PublicKeys = holderPublicKeys

	linksToPublicKey, err := k.ExportLinksVerificationIdToPublicKey(ctx)
	if err != nil {
		panic(err)
	}
	genesis.LinksToPublicKey = linksToPublicKey

	return genesis
}
