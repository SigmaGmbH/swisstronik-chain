// Copyright 2021 Evmos Foundation
// This file is part of Evmos' Ethermint library.
//
// The Ethermint library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Ethermint library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Ethermint library. If not, see https://github.com/evmos/ethermint/blob/main/LICENSE
package tests

import (
	"fmt"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"swisstronik/crypto/ethsecp256k1"
)

// RandomEthAddressWithPrivateKey generates an Ethereum address and its corresponding private key.
func RandomEthAddressWithPrivateKey() (common.Address, cryptotypes.PrivKey) {
	privkey, _ := ethsecp256k1.GenerateKey()
	key, err := privkey.ToECDSA()
	if err != nil {
		return common.Address{}, nil
	}

	addr := crypto.PubkeyToAddress(key.PublicKey)

	return addr, privkey
}

// RandomEthAddress generates an Ethereum address.
func RandomEthAddress() common.Address {
	addr, _ := RandomEthAddressWithPrivateKey()
	return addr
}

// RandomAccAddress generates Cosmos SDK address
func RandomAccAddress() sdk.AccAddress {
	addr, _ := RandomEthAddressWithPrivateKey()
	return sdk.AccAddress(addr.Bytes())
}

func RandomSimulationEthAccount() simulation.Account {
	privKey, _ := ethsecp256k1.GenerateKey()
	key, _ := privKey.ToECDSA()
	return simulation.Account{
		PrivKey: privKey,
		PubKey:  privKey.PubKey(),
		Address: sdk.AccAddress(crypto.PubkeyToAddress(key.PublicKey).Bytes()),
		ConsKey: ed25519.GenPrivKey(),
	}
}

func RandomSimulationEthAccounts(r *rand.Rand, n int) []simulation.Account {
	accs := make([]simulation.Account, n)
	for i := 0; i < n; i++ {
		accs[i] = RandomSimulationEthAccount()
	}
	return accs
}

var _ keyring.Signer = &Signer{}

// Signer defines a type that is used on testing for signing MsgEthereumTx
type Signer struct {
	privKey cryptotypes.PrivKey
}

func NewTestSigner(sk cryptotypes.PrivKey) keyring.Signer {
	return &Signer{
		privKey: sk,
	}
}

// Sign signs the message using the underlying private key
func (s Signer) Sign(_ string, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	if s.privKey.Type() != ethsecp256k1.KeyType {
		return nil, nil, fmt.Errorf(
			"invalid private key type for signing ethereum tx; expected %s, got %s",
			ethsecp256k1.KeyType,
			s.privKey.Type(),
		)
	}

	sig, err := s.privKey.Sign(msg)
	if err != nil {
		return nil, nil, err
	}

	return sig, s.privKey.PubKey(), nil
}

// SignByAddress sign byte messages with a user key providing the address.
func (s Signer) SignByAddress(address sdk.Address, msg []byte) ([]byte, cryptotypes.PubKey, error) {
	signer := sdk.AccAddress(s.privKey.PubKey().Address())
	if !signer.Equals(address) {
		return nil, nil, fmt.Errorf("address mismatch: signer %s ≠ given address %s", signer, address)
	}

	return s.Sign("", msg)
}
