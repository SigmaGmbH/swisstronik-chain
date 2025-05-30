package keeper

import (
	"context"
	"strconv"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"swisstronik/x/compliance/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) HandleAddOperator(goCtx context.Context, msg *types.MsgAddOperator) (*types.MsgAddOperatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Only operator can add regular operator
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		return nil, types.ErrNotOperator
	}

	// Check validity of operator addresses
	operator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return nil, err
	}

	// Do not allow to add duplicated operator
	exists, err := k.OperatorExists(ctx, operator)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Wrapf(types.ErrInvalidOperator, "operator already exists")
	}

	if err = k.AddOperator(ctx, operator, types.OperatorType_OT_REGULAR); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddOperator,
			sdk.NewAttribute(types.AttributeKeyOperator, msg.Operator),
		),
	)

	return &types.MsgAddOperatorResponse{}, nil
}

func (k msgServer) HandleRemoveOperator(goCtx context.Context, msg *types.MsgRemoveOperator) (*types.MsgRemoveOperatorResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Only operator can remove regular operator
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		return nil, types.ErrNotOperatorOrIssuerCreator
	}

	// Check validity of operator addresses
	operator, err := sdk.AccAddressFromBech32(msg.Operator)
	if err != nil {
		return nil, err
	}

	// Do not allow to remove itself
	if signer.Equals(operator) {
		return nil, errors.Wrapf(types.ErrInvalidOperator, "same operator")
	}

	// Only allowed to remove regular operator
	if err = k.RemoveRegularOperator(ctx, operator); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveOperator,
			sdk.NewAttribute(types.AttributeKeyOperator, msg.Operator),
		),
	)

	return &types.MsgRemoveOperatorResponse{}, nil
}

func (k msgServer) HandleRevokeVerification(goCtx context.Context, msg *types.MsgRevokeVerification) (*types.MsgRevokeVerificationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Check if verification details exist
	verificationDetails, err := k.GetVerificationDetails(ctx, msg.VerificationId)
	if err != nil {
		return nil, err
	}
	if verificationDetails.IsEmpty() {
		return nil, errors.Wrap(types.ErrInvalidParam, "verification does not exist")
	}

	// Check validity of issuer address
	issuer, err := sdk.AccAddressFromBech32(verificationDetails.IssuerAddress)
	if err != nil {
		return nil, err
	}

	// Check if signer is operator or issuer creator
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		// If signer is not an operator, check if it's issuer creator
		details, err := k.GetIssuerDetails(ctx, issuer)
		if err != nil || len(details.Name) < 1 {
			return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer details not found")
		}

		if details.Creator != signer.String() {
			return nil, errors.Wrap(types.ErrNotOperatorOrIssuerCreator, "issuer creator or operator does not match")
		}
	}

	if err = k.MarkVerificationDetailsAsRevoked(ctx, msg.VerificationId); err != nil {
		return nil, err
	}

	return &types.MsgRevokeVerificationResponse{}, nil
}

func (k msgServer) HandleSetVerificationStatus(goCtx context.Context, msg *types.MsgSetVerificationStatus) (*types.MsgSetVerificationStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Only operator can set issuer's verification status, (todo, for now, in centralized way)
	// NOTE, for now, use centralized solution, will move to decentralized way later.
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		return nil, types.ErrNotOperator
	}

	// Check validity of issuer addresses
	issuer, err := sdk.AccAddressFromBech32(msg.IssuerAddress)
	if err != nil {
		return nil, err
	}

	if exists, err := k.IssuerExists(ctx, issuer); !exists || err != nil {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer does not exist")
	}

	if err = k.SetAddressVerificationStatus(ctx, issuer, msg.IsVerified); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeVerifyIssuer,
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.IssuerAddress),
			sdk.NewAttribute(types.AttributeKeyVerificationStatus, strconv.FormatBool(msg.IsVerified)),
		),
	)

	return &types.MsgSetVerificationStatusResponse{}, nil
}

func (k msgServer) HandleCreateIssuer(goCtx context.Context, msg *types.MsgCreateIssuer) (*types.MsgCreateIssuerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Make sure that issuer does not exist
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}
	if exists, err := k.IssuerExists(ctx, issuer); exists || err != nil {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer already exists")
	}

	msg.Details.Creator = signer.String()

	// Store issuer details with creator address
	if err = k.SetIssuerDetails(ctx, issuer, msg.Details); err != nil {
		return nil, err
	}

	// If issuer is freshly created, verification status should be false
	if err = k.SetAddressVerificationStatus(ctx, issuer, false); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddIssuer,
			sdk.NewAttribute(types.AttributeKeyIssuerCreator, msg.Signer),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.Issuer),
			sdk.NewAttribute(types.AttributeKeyIssuerDetails, msg.Details.String()),
		),
	)

	return &types.MsgCreateIssuerResponse{}, nil
}

