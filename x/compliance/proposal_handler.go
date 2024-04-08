package compliance

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"swisstronik/x/compliance/keeper"
	"swisstronik/x/compliance/types"
)

func NewComplianceProposalHandler(k *keeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *types.VerifyIssuerProposal:
			return handleVerifyIssuerProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", types.ModuleName, c)
		}
	}
}

func handleVerifyIssuerProposal(ctx sdk.Context, k *keeper.Keeper, p *types.VerifyIssuerProposal) error {
	issuerAddress, err := sdk.AccAddressFromBech32(p.IssuerAddress)
	if err != nil {
		return err
	}
	// Issuer should be verified
	verified, err := k.IsAddressVerified(ctx, issuerAddress)
	if err != nil {
		return err
	}
	if !verified {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "issuer not verified")
	}

	// Issuer should exist
	issuerDetails, err := k.GetIssuerDetails(ctx, issuerAddress)
	if err != nil {
		return err
	}
	// Same checking with IssuerExists
	if len(issuerDetails.Operator) == 0 {
		// If issuer not exist
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "issuer not exist")
	}

	// Compare issuer details with state in keeper
	if issuerDetails != p.IssuerDetails {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "issuer details not match")
	}

	return nil
}
