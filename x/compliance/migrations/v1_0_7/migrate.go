package v1_0_7

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"swisstronik/x/compliance/types"
)

func MigrateStore(ctx sdk.Context, k types.ComplianceKeeper) error {
	println("MIGRATING TO 1.0.7")
	// Link verification id -> holder
	k.IterateAddressDetails(ctx, func(addr sdk.AccAddress) (continue_ bool) {
		addressDetails, err := k.GetAddressDetails(ctx, addr)
		if err != nil {
			return false
		}

		for _, verification := range addressDetails.Verifications {
			if err = k.LinkVerificationToHolder(ctx, addr, verification.VerificationId); err != nil {
				return false
			}
		}
		return true
	})

	return nil
}
