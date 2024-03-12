package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	codeErrBadRequest = uint32(iota) + 2 // NOTE: code 1 is reserved for internal errors
	codeErrInvalidSignature
	codeErrSignatureNotFound
	codeErrBasicValidation
	codeErrInvalidParam
)

var (
	ErrBadRequest        = sdkerrors.Register(ModuleName, codeErrBadRequest, "bad request")
	ErrInvalidSignature  = sdkerrors.Register(ModuleName, codeErrInvalidSignature, "invalid signature detected")
	ErrSignatureNotFound = sdkerrors.Register(ModuleName, codeErrSignatureNotFound, "signature is required but not found")
	ErrBasicValidation   = sdkerrors.Register(ModuleName, codeErrBasicValidation, "basic validation failed")
	ErrInvalidParam      = sdkerrors.Register(ModuleName, codeErrInvalidParam, "invalid param provided")
)
