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
	"cosmossdk.io/simapp/params"
	amino "github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"

	enccodec "swisstronik/encoding/codec"

	"cosmossdk.io/x/tx/signing"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogoproto "github.com/cosmos/gogoproto/proto"
)

// MakeConfig creates an EncodingConfig for testing
func MakeConfig(mb module.BasicManager) params.EncodingConfig {
	cdc := amino.NewLegacyAmino()
	interfaceRegistry, err := NewInterfaceRegistry(
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		sdk.GetConfig().GetBech32ValidatorAddrPrefix(),
	)

	if err != nil {
		panic(err)
	}
	codec := amino.NewProtoCodec(interfaceRegistry)

	encodingConfig := params.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             codec,
		TxConfig:          tx.NewTxConfig(codec, tx.DefaultSignModes),
		Amino:             cdc,
	}

	enccodec.RegisterLegacyAminoCodec(encodingConfig.Amino)
	mb.RegisterLegacyAminoCodec(encodingConfig.Amino)
	enccodec.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	mb.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	return encodingConfig
}

func NewInterfaceRegistry(addrPrefix string, valAddrPrefix string) (types.InterfaceRegistry, error) {
	return types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles: gogoproto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          addresscodec.NewBech32Codec(addrPrefix),
			ValidatorAddressCodec: addresscodec.NewBech32Codec(valAddrPrefix),
			// TODO(CORE-840): cosmos.msg.v1.signer annotation doesn't supported nested messages beyond a depth of 1
			// which requires any message where the authority is nested further to implement its own accessor. Once
			// https://github.com/cosmos/cosmos-sdk/issues/18722 is fixed, replace this with the cosmos.msg.v1.signing
			// annotation on the protos.
		},
	})
}
