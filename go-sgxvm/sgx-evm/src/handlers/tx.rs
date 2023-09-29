use protobuf::Message;
use primitive_types::{H160, H256, U256};
use std::{vec::Vec, string::String};
use evm::ExitReason;
use internal_types::ExecutionResult;
use protobuf::RepeatedField;
use evm::executor::stack::{MemoryStackState, StackExecutor, StackSubstateMetadata};
use evm::backend::Backend;

use crate::AllocationWithResult;
use crate::encryption::{decrypt_transaction_data, extract_public_key_and_data, ENCRYPTED_DATA_LEN, encrypt_transaction_data};
use crate::protobuf_generated::ffi::{
    AccessListItem, HandleTransactionResponse, Log,
    SGXVMCallRequest, SGXVMCreateRequest, Topic, TransactionContext as ProtoTransactionContext,
};
use crate::backend;
use crate::GoQuerier;
use crate::types::{Vicinity, GASOMETER_CONFIG};

/// Handles incoming request for calling contract or transferring value
pub fn handle_call_request(querier: *mut GoQuerier, data: SGXVMCallRequest) -> AllocationWithResult {
    let res = handle_call_request_inner(querier, data);
    post_transaction_handling(res)
}

/// Handles incoming request for creation of a new contract
pub fn handle_create_request(querier: *mut GoQuerier, data: SGXVMCreateRequest) -> AllocationWithResult {
    let res = handle_create_request_inner(querier, data);
    post_transaction_handling(res)
}

fn post_transaction_handling(execution_result: ExecutionResult) -> AllocationWithResult {
    let mut response = HandleTransactionResponse::new();
    response.set_gas_used(execution_result.gas_used);
    response.set_vm_error(execution_result.vm_error);
    response.set_ret(execution_result.data);

    // Convert logs into proper format
    let converted_logs = execution_result
        .logs
        .into_iter()
        .map(|log| {
            let mut proto_log = Log::new();
            proto_log.set_address(log.address.as_fixed_bytes().to_vec());
            proto_log.set_data(log.data);

            let converted_topics: Vec<Topic> =
                log.topics.into_iter().map(convert_topic_to_proto).collect();
            proto_log.set_topics(converted_topics.into());

            proto_log
        })
        .collect();

    response.set_logs(converted_logs);

    let encoded_response = match response.write_to_bytes() {
        Ok(res) => res,
        Err(err) => {
            println!("Cannot encode protobuf result. Reason: {:?}", err);
            return AllocationWithResult::default();
        }
    };

    super::allocate_inner(encoded_response)
}

fn handle_call_request_inner(querier: *mut GoQuerier, data: SGXVMCallRequest) -> ExecutionResult {
    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let vicinity = Vicinity {
        origin: H160::from_slice(&params.from),
        nonce: U256::from(params.nonce),
    };
    let mut storage = crate::storage::FFIStorage::new(querier);
    let mut backend = backend::FFIBackend::new(
        querier,
        &mut storage,
        vicinity,
        build_transaction_context(context),
    );

    // If data is empty, there should be no encryption of result. Otherwise we should try
    // to extract user public key and encrypted data
    match params.data.len() {
        0 => {
            sgxvm_call(
                &mut backend,
                params.gasLimit,
                H160::from_slice(&params.from),
                H160::from_slice(&params.to),
                U256::from_big_endian(&params.value),
                params.data,
                parse_access_list(params.accessList),
                params.commit,
            )
        },
        _ => {
            // Extract user public key from transaction data
            let (user_public_key, data) = match extract_public_key_and_data(params.data) {
                Ok((user_public_key, data)) => (user_public_key, data),
                Err(err) => {
                    return ExecutionResult::from_error(
                        format!("{:?}", err),
                        Vec::default(),
                        None
                    );
                }
            };

            // If encrypted data presents, decrypt it
            let decrypted_data = if !data.is_empty() {
                match decrypt_transaction_data(data, user_public_key.clone()) {
                    Ok(decrypted_data) => decrypted_data,
                    Err(err) => {
                        return ExecutionResult::from_error(
                            format!("{:?}", err),
                            Vec::default(),
                            None
                        );
                    }
                }
            } else { Vec::default() };

            let mut exec_result = sgxvm_call(
                &mut backend,
                params.gasLimit,
                H160::from_slice(&params.from),
                H160::from_slice(&params.to),
                U256::from_big_endian(&params.value),
                decrypted_data,
                parse_access_list(params.accessList),
                params.commit,
            );

            // Encrypt transaction data output
            let encrypted_data = match encrypt_transaction_data(exec_result.data, user_public_key) {
                Ok(data) => data,
                Err(err) => {
                    return ExecutionResult::from_error(
                        format!("{:?}", err),
                        Vec::default(),
                        None
                    );
                }
            };

            exec_result.data = encrypted_data;
            exec_result
        }
    }
}

