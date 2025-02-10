package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ComplianceKeeper interface {
	IterateOperatorDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool))
	IterateIssuerDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool))
	IterateAddressDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool))
	GetOperatorDetails(ctx sdk.Context, operator sdk.AccAddress) (*OperatorDetails, error)
	GetIssuerDetails(ctx sdk.Context, issuerAddress sdk.AccAddress) (*IssuerDetails, error)
	SetIssuerDetails(ctx sdk.Context, issuerAddress sdk.AccAddress, details *IssuerDetails) error
	GetAddressDetails(ctx sdk.Context, address sdk.AccAddress) (*AddressDetails, error)
	LinkVerificationToHolder(ctx sdk.Context, userAddress sdk.AccAddress, verificationId []byte) error
}
