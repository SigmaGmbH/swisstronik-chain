// use std::ptr::null;

use crate::enclave::attestation::dcap_utils;
// use crate::enclave::ecall_get_target_info;
use crate::errors::Error;
use crate::types::{AllocationWithResult, GoQuerier};
use sgx_types::*;
// use std::time::*;

// use super::ecall_create_report;

pub struct EnclaveApi;

// fn create_app_enclave_report(
//     eid: sgx_enclave_id_t,
//     qe_target_info: sgx_target_info_t,
// ) -> Result<sgx_report_t, Error> {
//     let mut retval = sgx_status_t::SGX_SUCCESS;
//     let mut app_report = sgx_report_t::default();
//     let res = unsafe { ecall_create_report(eid, &mut retval, &qe_target_info, &mut app_report) };

//     if res != sgx_status_t::SGX_SUCCESS && retval != sgx_status_t::SGX_SUCCESS {
//         return Err(Error::enclave_error("Cannot create app enclave report"));
//     }

//     Ok(app_report)
// }

// fn get_qe_target_info() -> Result<sgx_target_info_t, Error> {
//     let mut qe_target_info = sgx_target_info_t::default();
//     let qe3_ret = unsafe { sgx_qe_get_target_info(&mut qe_target_info) };
//     if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!("[Enclave Wrapper] sgx_qe_get_target_info failed. Status code: {:?}", qe3_ret);
//         return Err(Error::enclave_error("sgx_qe_get_target_info failed"));
//     }

//     Ok(qe_target_info)
// }

// fn get_qe_quote_size() -> Result<u32, Error> {
//     let mut quote_size = 0u32;
//     let qe3_ret = unsafe { sgx_qe_get_quote_size(&mut quote_size) };
//     if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!("[Enclave Wrapper] sgx_qe_get_quote_size failed. Status code: {:?}", qe3_ret);
//         return Err(Error::enclave_error("sgx_qe_get_quote_size failed"));
//     }

//     Ok(quote_size)
// }

// fn get_qe_quote(app_report: sgx_report_t, quote_size: u32) -> Result<Vec<u8>, Error> {
//     let mut quote_buffer = vec![0u8; quote_size as usize];
//     let p_quote_buffer = quote_buffer.as_mut_ptr();
//     let qe3_ret = unsafe { sgx_qe_get_quote(&app_report, quote_size, p_quote_buffer) };
//     if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!("[Enclave Wrapper] sgx_qe_get_quote failed. Status code: {:?}", qe3_ret);
//         return Err(Error::enclave_error("sgx_qe_get_quote failed"));
//     }

//     Ok(quote_buffer)
// }

// fn generate_quote(eid: sgx_enclave_id_t) -> Result<Vec<u8>, Error> {
//     println!("[Enclave Wrapper] Step 1. Get target info");
//     let qe_target_info = get_qe_target_info()?;

//     println!("[Enclave Wrapper] Step 2. Create app report");
//     let app_report = create_app_enclave_report(eid, qe_target_info)?;

//     println!("[Enclave Wrapper] Step 3. Call sgx_qe_get_quote_size");
//     let quote_size = get_qe_quote_size()?;

//     println!("[Enclave Wrapper] Step 4. Call sgx_qe_get_quote");
//     let quote = get_qe_quote(app_report, quote_size)?;

//     Ok(quote)
// }

// fn get_app_enclave_target_info(eid: sgx_enclave_id_t) -> Result<sgx_target_info_t, Error> {
//     let mut retval = sgx_status_t::SGX_SUCCESS;
//     let mut app_enclave_target_info = sgx_target_info_t::default();
//     let res = unsafe {
//         ecall_get_target_info(eid, &mut retval, &mut app_enclave_target_info)
//     };
//     if res != sgx_status_t::SGX_SUCCESS && retval != sgx_status_t::SGX_SUCCESS {
//         println!("[Enclave Wrapper] ecall_get_target_info failed");
//         return Err(Error::enclave_error("ecall_get_target_info failed"));
//     } 

//     Ok(app_enclave_target_info)
// }

// fn set_qv_enclave_load_policy() -> Result<(), Error> {
//     let res = unsafe { sgx_qv_set_enclave_load_policy(sgx_ql_request_policy_t::SGX_QL_EPHEMERAL) };
//     if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!("[Enclave Wrapper] sgx_qv_set_enclave_load_policy failed. Status code: {:?}", res);
//         return Err(Error::enclave_error("sgx_qv_set_enclave_load_policy failed"));
//     }

//     Ok(())
// }

