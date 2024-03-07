package keeper

import (
	"swisstronik/x/compliance/types"
)

var _ types.QueryServer = Keeper{}
