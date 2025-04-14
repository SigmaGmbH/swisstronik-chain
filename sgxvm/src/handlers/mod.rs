use std::vec::Vec;
use std::slice;
use sgx_types::sgx_status_t;
use protobuf::Message;

use crate::protobuf_generated::ffi::{FFIRequest, FFIRequest_oneof_req, SGXVMEstimateGasRequest};
use crate::{AllocationWithResult, Allocation};
use crate::ocall;
use crate::key_manager::KeyManager;
use crate::protobuf_generated::ffi::{
    NodePublicKeyResponse,
    SGXVMCallRequest, 
    SGXVMCreateRequest,
};
use crate::GoQuerier;

pub mod tx;
mod utils;

/// Handles incoming request for calling contract or transferring value
/// * querier - GoQuerier which is used to interact with Go (Cosmos) from SGX Enclave
/// * data - EVM call data (destination, value, etc.)
pub fn handle_evm_call_request(querier: *mut GoQuerier, data: SGXVMCallRequest) -> AllocationWithResult {
    let res = tx::handle_call_request_inner(querier, data);
    tx::convert_and_allocate_transaction_result(res)
}

/// Handles incoming request for creation of a new contract
/// * querier - GoQuerier which is used to interact with Go (Cosmos) from SGX Enclave
/// * data - EVM call data (value, tx.data, etc.)
pub fn handle_evm_create_request(querier: *mut GoQuerier, data: SGXVMCreateRequest) -> AllocationWithResult {
    let res = tx::handle_create_request_inner(querier, data);
    tx::convert_and_allocate_transaction_result(res)
}

pub fn handle_evm_estimate_gas_request(querier: *mut GoQuerier, data: SGXVMEstimateGasRequest) -> AllocationWithResult {
    let res = tx::handle_estimate_gas_request_inner(querier, data);
    tx::convert_and_allocate_transaction_result(res)
}