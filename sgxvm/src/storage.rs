use primitive_types::{H160, H256, U256};
use std::vec::Vec;

use crate::{
    coder, encryption, error::Error, protobuf_generated::ffi, querier, types::Storage
};

/// This struct allows us to obtain state from keeper
/// that is located outside of Rust code
pub struct FFIStorage {
    pub querier: *mut querier::GoQuerier,
    pub context_timestamp: u64,
    pub context_block_number: u64,
}

impl Storage for FFIStorage {
    fn contains_key(&self, key: &H160) -> bool {
        let encoded_request = coder::encode_contains_key(key);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            // Decode protobuf
            let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryContainsKeyResponse>(result.as_slice()) {
                Ok(res) => res,
                Err(err) => {
                    println!("Cannot decode protobuf response: {:?}", err);
                    return false
                }
            };
            decoded_result.contains
        } else {
            println!("Contains key failed. Empty response");
            false
        }
    }

    fn get_account_storage_cell(&self, key: &H160, index: &H256) -> Option<H256> {
        let encoded_request = coder::encode_get_storage_cell(key, index);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            // Decode protobuf
            let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryGetAccountStorageCellResponse>(result.as_slice()) {
                Ok(res) => res,
                Err(err) => {
                    println!("Cannot decode protobuf response: {:?}", err);
                    return None
                }
            };

            // Decrypt result
            if decoded_result.value.is_empty() {
                return None;
            }

            let decrypted_result = match encryption::decrypt_storage_cell(key.as_bytes().to_vec(), decoded_result.value) {
                Ok(decrypted_result) => decrypted_result,
                Err(err) => {
                    println!("Cannot decrypt result. Reason: {:?}", err);
                    return None;
                }
            };

            Some(H256::from_slice(&decrypted_result))
        } else {
            println!("Get account storage cell failed. Empty response");
            None
        }
    }

    fn get_account_code(&self, key: &H160) -> Option<Vec<u8>> {
        let encoded_request = coder::encode_get_account_code(key);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            // Decode protobuf
            let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryGetAccountCodeResponse>(result.as_slice()) {
                Ok(res) => res,
                Err(err) => {
                    println!("Cannot decode protobuf response: {:?}", err);
                    return None
                }
            };

            Some(decoded_result.code)
        } else {
            println!("Get account code failed. Empty response");
            None
        }
    }

    fn get_account(&self, key: &H160) -> (U256, U256) {
        let encoded_request = coder::encode_get_account(key);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            // Decode protobuf
            let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryGetAccountResponse>(result.as_slice()) {
                Ok(res) => res,
                Err(err) => {
                    println!("Cannot decode protobuf response: {:?}", err);
                    return (U256::zero(), U256::zero());
                }
            };
            
            (
                U256::from_big_endian(decoded_result.balance.as_slice()),
                U256::from(decoded_result.nonce)
            )
        } else {
            println!("Get account failed. Empty response");
            (U256::zero(), U256::zero())
        }
    }

    fn insert_account_code(&self, key: H160, code: Vec<u8>) -> Result<(), Error>  {
        let encoded_request = coder::encode_insert_account_code(key, code);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryInsertAccountCodeResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                _ => Ok(())
            }
        } else {
            Err(Error::enclave_err("Insert account code failed. Empty response"))
        }
    }

    fn insert_storage_cell(&self, key: H160, index: H256, value: H256) -> Result<(), Error>  {
        // Encrypt value
        let encrypted_value = encryption::encrypt_storage_cell(
            key.as_bytes().to_vec(), 
            self.context_block_number,
            self.context_timestamp.to_be_bytes().to_vec(),
            value.as_bytes().to_vec()
        )?;

        let encoded_request = coder::encode_insert_storage_cell(key, index, encrypted_value);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryInsertStorageCellResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                _ => Ok(())
            }
        } else {
            Err(Error::enclave_err("Insert storage cell failed. Empty response"))
        }
    }

    fn remove(&self, key: &H160) -> Result<(), Error>  {
        let encoded_request = coder::encode_remove(key);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryRemoveResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                _ => Ok(())
            }
        } else {
            Err(Error::enclave_err("Remove failed. Empty response"))
        }
    }

    fn remove_storage_cell(&self, key: &H160, index: &H256) -> Result<(), Error>  {
        let encoded_request = coder::encode_remove_storage_cell(key, index);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryRemoveStorageCellResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                _ => Ok(())
            }
        } else {
            Err(Error::enclave_err("Remove storage cell failed. Empty response"))
        }
    }

    fn insert_account_balance(&self, address: &H160, balance: &U256) -> Result<(), Error> {
        let encoded_request = coder::encode_insert_account_balance(address, balance);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryInsertAccountBalanceResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                _ => Ok(())
            }
        } else {
            Err(Error::enclave_err("Insert account balance failed. Empty response"))
        }
    }

    fn insert_account_nonce(&self, address: &H160, nonce: &U256) -> Result<(), Error> {
        let encoded_request = coder::encode_insert_account_nonce(address, nonce);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryInsertAccountNonceResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                _ => Ok(())
            }
        } else {
            Err(Error::enclave_err("Insert account nonce failed. Empty response"))
        }
    }

    fn get_account_code_size(&self, address: &H160) -> Result<U256, Error> {
        let encoded_request = coder::encode_get_account_code_size(address);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryGetAccountCodeSizeResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                Ok(res) => {
                    // println!("Got account code size: {:?}", U256::from(res.size));
                    Ok(U256::from(res.size))
                }
            }
        } else {
            Err(Error::enclave_err("Get account code size failed. Empty response"))
        }
    }

    fn get_account_code_hash(&self, address: &H160) -> Result<H256, Error> {
        let encoded_request = coder::encode_get_account_code_hash(address);
        if let Some(result) = querier::make_request(self.querier, encoded_request) {
            match protobuf::parse_from_bytes::<ffi::QueryGetAccountCodeHashResponse>(result.as_slice()) {
                Err(err) => {
                    Err(err.into())
                },
                Ok(res) => {
                    // println!("Got account code hash: {:?}", H256::from_slice(&res.hash));
                    Ok(H256::from_slice(&res.hash))
                }
            }
        } else {
            Err(Error::enclave_err("Get account code size failed. Empty response"))
        }
    }
}

impl FFIStorage {
    pub fn new(querier: *mut querier::GoQuerier, context_timestamp: u64, context_block_number: u64) -> Self {
        Self { querier, context_timestamp, context_block_number }
    }
}
