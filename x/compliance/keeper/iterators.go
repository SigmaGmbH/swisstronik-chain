package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/compliance/types"
)

func (k Keeper) IterateOperatorDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixOperatorDetails)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		address := types.AccAddressFromKey(key)
		if !callback(address) {
			break
		}
	}
}

func (k Keeper) IterateVerificationDetails(ctx sdk.Context, callback func(id []byte) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationDetails)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		id := types.VerificationIdFromKey(key)
		if !callback(id) {
			break
		}
	}
}

func (k Keeper) IterateAddressDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixAddressDetails)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		address := types.AccAddressFromKey(key)
		if !callback(address) {
			break
		}
	}
}

func (k Keeper) IterateIssuerDetails(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixIssuerDetails)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		address := types.AccAddressFromKey(key)
		if !callback(address) {
			break
		}
	}
}

func (k Keeper) IterateHolderPublicKeys(ctx sdk.Context, callback func(address sdk.AccAddress) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixHolderPublicKeys)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		address := types.AccAddressFromKey(key)
		if !callback(address) {
			break
		}
	}
}

func (k Keeper) IterateLinksToHolder(ctx sdk.Context, callback func(verificationId []byte) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationToHolder)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		id := types.VerificationIdFromKey(key)
		if !callback(id) {
			break
		}
	}
}

func (k Keeper) IterateLinksToPublicKey(ctx sdk.Context, callback func(verificationId []byte) (continue_ bool)) {
	latestVersionIterator := sdk.KVStorePrefixIterator(ctx.KVStore(k.storeKey), types.KeyPrefixVerificationToPubKey)
	defer closeIteratorOrPanic(latestVersionIterator)

	for ; latestVersionIterator.Valid(); latestVersionIterator.Next() {
		key := latestVersionIterator.Key()
		id := types.VerificationIdFromKey(key)
		if !callback(id) {
			break
		}
	}
}

func closeIteratorOrPanic(iterator sdk.Iterator) {
	err := iterator.Close()
	if err != nil {
		panic(err.Error())
	}
}