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
	codeErrNotAuthorized
	codeErrInvalidOperator
	codeErrNotOperator
	codeErrInvalidIssuer
)

var (
	ErrBadRequest                 = sdkerrors.Register(ModuleName, codeErrBadRequest, "bad request")
	ErrInvalidSignature           = sdkerrors.Register(ModuleName, codeErrInvalidSignature, "invalid signature detected")
	ErrSignatureNotFound          = sdkerrors.Register(ModuleName, codeErrSignatureNotFound, "signature is required but not found")
	ErrBasicValidation            = sdkerrors.Register(ModuleName, codeErrBasicValidation, "basic validation failed")
	ErrInvalidParam               = sdkerrors.Register(ModuleName, codeErrInvalidParam, "invalid param provided")
	ErrNotAuthorized              = sdkerrors.Register(ModuleName, codeErrNotAuthorized, "not authorized")
	ErrInvalidOperator            = sdkerrors.Register(ModuleName, codeErrInvalidOperator, "invalid operator")
	ErrNotOperatorOrIssuerCreator = sdkerrors.Register(ModuleName, codeErrNotOperator, "signer is not operator or issuer creator")
	ErrInvalidIssuer              = sdkerrors.Register(ModuleName, codeErrInvalidIssuer, "invalid issuer")
)
