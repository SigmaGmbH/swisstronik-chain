package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"swisstronik/testutil/sample"
)

func TestMsgCreateMonthlyVestingAccount_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateMonthlyVestingAccount
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateMonthlyVestingAccount{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateMonthlyVestingAccount{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
