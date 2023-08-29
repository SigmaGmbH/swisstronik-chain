use evm::backend::Basic;
use primitive_types::{H160, H256};

use std::vec::Vec;

pub mod mocked_storage;

/// A key-value storage trait
pub trait Storage {
    /// Checks if there is entity with such key exists in DB
    fn contains_key(&self, key: &H160) -> bool;

    /// Returns 32-byte cell from account storage
    fn get_account_storage_cell(&self, key: &H160, index: &H256) -> Option<H256>;

    /// Returns bytecode of contract with provided address
    fn get_account_code(&self, key: &H160) -> Option<Vec<u8>>;

    /// Returns account basic data (balance and nonce)
    fn get_account(&self, account: &H160) -> Basic;

    /// Updates account balance and nonce
    fn insert_account(&mut self, key: H160, data: Basic);

    /// Updates contract bytecode
    fn insert_account_code(&mut self, key: H160, code: Vec<u8>);

    /// Update storage cell value
    fn insert_storage_cell(&mut self, key: H160, index: H256, value: H256);

    /// Removes account (selfdestruct)
    fn remove(&mut self, key: &H160);

    /// Removes storage cell value
    fn remove_storage_cell(&mut self, key: &H160, index: &H256);
}
