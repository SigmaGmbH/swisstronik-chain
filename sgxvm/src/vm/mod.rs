use evm::standard::{Etable, EtableResolver, TransactArgs, TransactValue};
use primitive_types::{H160, H256, U256};
use protobuf::Message;
use std::vec::Vec;

use crate::encryption::{
    decrypt_transaction_data, encrypt_transaction_data, extract_public_key_and_data,
};
use crate::key_manager::utils::random_nonce;
use crate::protobuf_generated::ffi::{
    HandleTransactionResponse, 
    Log, 
    SGXVMCallRequest, 
    SGXVMCreateRequest, 
    SGXVMEstimateGasRequest, 
    Topic, 
    TransactionContext
};
use crate::utils::allocate_inner;
use crate::std::string::ToString;
use crate::AllocationWithResult;
use crate::GoQuerier;

use crate::vm::{
    precompiles::EVMPrecompiles,
    invoker::OverlayedInvoker,
    backend::{TxEnvironment, Backend},
    storage::StorageWithQuerier,
    types::{ExecutionResult, GASOMETER_CONFIG, Transaction},
    utils::{recover_sender, parse_access_list, convert_logs},
};

pub mod backend;
pub mod invoker;
pub mod storage;
pub mod precompiles;
pub mod types;
pub mod utils;

/// Inner handler for EVM call request
pub fn handle_call_request_inner(
    querier: *mut GoQuerier,
    data: SGXVMCallRequest,
) -> ExecutionResult {
    let tx = Transaction::from(data.clone());
    let tx_hash = tx.hash();

    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let tx_sender = if params.signature.is_empty() || H160::from_slice(&params.from).eq(&H160::zero()) {
        H160::default()
    } else {
        match recover_sender(&tx_hash, &params.signature) {
            Some(sender) => H160::from_slice(&sender),
            None => H160::default()
        }
    };

    if !tx_sender.eq(&H160::from_slice(&params.from)) {
        return ExecutionResult::from_error("Corrupted signature. Provided sender is invalid".to_string(), Vec::new(), None)
    }

    let should_commit = params.commit;
    let block_number = context.block_number;

    // Check if transaction is unencrypted, handle it as regular EVM transaction
    let is_unencrypted = params.data.is_empty() || params.unencrypted;
    if is_unencrypted {
        return run_tx(querier, context, params.into(), should_commit)
    }

    // Otherwise, we should decrypt input, execute tx and encrypt output
    let (user_public_key, data, nonce) = match extract_public_key_and_data(params.data) {
        Ok(res) => res,
        Err(err) => return ExecutionResult::from_error(err.to_string(), Vec::new(), None)
    };

    let decrypted_data = match !data.is_empty() {
        true => match decrypt_transaction_data(data, user_public_key.clone(), block_number) {
            Ok(data) => data,
            Err(err) => return ExecutionResult::from_error(err.to_string(), Vec::new(), None)
        },
        false => Vec::new()
    };

    let transact_args = TransactArgs::Call {
        caller: H160::from_slice(&params.from),
        address: H160::from_slice(&params.to),
        value: U256::from_big_endian(&params.value),
        data: decrypted_data,
        gas_limit: U256::from(params.gasLimit),
        gas_price: U256::from_big_endian(&params.gasPrice),
        access_list: parse_access_list(params.accessList),
    };
    let mut execution_result = run_tx(querier, context, transact_args, should_commit);

    let nonce = match nonce.is_empty() {
        true => {
            match random_nonce() {
                Ok(nonce) => nonce.to_vec(),
                Err(err) => return ExecutionResult::from_error(err.to_string(), Vec::new(), None)
            }
        },
        false => nonce,
    };

    if execution_result.vm_error.is_empty() {
        let encrypted_response = match encrypt_transaction_data(execution_result.data, user_public_key, nonce, block_number) {
            Ok(data) => data,
            Err(err) => return ExecutionResult::from_error(err.to_string(), Vec::new(), None)
        };

        execution_result.data = encrypted_response;
    }

    execution_result
}

/// Inner handler for EVM create request
pub fn handle_create_request_inner(
    querier: *mut GoQuerier,
    data: SGXVMCreateRequest,
) -> ExecutionResult {
    let tx = Transaction::from(data.clone());
    let tx_hash = tx.hash();

    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let tx_sender = if params.signature.is_empty() || H160::from_slice(&params.from).eq(&H160::zero()) {
        H160::default()
    } else {
        match recover_sender(&tx_hash, &params.signature) {
            Some(sender) => H160::from_slice(&sender),
            None => H160::default()
        }
    };

    if !tx_sender.eq(&H160::from_slice(&params.from)) {
        return ExecutionResult::from_error("Corrupted signature. Provided sender is invalid".to_string(), Vec::new(), None)
    }

    let should_commit = params.commit;

    run_tx(querier, context, params.into(), should_commit)
}

