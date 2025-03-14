use alloc::string::ToString;
use primitive_types::{H160, H256, U256};
use protobuf::RepeatedField;
use std::vec::Vec;
use rlp::{RlpStream};
use sha3::{Digest, Keccak256};
use crate::protobuf_generated::ffi::{AccessListItem, SGXVMCallRequest, SGXVMCreateRequest};

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
        println!("APPEND LEGACY");
        println!("Chain id: {}", self.chain_id);
        println!("nonce: {}", self.nonce);
        println!("max_priority_fee_per_gas: {}", self.max_priority_fee_per_gas.unwrap_or_default().to_string());
        println!("max_fee_per_gas: {}", self.max_fee_per_gas.unwrap_or_default().to_string());
        println!("gas_limit: {}", self.gas_limit.to_string() );
        println!("to: {}", self.to.unwrap_or_default().to_string() );
        println!("value: {}", self.value.to_string());
        println!("data: {:?}", self.data);
        println!("access list: {:?}", self.access_list);
        println!("gas price: {:?}", self.gas_price.unwrap_or_default().to_string());

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
        println!("APPEND 2930");
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
        println!("APPEND 1559");
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

        let tx_type = match (params.accessList.is_empty(), params.maxFeePerGas.is_empty(), params.maxPriorityFeePerGas.is_empty()) {
            (true, true, true) => TransactionType::Legacy,
            (false, _, _) => TransactionType::EIP2930,
            (_, false, _) => TransactionType::EIP1559,
            _ => TransactionType::EIP1559,
        };

        println!("params gas price: {:?}", params.gasPrice);
        println!("params max fee per gas: {:?}", params.maxFeePerGas);
        println!("params max priority fee per gas: {:?}", params.maxPriorityFeePerGas);
        println!("params access list: {:?}", params.accessList);

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

        let tx_type = match (params.accessList.is_empty(), params.maxFeePerGas.is_empty(), params.maxPriorityFeePerGas.is_empty()) {
            (true, true, true) => TransactionType::Legacy,
            (false, _, _) => TransactionType::EIP2930,
            (_, false, _) => TransactionType::EIP1559,
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

pub fn parse_access_list(data: RepeatedField<AccessListItem>) -> Vec<(H160, Vec<H256>)> {
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