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
package backend

import (
	"fmt"
	"math"
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"

	rpctypes "swisstronik/rpc/types"
	compliancetypes "swisstronik/x/compliance/types"
	evmtypes "swisstronik/x/evm/types"
)

// GetCode returns the contract code at the given address and block number.
func (b *Backend) GetCode(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryCodeRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Code(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	return res.Code, nil
}

// GetProof returns an account object with proof and any storage proofs
func (b *Backend) GetProof(address common.Address, storageKeys []string, blockNrOrHash rpctypes.BlockNumberOrHash) (*rpctypes.AccountResult, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	height := blockNum.Int64()
	_, err = b.TendermintBlockByNumber(blockNum)
	if err != nil {
		// the error message imitates geth behavior
		return nil, errors.New("header not found")
	}
	ctx := rpctypes.ContextWithHeight(height)

	// if the height is equal to zero, meaning the query condition of the block is either "pending" or "latest"
	if height == 0 {
		bn, err := b.BlockNumber()
		if err != nil {
			return nil, err
		}

		if bn > math.MaxInt64 {
			return nil, fmt.Errorf("not able to query block number greater than MaxInt64")
		}

		height = int64(bn)
	}

	clientCtx := b.clientCtx.WithHeight(height)

	// query storage proofs
	storageProofs := make([]rpctypes.StorageResult, len(storageKeys))

	for i, key := range storageKeys {
		hexKey := common.HexToHash(key)
		valueBz, proof, err := b.queryClient.GetProof(clientCtx, evmtypes.StoreKey, evmtypes.StateKey(address, hexKey.Bytes()))
		if err != nil {
			return nil, err
		}

		storageProofs[i] = rpctypes.StorageResult{
			Key:   key,
			Value: (*hexutil.Big)(new(big.Int).SetBytes(valueBz)),
			Proof: GetHexProofs(proof),
		}
	}

	// query EVM account
	req := &evmtypes.QueryAccountRequest{
		Address: address.String(),
	}

	res, err := b.queryClient.Account(ctx, req)
	if err != nil {
		return nil, err
	}

	// query account proofs
	accountKey := authtypes.AddressStoreKey(sdk.AccAddress(address.Bytes()))
	_, proof, err := b.queryClient.GetProof(clientCtx, authtypes.StoreKey, accountKey)
	if err != nil {
		return nil, err
	}

	balance, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	return &rpctypes.AccountResult{
		Address:      address,
		AccountProof: GetHexProofs(proof),
		Balance:      (*hexutil.Big)(balance.BigInt()),
		CodeHash:     common.HexToHash(res.CodeHash),
		Nonce:        hexutil.Uint64(res.Nonce),
		StorageHash:  common.Hash{}, // NOTE: Ethermint doesn't have a storage hash. TODO: implement?
		StorageProof: storageProofs,
	}, nil
}

// GetStorageAt returns the contract storage at the given address, block number, and key.
func (b *Backend) GetStorageAt(address common.Address, key string, blockNrOrHash rpctypes.BlockNumberOrHash) (hexutil.Bytes, error) {
	return nil, errors.New("eth_getStorageAt was disabled, since storage is encrypted. Check docs at https://swisstronik.gitbook.io/swisstronik-docs/ for more information")
}

// GetBalance returns the provided account's balance up to the provided block number.
func (b *Backend) GetBalance(address common.Address, blockNrOrHash rpctypes.BlockNumberOrHash) (*hexutil.Big, error) {
	blockNum, err := b.BlockNumberFromTendermint(blockNrOrHash)
	if err != nil {
		return nil, err
	}

	req := &evmtypes.QueryBalanceRequest{
		Address: address.String(),
	}

	_, err = b.TendermintBlockByNumber(blockNum)
	if err != nil {
		return nil, err
	}

	res, err := b.queryClient.Balance(rpctypes.ContextWithHeight(blockNum.Int64()), req)
	if err != nil {
		return nil, err
	}

	val, ok := sdkmath.NewIntFromString(res.Balance)
	if !ok {
		return nil, errors.New("invalid balance")
	}

	// balance can only be negative in case of pruned node
	if val.IsNegative() {
		return nil, errors.New("couldn't fetch balance. Node state is pruned")
	}

	return (*hexutil.Big)(val.BigInt()), nil
}

// GetTransactionCount returns the number of transactions at the given address up to the given block number.
func (b *Backend) GetTransactionCount(address common.Address, blockNum rpctypes.BlockNumber) (*hexutil.Uint64, error) {
	n := hexutil.Uint64(0)
	bn, err := b.BlockNumber()
	if err != nil {
		return &n, err
	}
	height := blockNum.Int64()

	currentHeight := int64(bn) //#nosec G701 -- checked for int overflow already
	if height > currentHeight {
		return &n, errorsmod.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"cannot query with height in the future (current: %d, queried: %d); please provide a valid height",
			currentHeight, height,
		)
	}

	// Get nonce (sequence) from account
	from := sdk.AccAddress(address.Bytes())
	accRet := b.clientCtx.AccountRetriever

	err = accRet.EnsureExists(b.clientCtx, from)
	if err != nil {
		// account doesn't exist yet, return 0
		return &n, nil
	}

	includePending := blockNum == rpctypes.EthPendingBlockNumber
	nonce, err := b.getAccountNonce(address, includePending, blockNum.Int64(), b.logger)
	if err != nil {
		return nil, err
	}

	n = hexutil.Uint64(nonce)
	return &n, nil
}

// GetNodePublicKey returns x25519 based public key in hex
func (b *Backend) GetNodePublicKey(blockNum rpctypes.BlockNumber) (string, error) {
	// If provided latest block num, check exact value
	var epochBlockNum = blockNum.Int64()
	if epochBlockNum == rpctypes.EthLatestBlockNumber.Int64() {
		blockNumberRes, err := b.BlockNumber()
		if err != nil {
			return "", err
		}

		epochBlockNum = int64(blockNumberRes)
	}

	req := &evmtypes.QueryNodePublicKey{BlockNumber: uint64(epochBlockNum)}
	res, err := b.queryClient.NodePublicKey(rpctypes.ContextWithHeight(0), req)
	if err != nil {
		return "", err
	}

	return res.NodePublicKey, nil
}

func (b *Backend) GetIssuanceProof(credentialHash hexutil.Bytes) (string, error) {
	req := &compliancetypes.QueryIssuanceProofRequest{
		CredentialHash: credentialHash,
	}
	res, err := b.queryClient.ComplianceClient.IssuanceProof(rpctypes.ContextWithHeight(0), req)
	if err != nil {
		return "", err
	}

	return string(res.EncodedProof), nil
}

func (b *Backend) GetNonRevocationProof(credentialHash hexutil.Bytes) (string, error) {
	req := &compliancetypes.QueryRevocationProofRequest{
		CredentialHash: credentialHash,
	}
	res, err := b.queryClient.ComplianceClient.RevocationProof(rpctypes.ContextWithHeight(0), req)
	if err != nil {
		return "", err
	}

	return string(res.EncodedProof), nil
}

func (b *Backend) GetCredentialHash(verificationId hexutil.Bytes) (hexutil.Bytes, error) {
	req := &compliancetypes.QueryCredentialHashRequest{
		VerificationId: verificationId,
	}
	res, err := b.queryClient.ComplianceClient.CredentialHash(rpctypes.ContextWithHeight(0), req)
	if err != nil {
		return nil, err
	}

	return res.CredentialHash, nil
}