fn handle_create_request_inner(querier: *mut GoQuerier, data: SGXVMCreateRequest) -> ExecutionResult {
    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let vicinity = Vicinity {
        origin: H160::from_slice(&params.from),
        nonce: U256::from(params.nonce),
    };
    let mut storage = crate::storage::FFIStorage::new(querier);
    let mut backend = backend::FFIBackend::new(
        querier,
        &mut storage,
        vicinity,
        build_transaction_context(context),
    );

    sgxvm_create(
        &mut backend,
        params.gasLimit,
        H160::from_slice(&params.from),
        U256::from_big_endian(&params.value),
        params.data,
        parse_access_list(params.accessList),
        params.commit,
    )
}

fn parse_access_list(data: RepeatedField<AccessListItem>) -> Vec<(H160, Vec<H256>)> {
    let mut access_list = Vec::default();
    for access_list_item in data.to_vec() {
        let address = H160::from_slice(&access_list_item.address);
        let slots = access_list_item
            .storageSlot
            .to_vec()
            .into_iter()
            .map(|item| H256::from_slice(&item))
            .collect();

        access_list.push((address, slots));
    }

    access_list
}

fn build_transaction_context(context: ProtoTransactionContext) -> backend::TxContext {
    backend::TxContext {
        chain_id: U256::from(context.chain_id),
        gas_price: U256::from_big_endian(&context.gas_price),
        block_number: U256::from(context.block_number),
        timestamp: U256::from(context.timestamp),
        block_gas_limit: U256::from(context.block_gas_limit),
        block_base_fee_per_gas: U256::from_big_endian(&context.block_base_fee_per_gas),
        block_coinbase: H160::from_slice(&context.block_coinbase),
    }
}

fn convert_topic_to_proto(topic: H256) -> Topic {
    let mut protobuf_topic = Topic::new();
    protobuf_topic.set_inner(topic.as_fixed_bytes().to_vec());

    protobuf_topic
}

/// Handles incoming request for calling some contract / funds transfer
fn sgxvm_call(
    backend: &mut impl ExtendedBackend,
    gas_limit: u64,
    from: H160,
    to: H160,
    value: U256,
    data: Vec<u8>,
    access_list: Vec<(H160, Vec<H256>)>,
    commit: bool,
) -> ExecutionResult {
    let metadata = StackSubstateMetadata::new(gas_limit, &GASOMETER_CONFIG);
    let state = MemoryStackState::new(metadata, backend);
    let precompiles = EVMPrecompiles::<Backend>::new();

    let mut executor = StackExecutor::new_with_precompiles(state, &GASOMETER_CONFIG, &precompiles);
    let (exit_reason, ret) = executor.transact_call(from, to, value, data, gas_limit, access_list);

    let gas_used = executor.used_gas();
    let exit_value = match handle_evm_result(exit_reason, ret) {
        Ok(data) => data,
        Err((err, data)) => {
            return ExecutionResult::from_error(err, data, Some(gas_used))
        }
    };

    if commit {
        let (vals, logs) = executor.into_state().deconstruct();
        backend.apply(vals, logs, false);
    }

    ExecutionResult {
        logs: backend.get_logs(),
        data: exit_value,
        gas_used,
        vm_error: "".to_string(),
    }
}

/// Handles incoming request for creation of a new contract
fn sgxvm_create(
    backend: &mut impl ExtendedBackend,
    gas_limit: u64,
    from: H160,
    value: U256,
    data: Vec<u8>,
    access_list: Vec<(H160, Vec<H256>)>,
    commit: bool,
) -> ExecutionResult {
    let metadata = StackSubstateMetadata::new(gas_limit, &GASOMETER_CONFIG);
    let state = MemoryStackState::new(metadata, backend);
    let precompiles = EVMPrecompiles::<Backend>::new();

    let mut executor = StackExecutor::new_with_precompiles(state, &GASOMETER_CONFIG, &precompiles);
    let (exit_reason, ret) = executor.transact_create(from, value, data, gas_limit, access_list);

    let gas_used = executor.used_gas();
    let exit_value = match handle_evm_result(exit_reason, ret) {
        Ok(data) => data,
        Err((err, data)) => {
            return ExecutionResult::from_error(err, data, Some(gas_used))
        }
    };

    if commit {
        let (vals, logs) = executor.into_state().deconstruct();
        backend.apply(vals, logs, false);
    }

    ExecutionResult {
        logs: backend.get_logs(),
        data: exit_value,
        gas_used,
        vm_error: "".to_string(),
    }
}

/// Handles an EVM result to return either a successful result or a (readable) error reason.
fn handle_evm_result(exit_reason: ExitReason, data: Vec<u8>) -> Result<Vec<u8>, (String, Vec<u8>)> {
    match exit_reason {
        ExitReason::Succeed(_) => Ok(data),
        ExitReason::Revert(err) => Err((format!("execution reverted: {:?}", err), data)),
        ExitReason::Error(err) => Err((format!("evm error: {:?}", err), data)),
        ExitReason::Fatal(err) => Err((format!("fatal evm error: {:?}", err), data)),
    }
}
