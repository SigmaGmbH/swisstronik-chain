package testutil

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TestSigner struct {
	privateKey 	cryptotypes.PrivKey
	address 	sdk.Address
}

func (ts TestSigner) NewTestSigner(privateKey cryptotypes.PrivKey) keyring.Signer {
	return TestSigner{
		privateKey: privateKey,
		address: sdk.AccAddress(privateKey.PubKey().Address()),
	}
}

func (ts TestSigner) Sign(uid string, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	sig, err := ts.privateKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, ts.privateKey.PubKey(), nil
}

func (ts TestSigner) SignByAddress(address sdk.Address, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	if !ts.address.Equals(address) {
		return nil, nil, fmt.Errorf("cannot find address. Given address: %s, Signer address: %s", address, ts.address)
	}

	return ts.Sign("", msg)
}