pub fn handle_estimate_gas_request_inner(
    querier: *mut GoQuerier,
    data: SGXVMEstimateGasRequest,
) -> ExecutionResult {
    let params = data.params.unwrap();
    let context = data.context.unwrap();

    let mut execution_result = match params.to.len() {
        0 => {
            // Handle `create`
            let transact_args = TransactArgs::Create {
                caller: H160::from_slice(&params.from),
                value: U256::from_big_endian(&params.value),
                init_code: params.data,
                salt: None,
                gas_limit: U256::from(params.gasLimit),
                gas_price: U256::from_big_endian(&params.gasPrice),
                access_list: parse_access_list(params.accessList),
            };

            run_tx(querier, context, transact_args, false)
        },
        _ => {
            // Handle `call`
            let block_number = context.block_number;

            // Check if transaction is unencrypted, handle it as regular EVM transaction
            let is_unencrypted = params.data.is_empty() || params.unencrypted;
            if is_unencrypted {
                let transact_args = TransactArgs::Call {
                    caller: H160::from_slice(&params.from),
                    address: H160::from_slice(&params.to),
                    value: U256::from_big_endian(&params.value),
                    data: params.data,
                    gas_limit: U256::from(params.gasLimit),
                    gas_price: U256::from_big_endian(&params.gasPrice),
                    access_list: parse_access_list(params.accessList),
                };
                return run_tx(querier, context, transact_args, false)
            } else {
                // Otherwise, we should decrypt input, execute tx and encrypt output
                let (user_public_key, data, _) = match extract_public_key_and_data(params.data) {
                    Ok(res) => res,
                    Err(err) => return ExecutionResult::from_error(err.to_string(), Vec::new(), None)
                };

                let decrypted_data = match !data.is_empty() {
                    true => match decrypt_transaction_data(data, user_public_key, block_number) {
                        Ok(data) => data,
                        Err(err) => return ExecutionResult::from_error(err.to_string(), Vec::new(), None)
                    },
                    false => Vec::new()
                };

                let transact_args = TransactArgs::Call {
                    caller: H160::from_slice(&params.from),
                    address: H160::from_slice(&params.to),
                    value: U256::from_big_endian(&params.value),
                    data: decrypted_data,
                    gas_limit: U256::from(params.gasLimit),
                    gas_price: U256::from_big_endian(&params.gasPrice),
                    access_list: parse_access_list(params.accessList),
                };

                run_tx(querier, context, transact_args, false)
            }
        }
    };

    if execution_result.vm_error.is_empty() {
        execution_result.data = Vec::default();
        execution_result.logs = Vec::default();
    }

    execution_result
}

fn run_tx(
    querier: *mut GoQuerier,
    context: TransactionContext,
    args: TransactArgs,
    should_commit: bool,
) -> ExecutionResult {
    let gas_etable = Etable::single(evm::standard::eval_gasometer);
    let exec_etable = Etable::runtime();
    let etable = (gas_etable, exec_etable);
    let precompiles = EVMPrecompiles::new(querier);
    let resolver = EtableResolver::new(&GASOMETER_CONFIG, &precompiles, &etable);
    let invoker = OverlayedInvoker::new(&GASOMETER_CONFIG, &resolver);

    let storage = StorageWithQuerier::new(querier, context.timestamp, context.block_number);
    let tx_environment = TxEnvironment::from(context);
    let mut backend = Backend::new(querier, &storage, tx_environment);

    let res = evm::transact(args, None, &mut backend, &invoker);
    let changeset = backend.deconstruct();

    let used_gas = invoker.get_gas_used().map(|used_gas| used_gas.as_u64()).unwrap_or(21000);

    if should_commit {
        if let Err(err) = Backend::apply_changeset(&storage, &changeset) {
            return ExecutionResult::from_exit_error(
                err,
                invoker.get_return_value().unwrap_or_default(),
                used_gas,
            );
        }
    }

    match res {
        Ok(res) => {
            match res {
                TransactValue::Call {succeed: _, retval} => {
                    ExecutionResult {
                        logs: convert_logs(changeset.logs),
                        data: retval,
                        gas_used: used_gas,
                        vm_error: "".to_string()
                    }
                }
                TransactValue::Create {succeed: _, address} => {
                    // Check if run_tx was called in context of transaction or in context of eth_call or eth_estimateGas.
                    // We commit changes only in case of transaction context.
                    if should_commit {
                        ExecutionResult {
                            logs: convert_logs(changeset.logs),
                            data: address.to_fixed_bytes().to_vec(),
                            gas_used: used_gas,
                            vm_error: "".to_string()
                        }
                    } else {
                        ExecutionResult {
                            logs: convert_logs(changeset.logs),
                            data: invoker.get_return_value().unwrap_or_default(),
                            gas_used: used_gas,
                            vm_error: "".to_string()
                        }
                    }

                }
            }
        },
        Err(err) => {
            let error_data = invoker.get_return_value().unwrap_or_default();
            ExecutionResult::from_exit_error(err, error_data, used_gas)
        }
    }
}