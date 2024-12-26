package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"math/big"
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
func ExtractXCoordinate(compressedPublicKeyBytes []byte) (*big.Int, error) {
	compressedPubKeyBig := new(big.Int).SetBytes(compressedPublicKeyBytes)
	println("DEBUG: Restored compressed public key: ", compressedPubKeyBig.String())
	decodedBytes := bigIntToLittleEndianBytes(compressedPubKeyBig)
	pointBuf, err := babyjub.NewPoint().Decompress([32]byte(decodedBytes))
	if err != nil {
		return nil, err
	}

	return pointBuf.X, nil
}

func bigIntToLittleEndianBytes(n *big.Int) []byte {
	if n.Sign() == 0 {
		return []byte{0}
	}

	bytes := n.Bytes()
	for i, j := 0, len(bytes)-1; i < j; i, j = i+1, j-1 {
		bytes[i], bytes[j] = bytes[j], bytes[i]
	}
	return bytes
}
