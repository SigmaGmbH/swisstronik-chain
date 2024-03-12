package keeper

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
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

func (k msgServer) SetAddressInfo(goCtx context.Context, msg *types.MsgSetAddressInfo) (*types.MsgSetAddressInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ethUserAddress := common.HexToAddress(msg.UserAddress)
	if err := k.SetAddressInfoRaw(ctx, ethUserAddress.Bytes(), msg.Data); err != nil {
		return nil, err
	}

	return &types.MsgSetAddressInfoResponse{}, nil
}
