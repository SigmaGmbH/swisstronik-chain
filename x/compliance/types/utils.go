package types

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"strings"
)

// ParseAddress tries to convert provided bech32 or hex address into sdk.AccAddress
func ParseAddress(input string) (sdk.AccAddress, error) {
	cfg := sdk.GetConfig()

	var err error
	if !strings.HasPrefix(input, cfg.GetBech32AccountAddrPrefix()) {
		// Assume that was provided eth address
		hexAddress := common.HexToAddress(input)
		return hexAddress.Bytes(), nil
	}

	// Assume that was provided bech32 address
	accAddress, err := sdk.AccAddressFromBech32(input)
	if err != nil {
		return nil, err
	}

	return accAddress, nil
}

// ValidateEdDSAPublicKey tries to decompress provided BJJ public key
func ValidateEdDSAPublicKey(input []byte) error {
	if len(input) != 32 {
		return errors.New("invalid public key length. Expected 32 bytes")
	}

	pointBuf := babyjub.NewPoint()
	_, err := pointBuf.Decompress([32]byte(input))
	if err != nil {
		return err
	}

	return nil
}