func (k msgServer) HandleUpdateIssuerDetails(goCtx context.Context, msg *types.MsgUpdateIssuerDetails) (*types.MsgUpdateIssuerDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Check if issuer exists
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}

	details, err := k.GetIssuerDetails(ctx, issuer)
	if err != nil || len(details.Name) < 1 {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer does not exist")
	}

	// Operator or issuer creator can update issuer
	if details.Creator != signer.String() {
		if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
			// If signer is neither an operator nor issuer creator
			return nil, errors.Wrap(types.ErrNotOperatorOrIssuerCreator, "issuer creator does not match")
		}
	}

	// Revoke verification if address was verified
	if err = k.SetAddressVerificationStatus(ctx, issuer, false); err != nil {
		return nil, err
	}

	if err = k.SetIssuerDetails(ctx, issuer, msg.Details); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateIssuer,
			sdk.NewAttribute(types.AttributeKeyIssuerCreator, msg.Signer),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.Issuer),
			sdk.NewAttribute(types.AttributeKeyIssuerDetails, msg.Details.String()),
		),
	)

	return &types.MsgUpdateIssuerDetailsResponse{}, nil
}

func (k msgServer) HandleRemoveIssuer(goCtx context.Context, msg *types.MsgRemoveIssuer) (*types.MsgRemoveIssuerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Make sure that issuer exists
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}

	details, err := k.GetIssuerDetails(ctx, issuer)
	if err != nil || len(details.Name) < 1 {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer does not exist")
	}

	// Operator or issuer creator can remove issuer
	if details.Creator != signer.String() {
		if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
			// If signer is neither an operator nor issuer creator
			return nil, errors.Wrap(types.ErrNotOperatorOrIssuerCreator, "issuer creator does not match")
		}
	}

	k.RemoveIssuer(ctx, issuer)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveIssuer,
			sdk.NewAttribute(types.AttributeKeyIssuerCreator, msg.Signer),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.Issuer),
		),
	)

	return &types.MsgRemoveIssuerResponse{}, nil
}

func (k msgServer) HandleAttachHolderPublicKey(goCtx context.Context, msg *types.MsgAttachHolderPublicKey) (*types.MsgAttachHolderPublicKeyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	if err = k.SetHolderPublicKey(ctx, signer, msg.HolderPublicKey); err != nil {
		return nil, err
	}

	return &types.MsgAttachHolderPublicKeyResponse{}, nil
}

func (k msgServer) HandleConvertCredential(goCtx context.Context, msg *types.MsgConvertCredential) (*types.MsgConvertCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	holderAddress, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Check if signer is owner of credential
	credentialOwner := k.getHolderByVerificationId(ctx, msg.VerificationId)
	if credentialOwner.String() != msg.Signer {
		return nil, errors.Wrap(types.ErrBadRequest, "signer is not credential holder")
	}

	holderPublicKey := k.GetHolderPublicKey(ctx, holderAddress)
	if holderPublicKey == nil {
		return nil, errors.Wrap(types.ErrBadRequest, "holder public key not found. Please attach it")
	}

	isVerificationRevoked, err := k.IsVerificationRevoked(ctx, msg.VerificationId)
	if err != nil {
		return nil, err
	}
	if isVerificationRevoked {
		return nil, errors.Wrap(types.ErrBadRequest, "credential was revoked")
	}

	details, err := k.GetVerificationDetails(ctx, msg.VerificationId)
	if err != nil {
		return nil, err
	}

	if details.Type == types.VerificationType_VT_UNSPECIFIED {
		return nil, errors.Wrap(types.ErrBadRequest, "verification not found")
	}

	issuerAddress, err := sdk.AccAddressFromBech32(details.IssuerAddress)
	if err != nil {
		return nil, err
	}

	credentialValue := &types.ZKCredential{
		Type:                details.Type,
		IssuerAddress:       issuerAddress.Bytes(),
		HolderPublicKey:     holderPublicKey,
		ExpirationTimestamp: details.ExpirationTimestamp,
		IssuanceTimestamp:   details.IssuanceTimestamp,
	}
	credentialHash, err := credentialValue.Hash()
	if err != nil {
		return nil, err
	}

	isIncluded, err := k.IsIncludedInIssuanceTree(ctx, credentialHash)
	if err != nil {
		return nil, err
	}

	if isIncluded {
		return nil, errors.Wrap(types.ErrBadRequest, "credential is already included in issuance tree")
	}

	if err = k.AddCredentialHashToIssued(ctx, credentialHash); err != nil {
		return nil, err
	}

	if err = k.LinkVerificationIdToPubKey(ctx, holderPublicKey, msg.VerificationId); err != nil {
		return nil, err
	}

	return &types.MsgConvertCredentialResponse{}, nil
}
