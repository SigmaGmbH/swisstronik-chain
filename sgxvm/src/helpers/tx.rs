use primitive_types::{H160, H256, U256};
use std::vec::Vec;
use rlp::{RlpStream};
use sha3::{Digest, Keccak256};

enum TransactionType {
    Legacy,
    EIP2930,
    EIP1559,
}

struct Transaction {
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
    }

    fn rlp_append_access_list(&self, stream: &mut RlpStream) {
        stream.begin_list(self.access_list.len());
        for (address, storage_keys) in &self.access_list {
            stream.begin_list(2);
            stream.append(&address);

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

    fn hash(&self) -> H256 {
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