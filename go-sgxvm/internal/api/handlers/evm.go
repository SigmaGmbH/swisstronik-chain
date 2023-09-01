package handlers

import "C"
import (
	ffi "github.com/SigmaGmbH/librustgo/go_protobuf_gen"
	"github.com/SigmaGmbH/librustgo/internal/api"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/protobuf/proto"
	"log"
	"runtime"
)

// Call handles incoming call to contract or transfer of value
func Call(
	connector api.Connector,
	from, to, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit uint64,
	txContext *ffi.TransactionContext,
	commit bool,
) (*ffi.HandleTransactionResponse, error) {
	// Construct connector to rust code
	c := api.BuildConnector(connector)

	// Create protobuf-encoded transaction data
	params := &ffi.SGXVMCallParams{
		From:       from,
		To:         to,
		Data:       data,
		GasLimit:   gasLimit,
		Value:      value,
		AccessList: convertAccessList(accessList),
		Commit:     commit,
	}

	// Create protobuf encoded request
	req := ffi.FFIRequest{Req: &ffi.FFIRequest_CallRequest{
		CallRequest: &ffi.SGXVMCallRequest{
			Params:  params,
			Context: txContext,
		},
	}}
	reqBytes, err := proto.Marshal(&req)
	if err != nil {
		log.Fatalln("Failed to encode req:", err)
	}

	// Pass request to Rust
	d := api.MakeView(reqBytes)
	defer runtime.KeepAlive(reqBytes)

	errmsg := api.NewUnmanagedVector(nil)
	ptr, err := C.make_pb_request(c, d, &errmsg)
	if err != nil {
		return &ffi.HandleTransactionResponse{}, api.ErrorWithMessage(err, errmsg)
	}

	// Recover returned value
	executionResult := api.CopyAndDestroyUnmanagedVector(ptr)
	response := ffi.HandleTransactionResponse{}
	if err := proto.Unmarshal(executionResult, &response); err != nil {
		log.Fatalln("Failed to decode execution result:", err)
	}

	return &response, nil
}

// Create handles incoming request for creation of new contract
func Create(
	connector api.Connector,
	from, data, value []byte,
	accessList ethtypes.AccessList,
	gasLimit uint64,
	txContext *ffi.TransactionContext,
	commit bool,
) (*ffi.HandleTransactionResponse, error) {
	// Construct connector to rust code
	c := api.BuildConnector(connector)

	// Create protobuf-encoded transaction data
	params := &ffi.SGXVMCreateParams{
		From:       from,
		Data:       data,
		GasLimit:   gasLimit,
		Value:      value,
		AccessList: convertAccessList(accessList),
		Commit:     commit,
	}

	// Create protobuf encoded request
	req := ffi.FFIRequest{Req: &ffi.FFIRequest_CreateRequest{
		CreateRequest: &ffi.SGXVMCreateRequest{
			Params:  params,
			Context: txContext,
		},
	}}
	reqBytes, err := proto.Marshal(&req)
	if err != nil {
		log.Fatalln("Failed to encode req:", err)
	}

	// Pass request to Rust
	d := api.MakeView(reqBytes)
	defer runtime.KeepAlive(reqBytes)

	errmsg := api.NewUnmanagedVector(nil)
	ptr, err := C.make_pb_request(c, d, &errmsg)
	if err != nil {
		return &ffi.HandleTransactionResponse{}, api.ErrorWithMessage(err, errmsg)
	}

	// Recover returned value
	executionResult := api.CopyAndDestroyUnmanagedVector(ptr)
	response := ffi.HandleTransactionResponse{}
	if err := proto.Unmarshal(executionResult, &response); err != nil {
		log.Fatalln("Failed to decode execution result:", err)
	}

	return &response, nil
}

// Converts AccessList type from ethtypes to protobuf-compatible type
func convertAccessList(accessList ethtypes.AccessList) []*ffi.AccessListItem {
	var converted []*ffi.AccessListItem
	for _, item := range accessList {
		accessListItem := &ffi.AccessListItem{
			StorageSlot: convertAccessListStorageSlots(item.StorageKeys),
			Address:     item.Address.Bytes(),
		}

		converted = append(converted, accessListItem)
	}
	return converted
}

// Converts storage slots of access list in [][]byte format
func convertAccessListStorageSlots(slots []ethcommon.Hash) [][]byte {
	var converted [][]byte
	for _, slot := range slots {
		converted = append(converted, slot.Bytes())
	}
	return converted
}
