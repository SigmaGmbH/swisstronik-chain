use alloc::vec::Vec;
use ethereum::Log;
use evm::backend::{RuntimeBaseBackend, RuntimeEnvironment};
use primitive_types::{H160, H256, U256};
use crate::backend::{FFIBackend, TxEnvironment};
use crate::{coder, querier};
use crate::protobuf_generated::ffi;
use crate::storage::FFIStorage;
use crate::types::{Storage, Vicinity};

pub struct UpdatedBackend<'state> {
    // We keep GoQuerier to make it accessible for `OCALL` handlers
    pub querier: *mut querier::GoQuerier,
    // Contains gas price and original sender
    pub vicinity: Vicinity,
    // Data storage
    pub storage: &'state mut FFIStorage,
    // Emitted events
    pub logs: Vec<Log>,
    // Transaction context
    pub environment: TxEnvironment,
}

impl<'state> UpdatedBackend<'state> {
    pub fn new(
        querier: *mut querier::GoQuerier,
        storage: &'state mut FFIStorage,
        vicinity: Vicinity,
        environment: TxEnvironment,
    ) -> Self {
        Self {
            querier,
            vicinity,
            state: storage,
            logs: vec![],
            environment,
        }
    }
}

impl<'state> RuntimeEnvironment for UpdatedBackend<'state> {
    fn block_hash(&self, number: U256) -> H256 {
        let encoded_request = coder::encode_query_block_hash(number);
        match querier::make_request(self.querier, encoded_request) {
            Some(result) => {
                // Decode protobuf
                let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryBlockHashResponse>(
                    result.as_slice(),
                ) {
                    Ok(res) => res,
                    Err(err) => {
                        println!("Cannot decode protobuf response: {:?}", err);
                        return H256::default();
                    }
                };
                H256::from_slice(decoded_result.hash.as_slice())
            }
            None => {
                println!("Get block hash failed. Empty response");
                H256::default()
            }
        }
    }

    fn block_number(&self) -> U256 {
        self.environment.block_number
    }

    fn block_coinbase(&self) -> H160 {
        self.environment.block_coinbase
    }

    fn block_timestamp(&self) -> U256 {
        self.environment.timestamp
    }

    fn block_difficulty(&self) -> U256 {
        U256::zero()
    }

    fn block_randomness(&self) -> Option<H256> {
        None
    }

    fn block_gas_limit(&self) -> U256 {
        self.environment.block_gas_limit
    }

    fn block_base_fee_per_gas(&self) -> U256 {
        self.environment.block_base_fee_per_gas
    }

    fn chain_id(&self) -> U256 {
        self.environment.chain_id
    }
}

impl<'state> RuntimeBaseBackend for UpdatedBackend<'state> {
    fn balance(&self, address: H160) -> U256 {
        self.storage.get_account(&address).0
    }

    fn code_size(&self, address: H160) -> U256 {
        todo!()
    }

    fn code_hash(&self, address: H160) -> H256 {
        todo!()
    }

    fn code(&self, address: H160) -> Vec<u8> {
        self.storage.get_account_code(&address).unwrap_or(Vec::new())
    }

    fn storage(&self, address: H160, index: H256) -> H256 {
        todo!()
    }

    fn transient_storage(&self, address: H160, index: H256) -> H256 {
        // Should be implemented by overlayed backend
        H256::zero()
    }

    fn exists(&self, address: H160) -> bool {
        self.storage.contains_key(&address)
    }

    fn nonce(&self, address: H160) -> U256 {
        self.storage.get_account(&address).1
    }
}