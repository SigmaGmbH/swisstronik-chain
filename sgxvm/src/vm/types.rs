use alloc::string::ToString;
use std::{
    vec::Vec,
    string::String,
};
use ethereum::Log;
use evm::{
    standard::Config,
    interpreter::error::ExitError,
};
use primitive_types::{H160, H256, U256};
use rlp::RlpStream;
use sha3::{Digest, Keccak256};
use crate::protobuf_generated::ffi::{SGXVMCallRequest, SGXVMCreateRequest};
use crate::vm::utils::parse_access_list;

/// Current gasometer configuration. Was set to Cancun
pub static GASOMETER_CONFIG: Config = Config::cancun();

#[derive(Clone, Debug, PartialEq)]
/// Represents the result of a transaction or call execution.
/// Contains logs, data, gas used and error message
pub struct ExecutionResult {
    pub logs: Vec<Log>,
    pub data: Vec<u8>,
    pub gas_used: u64,
    pub vm_error: String
}

impl ExecutionResult {
    /// Creates execution result that only contains error reason and possible amount of used gas
    pub fn from_error(reason: String, data: Vec<u8>, gas_used: Option<u64>) -> Self {
        Self {
            logs: Vec::default(),
            gas_used: gas_used.unwrap_or(21000), // This is minimum gas fee to apply the transaction
            vm_error: reason,
            data,
        }
    }

    pub fn from_exit_error(error: ExitError, data: Vec<u8>, gas_used: u64) -> Self {
        let vm_error = match error {
            ExitError::Reverted => "reverted".to_string(),
            ExitError::Fatal(fatal) => {
                format!("{:?}", fatal)
            },
            ExitError::Exception(exit) => {
                format!("{:?}", exit)
            }
        };

        ExecutionResult { logs: vec![], data, gas_used, vm_error }
    }
}

enum TransactionType {
    Legacy,
    EIP2930,
    EIP1559,
}

pub struct Transaction {
    tx_type: TransactionType,

    nonce: U256,
    gas_limit: U256,
    to: Option<H160>,
    value: U256,
    data: Vec<u8>,
    chain_id: u64,
    gas_price: Option<U256>,

    // EIP1559
    max_priority_fee_per_gas: Option<U256>,
    max_fee_per_gas: Option<U256>,

    // EIP2930
    access_list: Vec<(H160, Vec<H256>)>,
}

impl Transaction {
    fn rlp_append_legacy(&self, stream: &mut RlpStream) {
        stream.begin_list(9);

        stream.append(&self.nonce);
        stream.append(&self.gas_price.unwrap_or_default());
        stream.append(&self.gas_limit);

        if let Some(to) = self.to {
            stream.append(&to);
        } else {
            stream.append_empty_data();
        }

        stream.append(&self.value);
        stream.append(&self.data);

        // EIP-155 fields for signing
        stream.append(&U256::from(self.chain_id));
        stream.append(&U256::zero());
        stream.append(&U256::zero());
    }

    fn rlp_append_access_list(&self, stream: &mut RlpStream) {
        stream.begin_list(self.access_list.len());
        for (address, storage_keys) in &self.access_list {
            stream.begin_list(2);
            stream.append(address);

            stream.begin_list(storage_keys.len());
            for key in storage_keys {
                stream.append(&key.as_bytes());
            }
        }
    }

    fn rlp_append_eip2930(&self, stream: &mut RlpStream) {
        stream.begin_list(8);

        stream.append(&self.chain_id);
        stream.append(&self.nonce);
        stream.append(&self.gas_price.unwrap_or_default());
        stream.append(&self.gas_limit);

        if let Some(to) = self.to {
            stream.append(&to);
        } else {
            stream.append_empty_data();
        }

        stream.append(&self.value);
        stream.append(&self.data);

        // Access list
        self.rlp_append_access_list(stream);
    }

    fn rlp_append_eip1559(&self, stream: &mut RlpStream) {
        stream.begin_list(9);

        stream.append(&self.chain_id);
        stream.append(&self.nonce);
        stream.append(&self.max_priority_fee_per_gas.unwrap_or_default());
        stream.append(&self.max_fee_per_gas.unwrap_or_default());
        stream.append(&self.gas_limit);

        if let Some(to) = self.to {
            stream.append(&to);
        } else {
            stream.append_empty_data();
        }

        stream.append(&self.value);
        stream.append(&self.data);

        // Access list
        self.rlp_append_access_list(stream);
    }

    pub fn hash(&self) -> H256 {
        match self.tx_type {
            TransactionType::Legacy => {
                let mut stream = RlpStream::new();
                self.rlp_append_legacy(&mut stream);
                let encoded = stream.out();
                H256::from_slice(Keccak256::digest(&encoded).as_slice())
            },
            TransactionType::EIP2930 => {
                let mut stream = RlpStream::new();
                self.rlp_append_eip2930(&mut stream);
                let encoded = stream.out();

                // EIP-2930 transactions are prefixed with 0x01 before hashing
                let mut prefix = vec![1];
                prefix.extend_from_slice(&encoded);

                H256::from_slice(Keccak256::digest(&prefix).as_slice())
            },
            TransactionType::EIP1559 => {
                let mut stream = RlpStream::new();
                self.rlp_append_eip1559(&mut stream);
                let encoded = stream.out();

                // EIP-1559 transactions are prefixed with 0x02 before hashing
                let mut prefix = vec![2];
                prefix.extend_from_slice(&encoded);

                H256::from_slice(Keccak256::digest(&prefix).as_slice())
            }
        }
    }
}

impl From<SGXVMCallRequest> for Transaction {
    fn from(request: SGXVMCallRequest) -> Transaction {
        let params = request.params.unwrap();
        let context = request.context.unwrap();

        let tx_type = match params.txType {
            0 => TransactionType::Legacy,
            1 => TransactionType::EIP2930,
            2 => TransactionType::EIP1559,
            _ => TransactionType::EIP1559,
        };

        Transaction {
            tx_type,
            nonce: U256::from(params.nonce),
            gas_limit: U256::from(params.gasLimit),
            to: Some(H160::from_slice(&params.to)),
            value: U256::from_big_endian(&params.value),
            data: params.data,
            chain_id: context.chain_id,
            gas_price: Some(U256::from_big_endian(&params.gasPrice)),
            max_priority_fee_per_gas: Some(U256::from_big_endian(&params.maxPriorityFeePerGas)),
            max_fee_per_gas: Some(U256::from_big_endian(&params.maxFeePerGas)),
            access_list: parse_access_list(params.accessList),
        }
    }
}

impl From<SGXVMCreateRequest> for Transaction {
    fn from(request: SGXVMCreateRequest) -> Transaction {
        let params = request.params.unwrap();
        let context = request.context.unwrap();

        let tx_type = match params.txType {
            0 => TransactionType::Legacy,
            1 => TransactionType::EIP2930,
            2 => TransactionType::EIP1559,
            _ => TransactionType::EIP1559,
        };

        Transaction {
            tx_type,
            nonce: U256::from(params.nonce),
            gas_limit: U256::from(params.gasLimit),
            to: None,
            value: U256::from_big_endian(&params.value),
            data: params.data,
            chain_id: context.chain_id,
            gas_price: Some(U256::from_big_endian(&params.gasPrice)),
            max_priority_fee_per_gas: Some(U256::from_big_endian(&params.maxPriorityFeePerGas)),
            max_fee_per_gas: Some(U256::from_big_endian(&params.maxFeePerGas)),
            access_list: parse_access_list(params.accessList),
        }
    }
}