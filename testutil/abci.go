package testutil

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"

	"swisstronik/app"
	"swisstronik/encoding"
	"swisstronik/utils"
)

var (
	DefaultFee = sdk.NewCoin(utils.BaseDenom, sdkmath.NewIntFromUint64(uint64(1000000)))
)

// CosmosTxArgs contains the params to create a cosmos tx
type CosmosTxArgs struct {
	// TxCfg is the client transaction config
	TxCfg client.TxConfig
	// Priv is the private key that will be used to sign the tx
	Priv cryptotypes.PrivKey
	// ChainID is the chain's id on cosmos format
	ChainID string
	// Gas to be used on the tx
	Gas uint64
	// GasPrice to use on tx
	GasPrice *sdkmath.Int
	// Fees is the fee to be used on the tx (amount and denom)
	Fees sdk.Coins
	// FeeGranter is the account address of the fee granter
	FeeGranter sdk.AccAddress
	// Msgs slice of messages to include on the tx
	Msgs []sdk.Msg
}

// PrepareTx creates a cosmos tx and signs it with the provided messages and private key.
// It returns the signed transaction and an error
func PrepareTx(
	ctx sdk.Context,
	app *app.App,
	args CosmosTxArgs,
) (authsigning.Tx, error) {
	txBuilder := args.TxCfg.NewTxBuilder()

	txBuilder.SetGasLimit(args.Gas)

	var fees sdk.Coins
	if args.GasPrice != nil {
		fees = sdk.Coins{{Denom: utils.BaseDenom, Amount: args.GasPrice.MulRaw(int64(args.Gas))}}
	} else {
		fees = sdk.Coins{DefaultFee}
	}
	txBuilder.SetFeeAmount(fees)
	if err := txBuilder.SetMsgs(args.Msgs...); err != nil {
		return nil, err
	}

	txBuilder.SetFeeGranter(args.FeeGranter)

	return signTx(ctx, app, args, txBuilder)
}

// signTx signs the cosmos transaction on the txBuilder provided using
// the provided private key
func signTx(
	ctx sdk.Context,
	app *app.App,
	args CosmosTxArgs,
	txBuilder client.TxBuilder,
) (authsigning.Tx, error) {
	addr := sdk.AccAddress(args.Priv.PubKey().Address().Bytes())
	seq, err := app.AccountKeeper.GetSequence(ctx, addr)
	if err != nil {
		return nil, err
	}

	// First round: we gather all the signer infos. We use the "set empty
	// signature" hack to do that.
	sigV2 := signing.SignatureV2{
		PubKey: args.Priv.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  args.TxCfg.SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: seq,
	}

	sigsV2 := []signing.SignatureV2{sigV2}

	if err := txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}

	// Second round: all signer infos are set, so each signer can sign.
	accNumber := app.AccountKeeper.GetAccount(ctx, addr).GetAccountNumber()
	signerData := authsigning.SignerData{
		ChainID:       args.ChainID,
		AccountNumber: accNumber,
		Sequence:      seq,
	}
	sigV2, err = tx.SignWithPrivKey(
		args.TxCfg.SignModeHandler().DefaultMode(),
		signerData,
		txBuilder, args.Priv, args.TxCfg,
		seq,
	)
	if err != nil {
		return nil, err
	}

	sigsV2 = []signing.SignatureV2{sigV2}
	if err = txBuilder.SetSignatures(sigsV2...); err != nil {
		return nil, err
	}
	return txBuilder.GetTx(), nil
}

// DeliverTx delivers a cosmos tx for a given set of msgs
func DeliverTx(
	ctx sdk.Context,
	appS *app.App,
	priv cryptotypes.PrivKey,
	gasPrice *sdkmath.Int,
	msgs ...sdk.Msg,
) (abci.ResponseDeliverTx, error) {
	txConfig := encoding.MakeConfig(app.ModuleBasics).TxConfig
	signedTx, err := PrepareTx(
		ctx,
		appS,
		CosmosTxArgs{
			TxCfg:    txConfig,
			Priv:     priv,
			ChainID:  ctx.ChainID(),
			Gas:      1_000_000,
			GasPrice: gasPrice,
			Msgs:     msgs,
		},
	)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}
	return BroadcastTxBytes(appS, txConfig.TxEncoder(), signedTx)
}

// BroadcastTxBytes encodes a transaction and calls DeliverTx on the app.
func BroadcastTxBytes(app *app.App, txEncoder sdk.TxEncoder, tx sdk.Tx) (abci.ResponseDeliverTx, error) {
	// bz are bytes to be broadcasted over the network
	bz, err := txEncoder(tx)
	if err != nil {
		return abci.ResponseDeliverTx{}, err
	}

	req := abci.RequestDeliverTx{Tx: bz}
	res := app.BaseApp.DeliverTx(req)
	if res.Code != 0 {
		return abci.ResponseDeliverTx{}, errorsmod.Wrapf(errortypes.ErrInvalidRequest, res.Log)
	}

	return res, nil
}

// CommitAndCreateNewCtx commits a block at a given time creating a ctx with the current settings
// This is useful to keep test settings that could be affected by EndBlockers, e.g.
// setting a baseFee == 0 and expecting this condition to continue after commit
func CommitAndCreateNewCtx(ctx sdk.Context, app *app.App, t time.Duration) (sdk.Context, error) {
	header := ctx.BlockHeader()
	app.EndBlock(abci.RequestEndBlock{Height: header.Height})
	_ = app.Commit()

	header.Height++
	header.Time = header.Time.Add(t)
	header.AppHash = app.LastCommitID().Hash
	app.BeginBlock(abci.RequestBeginBlock{
		Header: header,
	})

	// NewContext function keeps the multistore
	// but resets other context fields
	// GasMeter is set as InfiniteGasMeter
	newCtx := app.BaseApp.NewContext(false, header)
	// set the reseted fields to keep the current ctx settings
	newCtx = newCtx.WithMinGasPrices(ctx.MinGasPrices())
	newCtx = newCtx.WithEventManager(ctx.EventManager())
	newCtx = newCtx.WithKVGasConfig(ctx.KVGasConfig())
	newCtx = newCtx.WithTransientKVGasConfig(ctx.TransientKVGasConfig())

	return newCtx, nil
}
