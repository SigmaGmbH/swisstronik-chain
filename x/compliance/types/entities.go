package types

import (
	"encoding/binary"
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
	issuanceBig := big.NewInt(int64(c.IssuanceTimestamp))
	holderPublicKeyBig := new(big.Int).SetBytes(c.HolderPublicKey)
	println("DEBUG: Hash x-coord: ", holderPublicKeyBig.String())

	valuesToHash := []*big.Int{typeBig, issuerAddressBig, holderPublicKeyBig, expirationBig, issuanceBig}
	return poseidon.Hash(valuesToHash)
}
