package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/compliance/types"
)

func (k Keeper) ExportOperators(ctx sdk.Context) ([]*types.OperatorDetails, error) {
	var (
		allDetails []*types.OperatorDetails
		details    *types.OperatorDetails
		err        error
	)

	k.IterateOperatorDetails(ctx, func(address sdk.AccAddress) (continue_ bool) {
		details, err = k.GetOperatorDetails(ctx, address)
		if err != nil {
			return false
		}
		allDetails = append(allDetails, details)
		return true
	})
	if err != nil {
		return nil, err
	}

	return allDetails, nil
}

func (k Keeper) ExportVerificationDetails(ctx sdk.Context) ([]*types.GenesisVerificationDetails, error) {
	var (
		allVerificationDetails []*types.GenesisVerificationDetails
		details                *types.VerificationDetails
		err                    error
	)

	k.IterateVerificationDetails(ctx, func(id []byte) bool {
		details, err = k.GetVerificationDetails(ctx, id)
		if err != nil {
			return false
		}
		allVerificationDetails = append(allVerificationDetails, &types.GenesisVerificationDetails{Id: id, Details: details})
		return true
	})
	if err != nil {
		return nil, err
	}

	return allVerificationDetails, nil
}

func (k Keeper) ExportAddressDetails(ctx sdk.Context) ([]*types.GenesisAddressDetails, error) {
	var (
		allAddressDetails []*types.GenesisAddressDetails
		details           *types.AddressDetails
		err               error
	)

	k.IterateAddressDetails(ctx, func(address sdk.AccAddress) bool {
		details, err = k.GetAddressDetails(ctx, address)
		if err != nil {
			return false
		}
		allAddressDetails = append(allAddressDetails, &types.GenesisAddressDetails{Address: address.String(), Details: details})
		return true
	})
	if err != nil {
		return nil, err
	}

	return allAddressDetails, nil
}

func (k Keeper) ExportIssuerDetails(ctx sdk.Context) ([]*types.GenesisIssuerDetails, error) {
	var (
		issuerDetails []*types.GenesisIssuerDetails
		details       *types.IssuerDetails
		err           error
	)

	k.IterateIssuerDetails(ctx, func(address sdk.AccAddress) bool {
		details, err = k.GetIssuerDetails(ctx, address)
		if err != nil {
			return false
		}
		issuerDetails = append(issuerDetails, &types.GenesisIssuerDetails{
			Address: address.String(),
			Details: details,
		})
		return true
	})
	if err != nil {
		return nil, err
	}

	return issuerDetails, nil
}

func (k Keeper) ExportHolderPublicKeys(ctx sdk.Context) ([]*types.GenesisHolderPublicKeys, error) {
	var (
		holderPublicKeys []*types.GenesisHolderPublicKeys
	)

	k.IterateHolderPublicKeys(ctx, func(holder sdk.AccAddress) bool {
		publicKey := k.GetHolderPublicKey(ctx, holder)
		if publicKey != nil {
			holderPublicKeys = append(holderPublicKeys, &types.GenesisHolderPublicKeys{
				Address:   holder.String(),
				PublicKey: publicKey,
			})
		}
		return true
	})

	return holderPublicKeys, nil
}

func (k Keeper) ExportLinksVerificationIdToPublicKey(ctx sdk.Context) ([]*types.GenesisLinkVerificationIdToPublicKey, error) {
	var (
		links []*types.GenesisLinkVerificationIdToPublicKey
	)

	k.IterateLinksToPublicKey(ctx, func(verificationId []byte) bool {
		publicKey := k.GetPubKeyByVerificationId(ctx, verificationId)
		if publicKey != nil {
			links = append(links, &types.GenesisLinkVerificationIdToPublicKey{
				Id:        verificationId,
				PublicKey: publicKey,
			})
		}
		return true
	})

	return links, nil
}