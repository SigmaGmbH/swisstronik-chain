package types

import (
	"encoding/binary"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"math/big"
)

func (vt VerificationType) ToBytes() []byte {
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, uint32(vt))
	return bytes
}

func (c *ZKCredential) Hash() (*big.Int, error) {
	typeBig := big.NewInt(int64(c.Type))
	issuerAddressBig := new(big.Int).SetBytes(c.IssuerAddress)
	expirationBig := big.NewInt(int64(c.ExpirationTimestamp))

	// Decompress public key and hash only X-coordinate
	point, err := babyjub.NewPoint().Decompress([32]byte(c.HolderPublicKey))
	if err != nil {
		return nil, err
	}
	holderPublicKeyBig := point.X

	valuesToHash := []*big.Int{typeBig, issuerAddressBig, holderPublicKeyBig, expirationBig}
	return poseidon.Hash(valuesToHash)
}
