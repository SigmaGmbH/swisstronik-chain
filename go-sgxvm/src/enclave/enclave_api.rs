use sgx_types::*;
use crate::types::{
    GoQuerier,
    AllocationWithResult
};
use crate::errors::Error;

pub struct EnclaveApi;

impl EnclaveApi {
    pub fn check_node_status(eid: sgx_enclave_id_t) -> Result<(), Error> {
        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe { super::ecall_status(eid, &mut ret_val) };

        match (res, ret_val) {
            (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Ok(()),
            (_, _) => {
                let error_str = if res == sgx_status_t::SGX_SUCCESS {
                    res.as_str()
                } else {
                    ret_val.as_str()
                };
                Err(Error::enclave_error(error_str))
            }
        }
    }

    pub fn initialize_master_key(eid: sgx_enclave_id_t, reset: bool) -> Result<(), Error> {
        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe { super::ecall_init_master_key(eid, &mut ret_val, reset as i32) };

        match (res, ret_val) {
            (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Ok(()),
            (_, _) => {
                let error_str = if res == sgx_status_t::SGX_SUCCESS {
                    res.as_str()
                } else {
                    ret_val.as_str()
                };
                Err(Error::enclave_error(error_str))
            }
        }
    }

    pub fn start_bootstrap_server(eid: sgx_enclave_id_t, fd: i32) -> Result<(), Error> {
        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe { super::ecall_share_seed(eid, &mut ret_val, fd) };

        match (res, ret_val) {
            (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Ok(()),
            (_, _) => {
                let error_str = if res == sgx_status_t::SGX_SUCCESS {
                    res.as_str()
                } else {
                    ret_val.as_str()
                };
                Err(Error::enclave_error(error_str))
            }
        }
    }

    pub fn perform_epid_attestation(
        eid: sgx_enclave_id_t,
        hostname: String,
        fd: i32,
    ) -> Result<(), Error> {
        if hostname.is_empty() {
            return Err(Error::unset_arg("Hostname was not set"));
        }

        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe {
            super::ecall_request_seed(
                eid,
                &mut ret_val,
                hostname.as_ptr() as *const u8,
                hostname.len(),
                fd,
            )
        };

        match (res, ret_val) {
            (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Ok(()),
            (_, _) => {
                let error_str = if res == sgx_status_t::SGX_SUCCESS {
                    res.as_str()
                } else {
                    ret_val.as_str()
                };
                Err(Error::enclave_error(error_str))
            }
        }
    }

    pub fn perform_dcap_attestation(
        eid: sgx_enclave_id_t,
        hostname: String,
        fd: i32,
    ) -> Result<(), Error> {
        // Validate provided host
        if hostname.is_empty() {
            return Err(Error::unset_arg("Hostname was not set"));
        }

        // Prepare target info for DCAP attestation
        let target_info = match unsafe { crate::dcap::get_target_info() } {
            Ok(target_info) => target_info,
            Err(err) => return Err(Error::enclave_error(err.as_str())),
        };
        // Prepare quote size for DCAP attestation
        let quote_size = match unsafe { crate::dcap::get_quote_size() } {
            Ok(quote_size) => quote_size,
            Err(err) => return Err(Error::enclave_error(err.as_str())),
        };

        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe {
            super::ecall_dcap_attestation(
                eid,
                &mut ret_val,
                hostname.as_ptr() as *const u8,
                hostname.len(),
                fd,
                &target_info,
                quote_size,
            )
        };

        match (res, ret_val) {
            (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Ok(()),
            (_, _) => {
                let error_str = if res == sgx_status_t::SGX_SUCCESS {
                    res.as_str()
                } else {
                    ret_val.as_str()
                };
                Err(Error::enclave_error(error_str))
            }
        }
    }

    pub fn is_enclave_initialized(eid: sgx_enclave_id_t) -> Result<bool, Error> {
        let mut ret_val = 0i32;
        let res = unsafe { super::ecall_is_initialized(eid, &mut ret_val) };

        match res {
            sgx_status_t::SGX_SUCCESS => Ok(ret_val != 0),
            _ => Err(Error::enclave_error(res.as_str())),
        }
    }

    pub fn handle_evm_request(eid: sgx_enclave_id_t, request_bytes: &[u8], querier: GoQuerier) -> Result<Vec<u8>, Error> {
        let request_vec = Vec::from(request_bytes);
        let mut querier = querier;
        let mut ret_val = std::mem::MaybeUninit::<AllocationWithResult>::uninit();

        let evm_res = unsafe {
            super::handle_request(
                eid,
                ret_val.as_mut_ptr(),
                &mut querier as *mut GoQuerier,
                request_vec.as_ptr(),
                request_vec.len(),
            )
        };

        let handle_request_result = unsafe { ret_val.assume_init() };

        match evm_res {
            sgx_status_t::SGX_SUCCESS => (),
            err => {
                println!("[Enclave Wrapper] Call to handle_request failed. Status code: {:?}", evm_res);
                return Err(Error::enclave_error(err));
            }
        }

        // Parse execution result
        match handle_request_result.status {
            sgx_status_t::SGX_SUCCESS => {
                let data = unsafe { 
                    Vec::from_raw_parts(
                        handle_request_result.result_ptr, 
                        handle_request_result.result_size, 
                        handle_request_result.result_size
                    ) 
                };
                Ok(data)
            },
            err => {
                println!("[Enclave Wrapper] EVM call failed. Status code: {:?}", err);
                Err(Error::vm_err(err))
            }
        }
    }
}
