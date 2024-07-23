package types

import (
	errormod "cosmossdk.io/errors"
)

const (
	codeErrInvalidParam = uint32(iota) + 2 // NOTE: code 1 is reserved for internal errors
	codeErrNotFoundVestingAccount
	codeErrInsufficientUnlockedCoins
)

// x/vesting module sentinel errors
var (
	ErrInvalidParam              = errormod.Register(ModuleName, codeErrInvalidParam, "invalid param provided")
	ErrNotFoundVestingAccount    = errormod.Register(ModuleName, codeErrNotFoundVestingAccount, "not found vesting account")
	ErrInsufficientUnlockedCoins = errormod.Register(ModuleName, codeErrInsufficientUnlockedCoins, "insufficient unlocked coins")
)
