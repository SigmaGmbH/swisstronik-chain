package keeper

import (
	"swisstronik/x/vesting/types"
)

var _ types.QueryServer = Keeper{}
