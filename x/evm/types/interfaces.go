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
package types

import (
	"github.com/SigmaGmbH/go-merkletree-sql/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/core"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"math/big"

	compliancetypes "swisstronik/x/compliance/types"
	feemarkettypes "swisstronik/x/feemarket/types"
)

// AccountKeeper defines the expected account keeper interface
type AccountKeeper interface {
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
	GetAllAccounts(ctx sdk.Context) (accounts []authtypes.AccountI)
	IterateAccounts(ctx sdk.Context, cb func(account authtypes.AccountI) bool)
	GetSequence(sdk.Context, sdk.AccAddress) (uint64, error)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	SetAccount(ctx sdk.Context, account authtypes.AccountI)
	RemoveAccount(ctx sdk.Context, account authtypes.AccountI)
	GetParams(ctx sdk.Context) (params authtypes.Params)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	authtypes.BankKeeper
	SpendableCoin(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

// StakingKeeper returns the historical headers kept in store.
type StakingKeeper interface {
	GetHistoricalInfo(ctx sdk.Context, height int64) (stakingtypes.HistoricalInfo, bool)
	GetValidatorByConsAddr(ctx sdk.Context, consAddr sdk.ConsAddress) (validator stakingtypes.Validator, found bool)
}

// FeeMarketKeeper
type FeeMarketKeeper interface {
	GetBaseFee(ctx sdk.Context) *big.Int
	GetParams(ctx sdk.Context) feemarkettypes.Params
	AddTransientGasWanted(ctx sdk.Context, gasWanted uint64) (uint64, error)
}

// ComplianceKeeper
type ComplianceKeeper interface {
	AddVerificationDetails(ctx sdk.Context, userAddress sdk.AccAddress, verificationType compliancetypes.VerificationType, details *compliancetypes.VerificationDetails) ([]byte, error)
	AddVerificationDetailsV2(ctx sdk.Context, userAddress sdk.AccAddress, verificationType compliancetypes.VerificationType, details *compliancetypes.VerificationDetails, userPublicKey []byte) ([]byte, error)
	HasVerificationOfType(ctx sdk.Context, userAddress sdk.AccAddress, expectedType compliancetypes.VerificationType, expirationTimestamp uint32, expectedIssuers []sdk.AccAddress) (bool, error)
	GetVerificationDetailsByIssuer(ctx sdk.Context, userAddress, issuerAddress sdk.AccAddress) ([]*compliancetypes.Verification, []*compliancetypes.VerificationDetails, error)
	GetIssuanceTreeRoot(ctx sdk.Context) (*big.Int, error)
	GetRevocationTreeRoot(ctx sdk.Context) (*big.Int, error)
	SetTreeRoot(context sdk.Context, treeKey []byte, root *merkletree.Hash) error
	IsVerificationRevoked(ctx sdk.Context, verificationId []byte) (bool, error)
	RevokeVerification(ctx sdk.Context, verificationDetailsId []byte, issuerAddress sdk.AccAddress) error
	ConvertCredential(ctx sdk.Context, verificationId []byte, publicKeyToSet []byte, caller sdk.AccAddress) error
}

// Event Hooks
// These can be utilized to customize evm transaction processing.

// EvmHooks event hooks for evm tx processing
type EvmHooks interface {
	// Must be called after tx is processed successfully, if return an error, the whole transaction is reverted.
	PostTxProcessing(ctx sdk.Context, msg core.Message, receipt *ethtypes.Receipt) error
}

type (
	LegacyParams = paramtypes.ParamSet
	// Subspace defines an interface that implements the legacy Cosmos SDK x/params Subspace type.
	// NOTE: This is used solely for migration of the Cosmos SDK x/params managed parameters.
	Subspace interface {
		GetParamSetIfExists(ctx sdk.Context, ps LegacyParams)
	}
)
