//go:build nosgx
// +build nosgx

package api

// #include <stdlib.h>
// #include "bindings.h"
import "C"

import (
	"net"
	"github.com/SigmaGmbH/librustgo/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

// Value types
type (
	cint   = C.int
	cbool  = C.bool
	cusize = C.size_t
	cu8    = C.uint8_t
	cu32   = C.uint32_t
	cu64   = C.uint64_t
	ci8    = C.int8_t
	ci32   = C.int32_t
	ci64   = C.int64_t
)

// Pointers
type cu8_ptr = *C.uint8_t

// Connector is our custom connector
type Connector = types.Connector

// IsNodeInitialized checks if node was initialized and master key was sealed
func IsNodeInitialized() (bool, error) {
	return false, nil
}

// SetupSeedNode handles initialization of seed node which will share seed with other nodes
func InitializeMasterKey(shouldReset bool) error {
	return nil
}

// StartSeedServer handles initialization of seed server
func StartSeedServer(addr string) error {
	return nil
}

func attestPeer(connection net.Conn) error {
	return nil
}

// RequestSeed handles request of seed from seed server
func RequestSeed(hostname string, port int) error {
	return nil
}

// GetNodePublicKey handles request for node public key
func GetNodePublicKey() (*types.NodePublicKeyResponse, error) {
	return nil, nil
}

// Call handles incoming call to contract or transfer of value
func Call(
	connector Connector,
	from, to, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit, nonce uint64,
	txContext *types.TransactionContext,
	commit bool,
) (*types.HandleTransactionResponse, error) {
	return nil, nil
}

// Create handles incoming request for creation of new contract
func Create(
	connector Connector,
	from, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit, nonce uint64,
	txContext *types.TransactionContext,
	commit bool,
) (*types.HandleTransactionResponse, error) {
	return nil, nil
}
