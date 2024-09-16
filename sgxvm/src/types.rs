use alloc::string::ToString;
use evm::standard::Config;
use primitive_types::{H160, H256, U256};
use ethereum::Log;
use evm::interpreter::error::{ExitError, ExitFatal};
use evm::interpreter::error::ExitError::Fatal;
use std::{
    vec::Vec,
    string::String,
};
use std::boxed::Box;
use sgx_types::*;
use crate::error::Error;

pub static GASOMETER_CONFIG: Config = Config::cancun();

/// Information required by the evm
#[derive(Clone, Default, PartialEq, Eq)]
pub struct Vicinity {
    pub origin: H160,
    pub nonce: U256,
}

/// A key-value storage trait
pub trait Storage {
    /// Checks if there is entity with such key exists in DB
    fn contains_key(&self, key: &H160) -> bool;

    /// Returns 32-byte cell from account storage
    fn get_account_storage_cell(&self, key: &H160, index: &H256) -> Option<H256>;

    /// Returns bytecode of contract with provided address
    fn get_account_code(&self, key: &H160) -> Option<Vec<u8>>;

    /// Returns account basic data (balance and nonce)
    fn get_account(&self, account: &H160) -> (U256, U256);

    /// Updates account balance and nonce
    fn insert_account(&self, key: H160, data: (&U256, &U256)) -> Result<(), Error>;

    /// Updates contract bytecode
    fn insert_account_code(&self, key: H160, code: Vec<u8>) -> Result<(), Error>;

    /// Update storage cell value
    fn insert_storage_cell(&self, key: H160, index: H256, value: H256) -> Result<(), Error>;

    /// Removes account (selfdestruct)
    fn remove(&self, key: &H160) -> Result<(), Error>;

    /// Removes storage cell value
    fn remove_storage_cell(&self, key: &H160, index: &H256) -> Result<(), Error>;
}

// Struct for allocated buffer outside SGX Enclave
#[repr(C)]
#[allow(dead_code)]
pub struct AllocatedBuffer {
    pub ptr: *mut u8,
}

/// Recovers boxed value from pointer
#[allow(dead_code)]
pub unsafe fn recover_buffer(buf: AllocatedBuffer) -> Option<Vec<u8>> {
    if buf.ptr.is_null() {
        return None;
    }
    let boxed_vector = Box::from_raw(buf.ptr as *mut Vec<u8>);
    Some(*boxed_vector)
}

#[derive(Clone, Debug, PartialEq)]
pub struct ExecutionResult {
    pub logs: Vec<Log>,
    pub data: Vec<u8>,
    pub gas_used: u64,
    pub vm_error: String
}

impl From<ExitError> for ExecutionResult {
    fn from(value: ExitError) -> Self {
        ExecutionResult {
            logs: vec![],
            data: vec![],
            gas_used: 0,
            vm_error: value.to_string(),
        }
    }
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
}

#[repr(C)]
pub struct AllocationWithResult {
    pub result_ptr: *mut u8,
    pub result_len: usize,
    pub status: sgx_status_t
}

impl Default for AllocationWithResult {
    fn default() -> Self {
        AllocationWithResult {
            result_ptr: std::ptr::null_mut(),
            result_len: 0,
            status: sgx_status_t::SGX_ERROR_UNEXPECTED,
        }
    }
}

#[repr(C)]
pub struct Allocation {
    pub result_ptr: *mut u8,
    pub result_size: usize,
}