package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	codeErrInvalidParam = uint32(iota) + 2 // NOTE: code 1 is reserved for internal errors
)

// x/vesting module sentinel errors
var (
	ErrInvalidParam = sdkerrors.Register(ModuleName, codeErrInvalidParam, "invalid param provided")
)
