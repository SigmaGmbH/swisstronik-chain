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

	if err := k.AddOperator(ctx, operator, types.OperatorType_OT_REGULAR); err != nil {
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
		return nil, types.ErrNotOperator
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
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer not exists")
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

func (k msgServer) HandleSetIssuerDetails(goCtx context.Context, msg *types.MsgSetIssuerDetails) (*types.MsgSetIssuerDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Only operator can add issuer
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		return nil, types.ErrNotOperator
	}

	// Make sure that issuer does not exist
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}
	if exists, err := k.IssuerExists(ctx, issuer); exists || err != nil {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer already exists")
	}

	if err := k.SetIssuerDetails(ctx, issuer, msg.Details); err != nil {
		return nil, err
	}

	// If issuer is freshly created, verification status should be false
	if err = k.SetAddressVerificationStatus(ctx, issuer, false); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeAddIssuer,
			sdk.NewAttribute(types.AttributeKeyOperator, msg.Signer),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.Issuer),
			sdk.NewAttribute(types.AttributeKeyIssuerDetails, msg.Details.String()),
		),
	)

	return &types.MsgSetIssuerDetailsResponse{}, nil
}

func (k msgServer) HandleUpdateIssuerDetails(goCtx context.Context, msg *types.MsgUpdateIssuerDetails) (*types.MsgUpdateIssuerDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check validity of signer address
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return nil, err
	}

	// Only operator can update issuer
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		return nil, types.ErrNotOperator
	}

	// Check if issuer exists
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}
	if exists, err := k.IssuerExists(ctx, issuer); !exists || err != nil {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer not exists")
	}

	// Revoke verification if address was verified
	if err := k.SetAddressVerificationStatus(ctx, issuer, false); err != nil {
		return nil, err
	}

	if err := k.SetIssuerDetails(ctx, issuer, msg.Details); err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateIssuer,
			sdk.NewAttribute(types.AttributeKeyOperator, msg.Signer),
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

	// Only operator can remove issuer
	if exists, err := k.OperatorExists(ctx, signer); !exists || err != nil {
		return nil, types.ErrNotOperator
	}

	// Make sure that issuer exists
	issuer, err := sdk.AccAddressFromBech32(msg.Issuer)
	if err != nil {
		return nil, err
	}
	if exists, err := k.IssuerExists(ctx, issuer); !exists || err != nil {
		return nil, errors.Wrap(types.ErrInvalidIssuer, "issuer not exists")
	}

	k.RemoveIssuer(ctx, issuer)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeRemoveIssuer,
			sdk.NewAttribute(types.AttributeKeyOperator, msg.Signer),
			sdk.NewAttribute(types.AttributeKeyIssuer, msg.Issuer),
		),
	)

	return &types.MsgRemoveIssuerResponse{}, nil
}
