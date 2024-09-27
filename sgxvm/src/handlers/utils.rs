use ethereum::Log;
use evm::interpreter::runtime::Log as RuntimeLog;
use evm::standard::TransactArgs;
use primitive_types::{H160, H256, U256};
use protobuf::RepeatedField;
use std::vec::Vec;
use crate::protobuf_generated::ffi::{AccessListItem, SGXVMCallParams, SGXVMCreateParams};

impl Into<TransactArgs> for SGXVMCallParams {
    fn into(self) -> TransactArgs {
        TransactArgs::Call {
            caller: H160::from_slice(&self.from),
            address: H160::from_slice(&self.to),
            value: U256::from_big_endian(&self.value),
            data: self.data,
            gas_limit: U256::from(self.gasLimit),
            gas_price: U256::from_big_endian(&self.gasPrice),
            access_list: parse_access_list(self.accessList),
        }
    }
}

impl Into<TransactArgs> for SGXVMCreateParams {
    fn into(self) -> TransactArgs {
        TransactArgs::Create {
            caller: H160::from_slice(&self.from),
            value: U256::from_big_endian(&self.value),
            init_code: self.data,
            salt: None,
            gas_limit: U256::from(self.gasLimit),
            gas_price: U256::from_big_endian(&self.gasPrice),
            access_list: parse_access_list(self.accessList),
        }
    }
}

pub fn construct_call_args(params: SGXVMCallParams, data: Vec<u8>) -> TransactArgs {
    TransactArgs::Call {
        caller: H160::from_slice(&params.from),
        address: H160::from_slice(&params.to),
        value: U256::from_big_endian(&params.value),
        data,
        gas_limit: U256::from(params.gasLimit),
        gas_price: U256::from_big_endian(&params.gasPrice),
        access_list: parse_access_list(params.accessList),
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

pub fn convert_logs(input: Vec<RuntimeLog>) -> Vec<Log> {
    input
        .into_iter()
        .map(|rl| Log {
            address: rl.address,
            topics: rl.topics,
            data: rl.data,
        })
        .collect()
}