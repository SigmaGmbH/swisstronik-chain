use evm::standard::TransactArgs;
use primitive_types::{H160, U256};
use crate::protobuf_generated::ffi::{SGXVMCallParams, SGXVMCreateParams};
use crate::vm::utils::parse_access_list;

impl From<SGXVMCallParams> for TransactArgs {
    fn from(val: SGXVMCallParams) -> Self {
        TransactArgs::Call {
            caller: H160::from_slice(&val.from),
            address: H160::from_slice(&val.to),
            value: U256::from_big_endian(&val.value),
            data: val.data,
            gas_limit: U256::from(val.gasLimit),
            gas_price: U256::from_big_endian(&val.gasPrice),
            access_list: parse_access_list(val.accessList),
        }
    }
}

impl From<SGXVMCreateParams> for TransactArgs {
    fn from(val: SGXVMCreateParams) -> Self {
        TransactArgs::Create {
            caller: H160::from_slice(&val.from),
            value: U256::from_big_endian(&val.value),
            init_code: val.data,
            salt: None,
            gas_limit: U256::from(val.gasLimit),
            gas_price: U256::from_big_endian(&val.gasPrice),
            access_list: parse_access_list(val.accessList),
        }
    }
}