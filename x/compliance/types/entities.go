package types

import (
	"encoding/binary"
	"errors"
	"github.com/iden3/go-iden3-crypto/mimc7"
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

	valuesToHash := []*big.Int{typeBig, issuerAddressBig, holderPublicKeyBig, expirationBig, issuanceBig}
	return mimc7.Hash(valuesToHash, big.NewInt(0))
}

func (vd *VerificationDetails) ValidateSize() error {
	if len(vd.OriginChain) > MaxOriginChainSize {
		return errors.New("origin chain too long")
	}

	if len(vd.IssuerVerificationId) > MaxIssuerVerificationIdSize {
		return errors.New("issuer verification id too long")
	}

	if len(vd.OriginalData) > MaxProofDataSize {
		return errors.New("original data too long")
	}

	if len(vd.Schema) > MaxSchemaSize {
		return errors.New("schema too long")
	}

	return nil
}
