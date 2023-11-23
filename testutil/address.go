package testutil

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"swisstronik/crypto/ethsecp256k1"
)

func RandomEthAddressWithPrivateKey() (common.Address, *ethsecp256k1.PrivKey) {
	privateKey, err := ethsecp256k1.GenerateKey()
	if err != nil {
		return common.Address{}, nil
	}

	ecdsaPrivateKey, err := privateKey.ToECDSA()
	if err != nil {
		return common.Address{}, nil
	}

	addr := crypto.PubkeyToAddress(ecdsaPrivateKey.PublicKey)
	return addr, privateKey
}

func RandomEthAddress() common.Address {
	address, _ := RandomEthAddressWithPrivateKey()
	return address
}

