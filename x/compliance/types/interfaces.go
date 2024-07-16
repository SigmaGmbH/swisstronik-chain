package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ComplianceKeeper interface {
	IterateOperatorDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool))
	IterateIssuerDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool))
	GetOperatorDetails(ctx sdk.Context, operator sdk.AccAddress) (*OperatorDetails, error)
	GetIssuerDetails(ctx sdk.Context, issuerAddress sdk.AccAddress) (*IssuerDetails, error)
	GetIssuerCreator(ctx sdk.Context, issuerAddress sdk.AccAddress) sdk.AccAddress
	SetIssuerCreator(ctx sdk.Context, issuerAddress, issuerCreatorAddress sdk.AccAddress)
}
