package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"swisstronik/app"
	evmtypes "github.com/SigmaGmbH/evm-module/types"
)

func InitSDKConfig() {
	// Set prefixes
	accountPubKeyPrefix := app.AccountAddressPrefix + "pub"
	validatorAddressPrefix := app.AccountAddressPrefix + "valoper"
	validatorPubKeyPrefix := app.AccountAddressPrefix + "valoperpub"
	consNodeAddressPrefix := app.AccountAddressPrefix + "valcons"
	consNodePubKeyPrefix := app.AccountAddressPrefix + "valconspub"

	config := sdk.GetConfig()

	// Set global prefixes to be used when serializing addresses and public keys to bech32 string
	config.SetBech32PrefixForAccount(app.AccountAddressPrefix, accountPubKeyPrefix)
	config.SetBech32PrefixForValidator(validatorAddressPrefix, validatorPubKeyPrefix)
	config.SetBech32PrefixForConsensusNode(consNodeAddressPrefix, consNodePubKeyPrefix)

	// Set global coin type to be used in HD wallets
	config.SetCoinType(evmtypes.Bip44CoinType)
	config.SetPurpose(sdk.Purpose)                      

	// Set and seal config
	config.Seal()
}
