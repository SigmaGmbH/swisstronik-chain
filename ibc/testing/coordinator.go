// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package ibctesting

import (
	"testing"
	"time"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/stretchr/testify/require"
)

const DefaultFeeAmt = int64(150_000_000_000_000_000) // 0.15 SWTR

var globalStartTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

// NewCoordinator initializes Coordinator with N EVM TestChain's and M Cosmos chains (Simulation Apps)
func NewCoordinator(t *testing.T, nEVMChains, mCosmosChains int) *ibctesting.Coordinator {
	chains := make(map[string]*ibctesting.TestChain)
	coord := &ibctesting.Coordinator{
		T:           t,
		CurrentTime: globalStartTime,
	}

	// setup EVM chains
	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp

	for i := 1; i <= nEVMChains; i++ {
		chainID := ibctesting.GetChainID(i)
		chains[chainID] = NewTestChain(t, coord, chainID)
	}

	// setup Cosmos chains
	ibctesting.DefaultTestingAppInit = ibctesting.SetupTestingApp

	for j := 1 + nEVMChains; j <= nEVMChains+mCosmosChains; j++ {
		chainID := ibctesting.GetChainID(j)
		chains[chainID] = ibctesting.NewTestChain(t, coord, chainID)
	}

	coord.Chains = chains

	return coord
}

// SetupPath constructs a TM client, connection, and channel on both chains provided. It will
// fail if any error occurs. The clientID's, TestConnections, and TestChannels are returned
// for both chains. The channels created are connected to the ibc-transfer application.
func SetupPath(coord *ibctesting.Coordinator, path *Path) {
	SetupConnections(coord, path)

	// channels can also be referenced through the returned connections
	CreateChannels(coord, path)
}

// SetupClientConnections is a helper function to create clients and the appropriate
// connections on both the source and counterparty chain. It assumes the caller does not
// anticipate any errors.
func SetupConnections(coord *ibctesting.Coordinator, path *Path) {
	SetupClients(coord, path)

	CreateConnections(coord, path)
}

// CreateChannel constructs and executes channel handshake messages in order to create
// OPEN channels on chainA and chainB. The function expects the channels to be successfully
// opened otherwise testing will fail.
func CreateChannels(coord *ibctesting.Coordinator, path *Path) {
	err := path.EndpointA.ChanOpenInit()
	require.NoError(coord.T, err)

	err = path.EndpointB.ChanOpenTry()
	require.NoError(coord.T, err)

	err = path.EndpointA.ChanOpenAck()
	require.NoError(coord.T, err)

	err = path.EndpointB.ChanOpenConfirm()
	require.NoError(coord.T, err)

	// ensure counterparty is up to date
	err = path.EndpointA.UpdateClient()
	require.NoError(coord.T, err)
}

// CreateConnection constructs and executes connection handshake messages in order to create
// OPEN channels on chainA and chainB. The connection information of for chainA and chainB
// are returned within a TestConnection struct. The function expects the connections to be
// successfully opened otherwise testing will fail.
func CreateConnections(coord *ibctesting.Coordinator, path *Path) {
	err := path.EndpointA.ConnOpenInit()
	require.NoError(coord.T, err)

	err = path.EndpointB.ConnOpenTry()
	require.NoError(coord.T, err)

	err = path.EndpointA.ConnOpenAck()
	require.NoError(coord.T, err)

	err = path.EndpointB.ConnOpenConfirm()
	require.NoError(coord.T, err)

	// ensure counterparty is up to date
	err = path.EndpointA.UpdateClient()
	require.NoError(coord.T, err)
}

// SetupClients is a helper function to create clients on both chains. It assumes the
// caller does not anticipate any errors.
func SetupClients(coord *ibctesting.Coordinator, path *Path) {
	err := path.EndpointA.CreateClient()
	require.NoError(coord.T, err)

	err = path.EndpointB.CreateClient()
	require.NoError(coord.T, err)
}
