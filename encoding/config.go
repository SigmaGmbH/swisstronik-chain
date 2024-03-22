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
package encoding

import (
	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/protobuf/proto"

	enccodec "swisstronik/encoding/codec"

	ethermint "swisstronik/types"

	"cosmossdk.io/x/tx/signing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	gogoproto "github.com/cosmos/gogoproto/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var InterfaceRegistry types.InterfaceRegistry

func customGetSignerFn(path string) func(msg proto.Message) ([][]byte, error) {
	return func(msg proto.Message) ([][]byte, error) {
		m := msg.ProtoReflect()
		fieldDesc := m.Descriptor().Fields().ByName(protoreflect.Name(path))
		addr := common.BytesToAddress((m.Get(fieldDesc).Bytes()))
		signer := sdk.AccAddress(addr.Bytes())
		return [][]byte{signer}, nil
	}
}

// MakeConfig creates an EncodingConfig for testing
func MakeConfig(mb module.BasicManager) ethermint.EncodingConfig {
	cdc := amino.NewLegacyAmino()
	signingOptions := signing.Options{
		AddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		ValidatorAddressCodec: address.Bech32Codec{
			Bech32Prefix: sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
		},
		CustomGetSigners: map[protoreflect.FullName]signing.GetSignersFunc{
			"ethermint.evm.v1.MsgHandleTx": customGetSignerFn("from"),
		},
	}
	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     gogoproto.HybridResolver,
		SigningOptions: signingOptions,
	})

	if err != nil {
		panic(err)
	}
	codec := amino.NewProtoCodec(interfaceRegistry)

	encodingConfig := ethermint.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             cdc,
	}

	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	// mb.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	// mb.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// This is needed for the EIP712 txs because currently is using
	// the deprecated method legacytx.StdSignBytes
	legacytx.RegressionTestingAminoCodec = cdc
	return encodingConfig
}
