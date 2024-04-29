package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/armon/go-metrics"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	atypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	"swisstronik/x/vesting/types"
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

func (k msgServer) HandleCreateMonthlyVestingAccount(goCtx context.Context, msg *types.MsgCreateMonthlyVestingAccount) (*types.MsgCreateMonthlyVestingAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ak := k.Keeper.accountKeeper
	bk := k.Keeper.bankKeeper

	if err := bk.IsSendEnabledCoins(ctx, msg.Amount...); err != nil {
		return nil, err
	}

	from := sdk.MustAccAddressFromBech32(msg.FromAddress)
	to := sdk.MustAccAddressFromBech32(msg.ToAddress)

	if bk.BlockedAddr(to) {
		return nil, errorsmod.Wrapf(
			errortypes.ErrUnauthorized,
			"%s is a blocked address and cannot receive funds", to,
		)
	}

	if acc := ak.GetAccount(ctx, to); acc != nil {
		return nil, errorsmod.Wrapf(errortypes.ErrInvalidRequest, "account %s already exists", msg.ToAddress)
	}

	// Calculate amount and period per each month
	totalCoins := msg.Amount
	amount := totalCoins.QuoInt(sdk.NewInt(msg.Months))

	var periods []atypes.Period
	for i := 0; i < int(msg.Months); i++ {
		period := atypes.Period{Length: types.SecondsOfMonth, Amount: amount}
		periods = append(periods, period)
	}

	baseAccount := authtypes.NewBaseAccountWithAddress(to)
	baseAccount = k.accountKeeper.NewAccount(ctx, baseAccount).(*authtypes.BaseAccount)
	vestingAccount := types.NewMonthlyVestingAccount(
		baseAccount,
		msg.Amount,
		ctx.BlockTime().Unix(),
		msg.CliffDays,
		msg.Months,
	)

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

	if err := bk.SendCoins(ctx, from, to, totalCoins); err != nil {
		return nil, err
	}

	// Emit events
	ctx.EventManager().EmitEvents(
		sdk.Events{
			sdk.NewEvent(
				types.EventTypeMonthlyVestingAccount,
				sdk.NewAttribute(types.AttributeKeyFromAddress, msg.FromAddress),
				sdk.NewAttribute(types.AttributeKeyToAddress, msg.ToAddress),
				sdk.NewAttribute(types.AttributeKeyStartTime, fmt.Sprintf("%d", ctx.BlockTime().Unix())),
				sdk.NewAttribute(types.AttributeKeyCliffDays, fmt.Sprintf("%d", msg.CliffDays)),
				sdk.NewAttribute(types.AttributeKeyMonths, fmt.Sprintf("%d", msg.Months)),
				sdk.NewAttribute(types.AttributeKeyCoins, totalCoins.String()),
			),
		},
	)

	return &types.MsgCreateMonthlyVestingAccountResponse{}, nil
}
