use alloc::collections::BTreeSet;
use evm::backend::OverlayedBackend;
use evm::standard::{Etable, EtableResolver, Invoker, TransactArgs};
use primitive_types::{H160, H256, U256};
use protobuf::Message;
use protobuf::RepeatedField;
use std::vec::Vec;

use crate::encryption::{
    decrypt_transaction_data, encrypt_transaction_data, extract_public_key_and_data,
};
use crate::key_manager::utils::random_nonce;
use crate::precompiles::EVMPrecompiles;
use crate::protobuf_generated::ffi::{AccessListItem, HandleTransactionResponse, Log, SGXVMCallParams, SGXVMCallRequest, SGXVMCreateParams, SGXVMCreateRequest, Topic};
use crate::std::string::ToString;
use crate::types::{ExecutionResult, GASOMETER_CONFIG};
use crate::AllocationWithResult;
use crate::GoQuerier;
use crate::updated_backend::{TxEnvironment, UpdatedBackend};

/// Converts raw execution result into protobuf and returns it outside of enclave
pub fn convert_and_allocate_transaction_result(
    execution_result: ExecutionResult,
) -> AllocationWithResult {
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

/// Inner handler for EVM call request
pub fn handle_call_request_inner(
    querier: *mut GoQuerier,
    data: SGXVMCallRequest,
) -> ExecutionResult {
    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let timestamp = context.timestamp;
    let block_number = context.block_number;
    let gas_price = U256::from_big_endian(&context.gas_price);

    let initial_accessed = {
        let mut hots = BTreeSet::new();
        for item in params.accessList.to_vec() {
            for slot in item.storageSlot.to_vec() {
                hots.insert((H160::from_slice(&item.address), Some(H256::from_slice(&slot))));
            }
        }
        hots
    };

    let tx_environment = TxEnvironment::from(context);
    let mut storage = crate::storage::FFIStorage::new(querier, timestamp, block_number);
    let base_backend = UpdatedBackend::new(querier, &mut storage, tx_environment);
    let mut backend = OverlayedBackend::new(base_backend, initial_accessed);
    let should_commit = params.commit;

    // We do not decrypt transaction if no tx.data provided, or provided explicit flag, that transaction is unencrypted
    let is_encrypted = params.data.len() != 0 && !params.unencrypted;
    match is_encrypted {
        false => {
            let tx_args = construct_call_args(params.clone(), gas_price, params.data);

            run_tx(
                querier,
                &mut backend,
                tx_args,
                should_commit,
            );

            // TODO: Fulfill
            ExecutionResult {
                logs: vec![],
                data: vec![],
                gas_used: 0,
                vm_error: "".to_string(),
            }
        },
        true => {
            // Extract user public key and nonce from transaction data
            let (user_public_key, data, nonce) = match extract_public_key_and_data(params.data.clone()) {
                Ok((user_public_key, data, nonce)) => (user_public_key, data, nonce),
                Err(err) => {
                    return ExecutionResult::from_error(format!("{:?}", err), Vec::default(), None);
                }
            };

            // If encrypted data presents, decrypt it
            let decrypted_data = if !data.is_empty() {
                match decrypt_transaction_data(data, user_public_key.clone(), block_number) {
                    Ok(decrypted_data) => decrypted_data,
                    Err(err) => {
                        return ExecutionResult::from_error(
                            format!("{:?}", err),
                            Vec::default(),
                            None,
                        );
                    }
                }
            } else {
                Vec::default()
            };

            let tx_args = construct_call_args(params.clone(), gas_price, decrypted_data);

            run_tx(
                querier,
                &mut backend,
                tx_args,
                should_commit,
            );

            let mut exec_result= ExecutionResult {
                logs: vec![],
                data: vec![],
                gas_used: 0,
                vm_error: "".to_string(),
            };

            // let mut exec_result = execute_call(
            //     querier,
            //     &mut backend,
            //     params.gasLimit,
            //     H160::from_slice(&params.from),
            //     H160::from_slice(&params.to),
            //     U256::from_big_endian(&params.value),
            //     decrypted_data,
            //     parse_access_list(params.accessList),
            //     params.commit,
            // );

            // If there is transaction with no incoming transaction data, use random nonce to encrypt output
            let nonce = if nonce.is_empty() {
                match random_nonce() {
                    Ok(nonce) => nonce.to_vec(),
                    Err(err) => {
                        return ExecutionResult::from_error(
                            format!("{:?}", err),
                            Vec::default(),
                            None,
                        );
                    } 
                }
            } else {
                nonce
            };

            // Return unencrypted transaction response in case of revert
            if !exec_result.vm_error.is_empty() {
                return exec_result;
            }

            // Encrypt transaction data output
            let encrypted_data =
                match encrypt_transaction_data(exec_result.data, user_public_key, nonce, block_number) {
                    Ok(data) => data,
                    Err(err) => {
                        return ExecutionResult::from_error(
                            format!("{:?}", err),
                            Vec::default(),
                            None,
                        );
                    }
                };

            exec_result.data = encrypted_data;
            exec_result
        }
    }
}

/// Inner handler for EVM create request
pub fn handle_create_request_inner(
    querier: *mut GoQuerier,
    data: SGXVMCreateRequest,
) -> ExecutionResult {
    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let block_number = context.block_number;
    let timestamp = context.timestamp;
    let gas_price = U256::from_big_endian(&context.gas_price);
    let should_commit = params.commit;

    let initial_accessed = {
        let mut hots = BTreeSet::new();
        for item in params.accessList.to_vec() {
            for slot in item.storageSlot.to_vec() {
                hots.insert((H160::from_slice(&item.address), Some(H256::from_slice(&slot))));
            }
        }
        hots
    };

    let mut storage = crate::storage::FFIStorage::new(querier, timestamp, block_number);
    let base_backend = UpdatedBackend::new(querier, &mut storage, TxEnvironment::from(context));
    let mut backend = OverlayedBackend::new(base_backend, initial_accessed);


    let tx_args = construct_create_args(params, gas_price);
    run_tx(
        querier,
        &mut backend,
        tx_args,
        should_commit,
    );

    // TODO: Fill
    ExecutionResult {
        logs: vec![],
        data: vec![],
        gas_used: 0,
        vm_error: "".to_string(),
    }
}

/// Converts access list from RepeatedField to Vec
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

/// Converts EVM topic into protobuf-generated `Topic
fn convert_topic_to_proto(topic: H256) -> Topic {
    let mut protobuf_topic = Topic::new();
    protobuf_topic.set_inner(topic.as_fixed_bytes().to_vec());

    protobuf_topic
}

fn construct_call_args(params: SGXVMCallParams, gas_price: U256, data: Vec<u8>) -> TransactArgs {
    TransactArgs::Call {
        caller: H160::from_slice(&params.from),
        address: H160::from_slice(&params.to),
        value: U256::from_big_endian(&params.value),
        data,
        gas_limit: U256::from(params.gasLimit),
        gas_price: gas_price,
        access_list: parse_access_list(params.accessList),
    }
}

fn construct_create_args(params: SGXVMCreateParams, gas_price: U256) -> TransactArgs {
    TransactArgs::Create {
        caller: H160::from_slice(&params.from),
        value: U256::from_big_endian(&params.value),
        init_code: params.data,
        salt: None,
        gas_limit: U256::from(params.gasLimit),
        gas_price: gas_price,
        access_list: parse_access_list(params.accessList),
    }
}

fn run_tx(
    querier: *mut GoQuerier,
    backend: &mut OverlayedBackend<UpdatedBackend>,
    args: TransactArgs,
    commit: bool
) {
    let gas_etable = Etable::single(evm::standard::eval_gasometer);
    let exec_etable = Etable::runtime();
    let etable = (gas_etable, exec_etable);
    let precompiles = EVMPrecompiles::new(querier);
    let resolver = EtableResolver::new(&GASOMETER_CONFIG, &precompiles, &etable);
    let invoker = Invoker::new(&GASOMETER_CONFIG, &resolver);

    let result = evm::transact(args, None, backend, &invoker);
}

// /// Executes call to smart contract or transferring value
// /// * querier - GoQuerier which is used to interact with Go (Cosmos) from SGX Enclave
// /// * backend - EVM backend for reading and writting state
// /// * gas_limit - gas limit for transaction
// /// * from - transaction origin address
// /// * to - destination address
// /// * data - encoded params for smart contract call or arbitrary data
// /// * access_list - EIP-2930 access list
// /// * commit - should apply changes. Provide `false` if you want to simulate transaction, without state changes
// ///
// /// Returns EVM execution result
// fn execute_call(
//     querier: *mut GoQuerier,
//     backend: &mut impl ExtendedBackend,
//     gas_limit: u64,
//     from: H160,
//     to: H160,
//     value: U256,
//     data: Vec<u8>,
//     access_list: Vec<(H160, Vec<H256>)>,
//     commit: bool,
// ) -> ExecutionResult {
//     let metadata = StackSubstateMetadata::new(gas_limit, &GASOMETER_CONFIG);
//     let state = MemoryStackState::new(metadata, backend);
//     let precompiles = EVMPrecompiles::new(querier);
//
//     let mut executor = StackExecutor::new_with_precompiles(state, &GASOMETER_CONFIG, &precompiles);
//     let (exit_reason, ret) = executor.transact_call(from, to, value, data, gas_limit, access_list);
//
//     let gas_used = executor.used_gas();
//     let exit_value = match handle_evm_result(exit_reason, ret) {
//         Ok(data) => data,
//         Err((err, data)) => {
//             return ExecutionResult::from_error(err, data, Some(gas_used));
//         }
//     };
//
//     if commit {
//         let (vals, logs) = executor.into_state().deconstruct();
//         if let Err(err) = backend.apply(vals, logs, false) {
//             return ExecutionResult::from_error(err.to_string(), Vec::default(), None)
//         }
//     }
//
//     ExecutionResult {
//         logs: backend.get_logs(),
//         data: exit_value,
//         gas_used,
//         vm_error: "".to_string(),
//     }
// }
//
// /// Creates new smart contract
// /// * querier - GoQuerier which is used to interact with Go (Cosmos) from SGX Enclave
// /// * backend - EVM backend for reading and writting state
// /// * gas_limit - gas limit for contract creation
// /// * from - creator address of smart contract
// /// * data - encoded bytecode and creation params
// /// * access_list - EIP-2930 access list
// /// * commit - should apply changes. Provide `false` if you want to simulate contract creation, without state changes
// ///
// /// Returns EVM execution result
// fn execute_create(
//     querier: *mut GoQuerier,
//     backend: &mut impl ExtendedBackend,
//     gas_limit: u64,
//     from: H160,
//     value: U256,
//     data: Vec<u8>,
//     access_list: Vec<(H160, Vec<H256>)>,
//     commit: bool,
// ) -> ExecutionResult {
//     let metadata = StackSubstateMetadata::new(gas_limit, &GASOMETER_CONFIG);
//     let state = MemoryStackState::new(metadata, backend);
//     let precompiles = EVMPrecompiles::new(querier);
//
//     let mut executor = StackExecutor::new_with_precompiles(state, &GASOMETER_CONFIG, &precompiles);
//     let (exit_reason, ret) = executor.transact_create(from, value, data, gas_limit, access_list);
//
//     let gas_used = executor.used_gas();
//     let exit_value = match handle_evm_result(exit_reason, ret) {
//         Ok(data) => data,
//         Err((err, data)) => return ExecutionResult::from_error(err, data, Some(gas_used)),
//     };
//
//     if commit {
//         let (vals, logs) = executor.into_state().deconstruct();
//         if let Err(err) = backend.apply(vals, logs, false) {
//             return ExecutionResult::from_error(err.to_string(), Vec::default(), None)
//         }
//     }
//
//     ExecutionResult {
//         logs: backend.get_logs(),
//         data: exit_value,
//         gas_used,
//         vm_error: "".to_string(),
//     }
// }
//
// /// Handles an EVM result to return either a successful result or a (readable) error reason.
// fn handle_evm_result(exit_reason: ExitReason, data: Vec<u8>) -> Result<Vec<u8>, (String, Vec<u8>)> {
//     match exit_reason {
//         ExitReason::Succeed(_) => Ok(data),
//         ExitReason::Revert(err) => Err((format!("execution reverted: {:?}", err), data)),
//         ExitReason::Error(err) => Err((format!("evm error: {:?}", err), data)),
//         ExitReason::Fatal(err) => Err((format!("fatal evm error: {:?}", err), data)),
//     }
// }
