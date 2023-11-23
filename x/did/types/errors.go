package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	codeErrBadRequest = uint32(iota) + 2 // NOTE: code 1 is reserved for internal errors
	codeErrInvalidSignature
	codeErrSignatureNotFound
	codeErrDIDDocumentExists
	codeErrDIDDocumentNotFound
	codeErrVerificationMethodNotFound
	codeErrUnexpectedDidVersion
	codeErrBasicValidation
	codeErrNamespaceValidation
	codeErrDIDDocDeactivated
	codeErrUnpackStateValue
	codeErrInternal
	codeErrResourceExists
)

var (
	ErrBadRequest                 = sdkerrors.Register(ModuleName, codeErrBadRequest, "bad request")
	ErrInvalidSignature           = sdkerrors.Register(ModuleName, codeErrInvalidSignature, "invalid signature detected")
	ErrSignatureNotFound          = sdkerrors.Register(ModuleName, codeErrSignatureNotFound, "signature is required but not found")
	ErrDIDDocumentExists          = sdkerrors.Register(ModuleName, codeErrDIDDocumentExists, "DID Document exists")
	ErrDIDDocumentNotFound        = sdkerrors.Register(ModuleName, codeErrDIDDocumentNotFound, "DID Document not found")
	ErrVerificationMethodNotFound = sdkerrors.Register(ModuleName, codeErrVerificationMethodNotFound, "verification method not found")
	ErrUnexpectedDidVersion       = sdkerrors.Register(ModuleName, codeErrUnexpectedDidVersion, "unexpected DID version")
	ErrBasicValidation            = sdkerrors.Register(ModuleName, codeErrBasicValidation, "basic validation failed")
	ErrNamespaceValidation        = sdkerrors.Register(ModuleName, codeErrNamespaceValidation, "DID namespace validation failed")
	ErrDIDDocumentDeactivated     = sdkerrors.Register(ModuleName, codeErrDIDDocDeactivated, "DID Document already deactivated")
	ErrUnpackStateValue           = sdkerrors.Register(ModuleName, codeErrUnpackStateValue, "invalid did state value")
	ErrInternal                   = sdkerrors.Register(ModuleName, codeErrInternal, "internal error")
	ErrResourceExists             = sdkerrors.Register(ModuleName, codeErrResourceExists, "resource exists")
)
