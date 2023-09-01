#![no_std]

#[macro_use]
extern crate sgx_tstd as std;

use backend::ExtendedBackend;
use internal_types::ExecutionResult;
pub use ethereum;
pub use evm;
use evm::executor::stack::{MemoryStackState, StackExecutor, StackSubstateMetadata};
use evm::ExitReason;
pub use primitive_types;
use primitive_types::{H160, H256, U256};

use std::{string::String, string::ToString, vec::Vec};

use crate::backend::{Backend, GASOMETER_CONFIG};
pub use crate::backend::Vicinity;
use crate::precompiles::EVMPrecompiles;

pub mod backend;
pub mod storage;

mod precompiles;

/// Handles incoming request for calling some contract / funds transfer
pub fn handle_sgxvm_call(
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
pub fn handle_sgxvm_create(
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

#[cfg(test)]
mod tests {
    use core::ops::{Add, Sub};
    use primitive_types::{H160, U256, H256};
    use sha3::{Digest, Keccak256};
    use crate::backend::Backend;
    use crate::storage::mocked_storage::MockedStorage;
    use crate::{handle_sgxvm_call, handle_sgxvm_create, Vicinity};

    fn create_address(address: H160, nonce: u64) -> H160 {
        let mut stream = rlp::RlpStream::new_list(2);
        stream.append(&address);
        stream.append(&nonce);
        H256::from_slice(Keccak256::digest(&stream.out()).as_slice()).into()
    }

    #[test]
    fn test_contract_deployment() {
        // Prepare environment
        let sender = H160::from_slice(
            &hex::decode("8c3FfC3600bCb365F7141EAf47b5921aEfB7917a").unwrap()
        );
        let vicinity = Vicinity {
            origin: sender.clone(),
        };
        let mut storage = MockedStorage::default();
        let mut backend = Backend {
            vicinity,
            state: &mut storage,
            logs: vec![],
        };

        // Deploy contract which emits logs
        // Deployment data was taken from solidity tests from `chain` repo
        let contract_address = create_address(sender.clone(), backend.state.get_account(&sender.clone()).nonce.as_u64());
        let deployment_data = hex::decode("608060405234801561001057600080fd5b50610280806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80632933c3c91461003b5780636057361d14610057575b600080fd5b61005560048036038101906100509190610168565b610073565b005b610071600480360381019061006c9190610168565b610123565b005b806000819055507f87199fbf46fb4529ad34a05f4a4704392dd5527b5c0e6f29591e4fccb7fd2717816040516100a991906101a4565b60405180910390a17fe409dd6b927a692d5f15854e2af1f02b98987acf9c5c4dbe265f2826e64b336b816040516100e0919061021c565b60405180910390a1807f932182c87b2d9b135ef769772728a1da9de5b81063424f9dbd99333f717f2cc382604051610118919061021c565b60405180910390a250565b8060008190555050565b600080fd5b6000819050919050565b61014581610132565b811461015057600080fd5b50565b6000813590506101628161013c565b92915050565b60006020828403121561017e5761017d61012d565b5b600061018c84828501610153565b91505092915050565b61019e81610132565b82525050565b60006020820190506101b96000830184610195565b92915050565b600082825260208201905092915050565b7f546573744d736700000000000000000000000000000000000000000000000000600082015250565b60006102066007836101bf565b9150610211826101d0565b602082019050919050565b60006040820190508181036000830152610235816101f9565b90506102446020830184610195565b9291505056fea2646970667358221220da89886bcfc76a0346e37a726a7b282db80890d5ec6d8b6f87d5222c72c6bfc464736f6c63430008110033").unwrap();
        let deployment_result = handle_sgxvm_create(
            &mut backend,
            200000,
            sender.clone(),
            U256::zero(),
            deployment_data,
            vec![],
            true
        );

        // Check if contract was deployed correctly
        let contract_code = backend.state.get_account_code(&contract_address);
        assert!(contract_code.is_some());
    }

    #[test]
    fn handle_sgxvm_call_and_emit_logs() {
        // Prepare environment
        let sender = H160::from_slice(
            &hex::decode("8c3FfC3600bCb365F7141EAf47b5921aEfB7917a").unwrap()
        );
        let vicinity = Vicinity {
            origin: sender.clone(),
        };
        let mut storage = MockedStorage::default();
        let mut backend = Backend {
            vicinity,
            state: &mut storage,
            logs: vec![],
        };

        // Deploy contract which emits logs
        // Deployment data was taken from solidity tests from `chain` repo
        let contract_address = create_address(sender.clone(), backend.state.get_account(&sender.clone()).nonce.as_u64());
        let deployment_data = hex::decode("608060405234801561001057600080fd5b50610280806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80632933c3c91461003b5780636057361d14610057575b600080fd5b61005560048036038101906100509190610168565b610073565b005b610071600480360381019061006c9190610168565b610123565b005b806000819055507f87199fbf46fb4529ad34a05f4a4704392dd5527b5c0e6f29591e4fccb7fd2717816040516100a991906101a4565b60405180910390a17fe409dd6b927a692d5f15854e2af1f02b98987acf9c5c4dbe265f2826e64b336b816040516100e0919061021c565b60405180910390a1807f932182c87b2d9b135ef769772728a1da9de5b81063424f9dbd99333f717f2cc382604051610118919061021c565b60405180910390a250565b8060008190555050565b600080fd5b6000819050919050565b61014581610132565b811461015057600080fd5b50565b6000813590506101628161013c565b92915050565b60006020828403121561017e5761017d61012d565b5b600061018c84828501610153565b91505092915050565b61019e81610132565b82525050565b60006020820190506101b96000830184610195565b92915050565b600082825260208201905092915050565b7f546573744d736700000000000000000000000000000000000000000000000000600082015250565b60006102066007836101bf565b9150610211826101d0565b602082019050919050565b60006040820190508181036000830152610235816101f9565b90506102446020830184610195565b9291505056fea2646970667358221220da89886bcfc76a0346e37a726a7b282db80890d5ec6d8b6f87d5222c72c6bfc464736f6c63430008110033").unwrap();
        let deployment_result = handle_sgxvm_create(
            &mut backend,
            200000,
            sender.clone(),
            U256::zero(),
            deployment_data,
            vec![],
            true
        );
        println!("Contract address: {:?}", hex::encode(contract_address));

        // Check if contract was deployed correctly
        let contract_code = backend.state.get_account_code(&contract_address);
        assert!(contract_code.is_some());

        // Send transaction to emit events
        let transaction_data = hex::decode("2933c3c90000000000000000000000000000000000000000000000000000000000000378").unwrap();
        let transaction_result = handle_sgxvm_call(
            &mut backend,
            200000,
            sender,
            contract_address,
            U256::zero(),
            transaction_data,
            vec![],
            true
        );

        // This transaction emits some events, so logs should not be empty
        assert_ne!(transaction_result.logs.len(), 0);
        // Since this transaction contains call to the contract, used gas should be greater than intrinsic gas (21000)
        assert!(transaction_result.gas_used > 21000);
    }

    #[test]
    fn test_deployment_in_dry_mode() {
        // Prepare environment
        let sender = H160::from_slice(
            &hex::decode("8c3FfC3600bCb365F7141EAf47b5921aEfB7917a").unwrap()
        );
        let vicinity = Vicinity {
            origin: sender.clone(),
        };
        let mut storage = MockedStorage::default();
        let mut backend = Backend {
            vicinity,
            state: &mut storage,
            logs: vec![],
        };

        let sender_nonce_before = backend.state.get_account(&sender.clone()).nonce.as_u64();

        // Deploy contract which emits logs
        // Deployment data was taken from solidity tests from `chain` repo
        let contract_address = create_address(sender.clone(), sender_nonce_before);
        let deployment_data = hex::decode("608060405234801561001057600080fd5b50610280806100206000396000f3fe608060405234801561001057600080fd5b50600436106100365760003560e01c80632933c3c91461003b5780636057361d14610057575b600080fd5b61005560048036038101906100509190610168565b610073565b005b610071600480360381019061006c9190610168565b610123565b005b806000819055507f87199fbf46fb4529ad34a05f4a4704392dd5527b5c0e6f29591e4fccb7fd2717816040516100a991906101a4565b60405180910390a17fe409dd6b927a692d5f15854e2af1f02b98987acf9c5c4dbe265f2826e64b336b816040516100e0919061021c565b60405180910390a1807f932182c87b2d9b135ef769772728a1da9de5b81063424f9dbd99333f717f2cc382604051610118919061021c565b60405180910390a250565b8060008190555050565b600080fd5b6000819050919050565b61014581610132565b811461015057600080fd5b50565b6000813590506101628161013c565b92915050565b60006020828403121561017e5761017d61012d565b5b600061018c84828501610153565b91505092915050565b61019e81610132565b82525050565b60006020820190506101b96000830184610195565b92915050565b600082825260208201905092915050565b7f546573744d736700000000000000000000000000000000000000000000000000600082015250565b60006102066007836101bf565b9150610211826101d0565b602082019050919050565b60006040820190508181036000830152610235816101f9565b90506102446020830184610195565b9291505056fea2646970667358221220da89886bcfc76a0346e37a726a7b282db80890d5ec6d8b6f87d5222c72c6bfc464736f6c63430008110033").unwrap();
        let deployment_result = handle_sgxvm_create(
            &mut backend,
            200000,
            sender.clone(),
            U256::zero(),
            deployment_data,
            vec![],
            false  // We set false here to run contract deployment in a simulation mode
        );

        // Check if contract was deployed correctly
        let contract_code = backend.state.get_account_code(&contract_address);
        assert!(contract_code.is_none());

        let sender_nonce_after = backend.state.get_account(&sender.clone()).nonce.as_u64();
        assert_eq!(sender_nonce_before, sender_nonce_after)
    }

    #[test]
    fn test_transfer() {
        // Prepare environment
        let sender = H160::from_slice(&hex::decode("8c3FfC3600bCb365F7141EAf47b5921aEfB7917a").unwrap());
        let receiver = H160::from_slice(&hex::decode("0000000000000000000000000000000000000000").unwrap());
        let vicinity = Vicinity {
            origin: sender.clone(),
        };
        let mut storage = MockedStorage::default();
        let mut backend = Backend {
            vicinity,
            state: &mut storage,
            logs: vec![],
        };

        let sender_account_before = backend.state.get_account(&sender);
        let receiver_account_before = backend.state.get_account(&receiver);

        let amount_to_send = 10000;
        let deployment_result = handle_sgxvm_call(
            &mut backend,
            200000,
            sender.clone(),
            receiver.clone(),
            U256::from(amount_to_send),
            vec![],
            vec![],
            true
        );

        let sender_account_after = backend.state.get_account(&sender);
        let receiver_account_after = backend.state.get_account(&receiver);


        assert_eq!(sender_account_after.balance, sender_account_before.balance.sub(amount_to_send));
        assert_eq!(receiver_account_after.balance, receiver_account_before.balance.add(amount_to_send));
        assert_eq!(sender_account_after.nonce, sender_account_before.nonce.add(1));
    }

    #[test]
    fn test_transfer_in_dry_mode() {
        // Prepare environment
        let sender = H160::from_slice(&hex::decode("8c3FfC3600bCb365F7141EAf47b5921aEfB7917a").unwrap());
        let receiver = H160::from_slice(&hex::decode("0000000000000000000000000000000000000000").unwrap());
        let vicinity = Vicinity {
            origin: sender.clone(),
        };
        let mut storage = MockedStorage::default();
        let mut backend = Backend {
            vicinity,
            state: &mut storage,
            logs: vec![],
        };

        let sender_account_before = backend.state.get_account(&sender);
        let receiver_account_before = backend.state.get_account(&receiver);

        let amount_to_send = 10000;
        let deployment_result = handle_sgxvm_call(
            &mut backend,
            200000,
            sender.clone(),
            receiver.clone(),
            U256::from(amount_to_send),
            vec![],
            vec![],
            false // We set false here to run contract deployment in a simulation mode
        );

        let sender_account_after = backend.state.get_account(&sender);
        let receiver_account_after = backend.state.get_account(&receiver);


        assert_eq!(sender_account_after.balance, sender_account_before.balance);
        assert_eq!(receiver_account_after.balance, receiver_account_before.balance);
        assert_eq!(sender_account_after.nonce, sender_account_before.nonce);
    }
}
