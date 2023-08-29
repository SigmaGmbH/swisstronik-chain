package ante

import (
	"fmt"
	errorsmod "cosmossdk.io/errors"
	sdkerror "github.com/cosmos/cosmos-sdk/types/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// Defines max depth of nested messages
const maxDepth = 5

// RejectNestedMessageDecorator validates if message contains restricted nested message types
type RejectNestedMessageDecorator struct {
	// disabledInnerMessages contains list of message types which cannot be nested 
	disabledInnerMessages []string
}

// NewRejectNestedMessageDecorator returns a decorator to block provided types of messages 
func NewRejectNestedMessageDecorator(disabledInnerMessages ...string) RejectNestedMessageDecorator {
	return RejectNestedMessageDecorator{
		disabledInnerMessages: disabledInnerMessages,
	}
}

func (rnmd RejectNestedMessageDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (newCtx sdk.Context, err error) {
	// Check for authz messages
	if err := rnmd.checkAuthzMessages(tx.GetMsgs(), 0, false); err != nil {
		return ctx, errorsmod.Wrapf(sdkerror.ErrUnauthorized, err.Error())
	}

	return next(ctx, tx, simulate)
}

func (rnmd RejectNestedMessageDecorator) checkAuthzMessages(msgs []sdk.Msg, currentDepth int, isAuthzNestedMessage bool) error {
	if currentDepth >= maxDepth {
		return fmt.Errorf("exceeded max depth of nested messages. Limit is: %d", maxDepth)
	}

	for _, msg := range msgs {
		switch msg := msg.(type) {
		case *authz.MsgExec:
			nestedMessages, err := msg.GetMessages()
			if err != nil {
				return err
			}
			currentDepth++
			if err := rnmd.checkAuthzMessages(nestedMessages, currentDepth, true); err != nil {
				return err
			}
		case *authz.MsgGrant:
			auth, err := msg.GetAuthorization()
			if err != nil {
				return err
			}
			msgType := auth.MsgTypeURL()
			if rnmd.isAuthzDisabledMessage(msgType) {
				return fmt.Errorf("message type is disabled: %s", msgType)
			}
		default:
			msgType := sdk.MsgTypeURL(msg)
			if isAuthzNestedMessage && rnmd.isAuthzDisabledMessage(msgType) {
				return fmt.Errorf("message type is disabled: %s", msgType)
			}	
		}
	}

	return nil
}

func (rnmd RejectNestedMessageDecorator) isAuthzDisabledMessage(msgType string) bool {
	for _, disabledMessageType := range rnmd.disabledInnerMessages {
		if disabledMessageType == msgType {
			return true
		}
	}

	return false
}