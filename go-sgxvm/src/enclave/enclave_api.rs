use std::ptr::null;

use crate::errors::Error;
use crate::types::{AllocationWithResult, GoQuerier};
use sgx_types::*;
use std::time::*;

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
        println!("[Enclave Wrapper] perform_dcap_attestation");
        println!("/////\n[Enclave Wrapper] generate quote\n/////");

        println!("[Enclave Wrapper] Step 1. Get target info");
        let mut qe_target_info = sgx_target_info_t::default();
        let res = unsafe { sgx_qe_get_target_info(&mut qe_target_info as *mut _) };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!("[Enclave Wrapper] sgx_qe_get_target_info failed");
            return Err(Error::enclave_error(res.as_str()));
        }

        println!("[Enclave Wrapper] Step 2. Create app report");
        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let mut app_report = sgx_report_t::default();
        let res = unsafe {
            super::ecall_create_report(
                eid,
                &mut ret_val,
                &qe_target_info,
                &mut app_report as *mut sgx_report_t,
            )
        };
        if res != sgx_status_t::SGX_SUCCESS {
            println!("[Enclave Wrapper] ecall_create_report failed");
            return Err(Error::enclave_error(res.as_str()));
        }
        if ret_val != sgx_status_t::SGX_SUCCESS {
            println!("[Enclave Wrapper] ecall_create_report returned error code");
            return Err(Error::enclave_error(ret_val.as_str()));
        }

        println!("[Enclave Wrapper] Step 3. Call sgx_qe_get_quote_size");
        let mut quote_size = 0u32;
        let res = unsafe { sgx_qe_get_quote_size(&mut quote_size) };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!("[Enclave Wrapper] sgx_qe_get_quote_size failed");
            return Err(Error::enclave_error(res.as_str()));
        }

        println!("[Enclave Wrapper] Step 4. Call sgx_qe_get_quote");
        let mut quote = vec![0u8; quote_size as usize];
        let res = unsafe { sgx_qe_get_quote(&app_report, quote_size, quote.as_mut_ptr()) };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!("[Enclave Wrapper] sgx_qe_get_quote failed");
            return Err(Error::enclave_error(res.as_str()));
        }

        println!("/////\n[Enclave Wrapper] verify quote\n/////");

        println!("[Enclave Wrapper] Step 1. Get target_info from our enclave");
        let mut app_enclave_target_info = sgx_target_info_t::default();
        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe {
            super::ecall_get_target_info(eid, &mut ret_val, &mut app_enclave_target_info)
        };
        if res != sgx_status_t::SGX_SUCCESS {
            println!("[Enclave Wrapper] ecall_get_target_info failed");
            return Err(Error::enclave_error(res.as_str()));
        }
        if ret_val != sgx_status_t::SGX_SUCCESS {
            println!("[Enclave Wrapper] ecall_get_target_info returned error code");
            return Err(Error::enclave_error(ret_val.as_str()));
        }

        println!("[Enclave Wrapper] Step 2. sgx_qv_set_enclave_load_policy");
        let res =
            unsafe { sgx_qv_set_enclave_load_policy(sgx_ql_request_policy_t::SGX_QL_EPHEMERAL) };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!("[Enclave Wrapper] sgx_qv_set_enclave_load_policy failed");
            return Err(Error::enclave_error(res.as_str()));
        }

        println!("[Enclave Wrapper] Step 3. sgx_qv_get_quote_supplemental_data_size");
        let mut supplemental_data_size = 0u32;
        let res = unsafe { sgx_qv_get_quote_supplemental_data_size(&mut supplemental_data_size) };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!("[Enclave Wrapper] sgx_qv_get_quote_supplemental_data_size failed");
            return Err(Error::enclave_error(res.as_str()));
        }

        println!("[Enclave Wrapper] Step 4. sgx_qv_verify_quote");
        let current_time = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap()
            .as_secs();
        let mut p_collateral_expiration_status = 0u32;
        let mut p_quote_verification_result = sgx_ql_qv_result_t::SGX_QL_QV_RESULT_UNSPECIFIED;
        let mut p_qve_report_info = sgx_ql_qe_report_info_t::default();
        let mut supplemental_data = vec![0u8; supplemental_data_size as usize];
        let res = unsafe {
            sgx_qv_verify_quote(
                quote.as_ptr(),
                quote.len() as u32,
                null(),
                current_time as i64,
                &mut p_collateral_expiration_status,
                &mut p_quote_verification_result,
                &mut p_qve_report_info,
                supplemental_data_size,
                supplemental_data.as_mut_ptr(),
            )
        };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!(
                "[Enclave Wrapper] sgx_qv_verify_quote failed. Reason: {:?}",
                res
            );
            return Err(Error::enclave_error(res.as_str()));
        }

        println!("[Enclave Wrapper] Step 5. sgx_tvl_verify_qve_report_and_identity");
        let qve_isvsvn_threshold: sgx_isv_svn_t = 3;
        let mut ret_val = sgx_quote3_error_t::SGX_QL_SUCCESS;
        let res = unsafe {
            super::sgx_tvl_verify_qve_report_and_identity(
                eid,
                &mut ret_val,
                quote.as_ptr(),
                quote.len() as u32,
                &p_qve_report_info,
                current_time as i64,
                p_collateral_expiration_status,
                p_quote_verification_result,
                supplemental_data.as_ptr(),
                supplemental_data_size,
                qve_isvsvn_threshold,
            )
        };
        if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!(
                "[Enclave Wrapper] sgx_tvl_verify_qve_report_and_identity failed. Reason: {:?}",
                res
            );
            return Err(Error::enclave_error(res.as_str()));
        }
        if ret_val != sgx_quote3_error_t::SGX_QL_SUCCESS {
            println!(
                "[Enclave Wrapper] sgx_tvl_verify_qve_report_and_identity failed. Status code: {:?}",
                ret_val
            );
            return Err(Error::enclave_error(ret_val.as_str()));
        }

        println!("Quote verified. Result: {:?}", p_quote_verification_result);

        Ok(())

        // // Validate provided host
        // if hostname.is_empty() {
        //     return Err(Error::unset_arg("Hostname was not set"));
        // }

        // // Prepare target info for DCAP attestation
        // let target_info = match unsafe { crate::dcap::get_target_info() } {
        //     Ok(target_info) => target_info,
        //     Err(err) => return Err(Error::enclave_error(err.as_str())),
        // };
        // // Prepare quote size for DCAP attestation
        // let quote_size = match unsafe { crate::dcap::get_quote_size() } {
        //     Ok(quote_size) => quote_size,
        //     Err(err) => return Err(Error::enclave_error(err.as_str())),
        // };

        // let mut ret_val = sgx_status_t::SGX_SUCCESS;
        // let res = unsafe {
        //     super::ecall_dcap_attestation(
        //         eid,
        //         &mut ret_val,
        //         hostname.as_ptr() as *const u8,
        //         hostname.len(),
        //         fd,
        //         &target_info,
        //         quote_size,
        //     )
        // };

        // match (res, ret_val) {
        //     (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => Ok(()),
        //     (_, _) => {
        //         let error_str = if res == sgx_status_t::SGX_SUCCESS {
        //             res.as_str()
        //         } else {
        //             ret_val.as_str()
        //         };
        //         Err(Error::enclave_error(error_str))
        //     }
        // }
    }

    pub fn is_enclave_initialized(eid: sgx_enclave_id_t) -> Result<bool, Error> {
        let mut ret_val = 0i32;
        let res = unsafe { super::ecall_is_initialized(eid, &mut ret_val) };

        match res {
            sgx_status_t::SGX_SUCCESS => Ok(ret_val != 0),
            _ => Err(Error::enclave_error(res.as_str())),
        }
    }

    pub fn handle_evm_request(
        eid: sgx_enclave_id_t,
        request_bytes: &[u8],
        querier: GoQuerier,
    ) -> Result<Vec<u8>, Error> {
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
                println!(
                    "[Enclave Wrapper] Call to handle_request failed. Status code: {:?}",
                    evm_res
                );
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
                        handle_request_result.result_size,
                    )
                };
                Ok(data)
            }
            err => {
                println!("[Enclave Wrapper] EVM call failed. Status code: {:?}", err);
                Err(Error::vm_err(err))
            }
        }
    }
}
