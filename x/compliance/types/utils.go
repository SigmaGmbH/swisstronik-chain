package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"math/big"
	"slices"
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

// ExtractXCoordinate tries to extract X-coordinate from provided BJJ public key
func ExtractXCoordinate(compressedPublicKeyBytes []byte, isLittleEndian bool) (*big.Int, error) {
	if len(compressedPublicKeyBytes) != 32 {
		return nil, fmt.Errorf("invalid compressed public key bytes. Got length: %d", len(compressedPublicKeyBytes))
	}

	if !isLittleEndian {
		// Convert to little endian
		slices.Reverse(compressedPublicKeyBytes)
	}

	pointBuf, err := babyjub.NewPoint().Decompress([32]byte(compressedPublicKeyBytes))
	if err != nil {
		return nil, err
	}

	return pointBuf.X, nil
}
