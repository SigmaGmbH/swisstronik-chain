#![no_std]

extern crate sgx_tstd as std;

use ethereum::Log;
use std::{vec::Vec, string::String};

pub mod ffi;

#[derive(Clone, Debug, PartialEq)]
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
            data: data,
            gas_used: gas_used.unwrap_or(21000), // This is minimum gas fee to apply the transaction
            vm_error: reason
        }
    }
}
