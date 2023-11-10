package keeper

import (
	"context"
	"swisstronik/x/vesting/types"

	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
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

func (k msgServer) CreateMonthlyVestingAccount(goCtx context.Context, msg *types.MsgCreateMonthlyVestingAccount) (*types.MsgCreateMonthlyVestingAccountResponse, error) {
	from, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'from' address: %s", err)
	}

	to, err := sdk.AccAddressFromBech32(msg.ToAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress.Wrapf("invalid 'to' address: %s", err)
	}

	// Is invalid start time
	if msg.StartTime < 1 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid start time of %d, length must be greater than 0", msg.StartTime)
	}

	// Is invalid months
	if msg.Month < 1 {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid months of %d, length must be greater than 0", msg.Month)
	}

	// Is invalid total amount
	totalCoins := msg.Amount
	if !totalCoins.IsAllPositive() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "invalid amount of %s, amount must be greater than 0", msg.Amount)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if acc := k.Keeper.accountKeeper.GetAccount(ctx, to); acc != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	if err := k.Keeper.bankKeeper.IsSendEnabledCoins(ctx, totalCoins...); err != nil {
		return nil, err
	}

	// Calculate monthly amount
	amount := totalCoins.QuoInt(sdk.NewInt(msg.Month))

	var periods []atypes.Period
	for i := 0; i < (int)(msg.Month); i++ {
		period := atypes.Period{Length: types.SecondsOfMonth, Amount: amount}
		periods = append(periods, period)
	}

	baseAccount := authtypes.NewBaseAccountWithAddress(to)
	baseAccount = k.accountKeeper.NewAccount(ctx, baseAccount).(*authtypes.BaseAccount)
	vestingAccount := atypes.NewPeriodicVestingAccount(baseAccount, totalCoins.Sort(), msg.StartTime, periods)

	k.accountKeeper.SetAccount(ctx, vestingAccount)

	defer func() {
		telemetry.IncrCounter(1, "new", "account")

		for _, a := range totalCoins {
			if a.Amount.IsInt64() {
				telemetry.SetGaugeWithLabels(
					[]string{"tx", "msg", "create_monthly_vesting_account"},
					float32(a.Amount.Int64()),
					[]metrics.Label{telemetry.NewLabel("denom", a.Denom)},
				)
			}
		}
	}()

	if err = k.Keeper.bankKeeper.SendCoins(ctx, from, to, totalCoins); err != nil {
		return nil, err
	}

	return &types.MsgCreateMonthlyVestingAccountResponse{}, nil
}
