package compliance

import (
	"fmt"

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

	// Issuer should exist and be verified
	exists, _ := k.IssuerExists(ctx, issuerAddress)
	verified, _ := k.IsAddressVerified(ctx, issuerAddress)
	verified = verified && exists

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVerifyIssuer,
			sdk.NewAttribute(types.AttributeKeyIssuer, p.IssuerAddress),
			sdk.NewAttribute(types.AttributeKeyVerified, fmt.Sprintf("%t", verified)),
		),
	)
	return nil
}