// fn get_qv_enclave_supplemental_data_size() -> Result<u32, Error> {
//     let mut supplemental_data_size = 0u32;
//     let res = unsafe { sgx_qv_get_quote_supplemental_data_size(&mut supplemental_data_size) };
//     if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!("[Enclave Wrapper] sgx_qv_get_quote_supplemental_data_size failed. Status code: {:?}", res);
//         return Err(Error::enclave_error("sgx_qv_get_quote_supplemental_data_size failed"));
//     }

//     Ok(supplemental_data_size)
// }

// fn verify_quote(eid: sgx_enclave_id_t, quote: Vec<u8>) -> Result<(), Error> {
//     println!("Verifying quote");

//     let rand_nonce = [42u8; 16];
//     let mut qve_report_info = sgx_ql_qe_report_info_t::default();
//     qve_report_info.nonce.rand = rand_nonce;
    
//     println!("[Enclave Wrapper] Step 1. Get target_info from our enclave");
//     let app_enclave_target_info = get_app_enclave_target_info(eid)?;
//     qve_report_info.app_enclave_target_info = app_enclave_target_info;

//     println!("[Enclave Wrapper] Step 2. sgx_qv_set_enclave_load_policy");
//     set_qv_enclave_load_policy()?;

//     println!("[Enclave Wrapper] Step 3. sgx_qv_get_quote_supplemental_data_size");
//     let supplemental_data_size = get_qv_enclave_supplemental_data_size()?;

//     // TODO: Maybe add check for supplemental data size

//     println!("[Enclave Wrapper] Step 4. sgx_qv_verify_quote");
//     let current_time = SystemTime::now()
//         .duration_since(UNIX_EPOCH)
//         .unwrap()
//         .as_secs();
//     let mut collateral_expiration_status = 0u32;
//     let mut quote_verification_result = sgx_ql_qv_result_t::SGX_QL_QV_RESULT_UNSPECIFIED;
//     let mut supplemental_data = vec![0u8; supplemental_data_size as usize];
//     let res = unsafe {
//         sgx_qv_verify_quote(
//             quote.as_ptr(),
//             quote.len() as u32,
//             null(),
//             current_time as i64,
//             &mut collateral_expiration_status,
//             &mut quote_verification_result,
//             &mut qve_report_info,
//             supplemental_data_size,
//             supplemental_data.as_mut_ptr(),
//         )
//     };
//     if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!(
//             "[Enclave Wrapper] sgx_qv_verify_quote failed. Reason: {:?}",
//             res
//         );
//         return Err(Error::enclave_error(res.as_str()));
//     }

//     println!("[Enclave Wrapper] Step 5. sgx_tvl_verify_qve_report_and_identity");
//     let qve_isvsvn_threshold: sgx_isv_svn_t = 3;
//     let mut ret_val = sgx_quote3_error_t::SGX_QL_SUCCESS;
//     let res = unsafe {
//         super::sgx_tvl_verify_qve_report_and_identity(
//             eid,
//             &mut ret_val,
//             quote.as_ptr(),
//             quote.len() as u32,
//             &qve_report_info,
//             current_time as i64,
//             collateral_expiration_status,
//             quote_verification_result,
//             supplemental_data.as_ptr(),
//             supplemental_data_size,
//             qve_isvsvn_threshold,
//         )
//     };
//     if res != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!(
//             "[Enclave Wrapper] sgx_tvl_verify_qve_report_and_identity failed. Reason: {:?}",
//             res
//         );
//         return Err(Error::enclave_error(res.as_str()));
//     }
//     if ret_val != sgx_quote3_error_t::SGX_QL_SUCCESS {
//         println!(
//             "[Enclave Wrapper] sgx_tvl_verify_qve_report_and_identity failed. Status code: {:?}",
//             ret_val
//         );
//         return Err(Error::enclave_error(ret_val.as_str()));
//     }

//     println!("Quote verified. Result: {:?}", quote_verification_result);

//     Ok(())
// }

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

        // Validate provided host
        if hostname.is_empty() {
            return Err(Error::unset_arg("Hostname was not set"));
        }

        let qe_target_info = dcap_utils::get_qe_target_info()?;
        let quote_size = dcap_utils::get_quote_size()?;

        let mut retval = sgx_status_t::SGX_SUCCESS;
        let res = unsafe {
            super::ecall_dcap_attestation(
                eid,
                &mut retval,
                hostname.as_ptr() as *const u8,
                hostname.len(),
                fd,
                &qe_target_info,
                quote_size,
            )
        };

        if res != sgx_status_t::SGX_SUCCESS {
            println!("[Enclave Wrapper] Cannot call `ecall_dcap_attestation`. Reason: {:?}", res);
            return Err(Error::enclave_error(res))
        }

        if retval != sgx_status_t::SGX_SUCCESS {
            println!("[Enclave Wrapper] `ecall_dcap_attestation` failed. Reason: {:?}", retval);
        }

        Ok(())
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
