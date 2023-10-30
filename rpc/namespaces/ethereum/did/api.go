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
package did

import (
	"github.com/cosmos/cosmos-sdk/server"

	rpctypes "swisstronik/rpc/types"

	"github.com/cometbft/cometbft/libs/log"

	"swisstronik/rpc/backend"
	didtypes "swisstronik/x/did/types"
)

// API is the private miner prefixed set of APIs in the Miner JSON-RPC spec.
type API struct {
	ctx     *server.Context
	logger  log.Logger
	backend backend.DIDBackend
}

// NewPrivateAPI creates an instance of the DID API.
func NewPrivateAPI(
	ctx *server.Context,
	backend backend.DIDBackend,
) *API {
	return &API{
		ctx:     ctx,
		logger:  ctx.Logger.With("api", "did"),
		backend: backend,
	}
}

// Resolve function gets did document for id
func (api *API) Resolve(blockNrOrHash rpctypes.BlockNumberOrHash, Id string) (*didtypes.DIDDocumentWithMetadata, error) {
	api.logger.Debug("did_resolve", "block number or hash", blockNrOrHash, "DID Id", Id)
	return api.backend.DIDResolve(blockNrOrHash, Id)
}
