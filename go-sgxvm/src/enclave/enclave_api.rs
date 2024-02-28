use std::ptr::null;

use crate::enclave::ecall_get_target_info;
use crate::errors::Error;
use crate::types::{AllocationWithResult, GoQuerier};
use sgx_types::*;
use std::time::*;

use super::ecall_create_report;

pub struct EnclaveApi;

// fn output(quote: Vec<u8>) {
//     let p_quote3: *const sgx_quote3_t = quote.as_ptr() as *const sgx_quote3_t;

//     // copy heading bytes to a sgx_quote3_t type to simplify access
//     let quote3: sgx_quote3_t = unsafe { *p_quote3 };

//     let quote_signature_data_vec: Vec<u8> = quote[std::mem::size_of::<sgx_quote3_t>()..].into();

//     //println!("quote3 header says signature data len = {}", quote3.signature_data_len);
//     //println!("quote_signature_data len = {}", quote_signature_data_vec.len());

//     assert_eq!(
//         quote3.signature_data_len as usize,
//         quote_signature_data_vec.len()
//     );

//     // signature_data has a header of sgx_ql_ecdsa_sig_data_t structure
//     //let p_sig_data: * const sgx_ql_ecdsa_sig_data_t = quote_signature_data_vec.as_ptr() as _;
//     // mem copy
//     //let sig_data = unsafe { * p_sig_data };

//     // sgx_ql_ecdsa_sig_data_t is followed by sgx_ql_auth_data_t
//     // create a new vec for auth_data
//     let auth_certification_data_offset = std::mem::size_of::<sgx_ql_ecdsa_sig_data_t>();
//     let p_auth_data: *const sgx_ql_auth_data_t =
//         (quote_signature_data_vec[auth_certification_data_offset..]).as_ptr() as _;
//     let auth_data_header: sgx_ql_auth_data_t = unsafe { *p_auth_data };
//     //println!("auth_data len = {}", auth_data_header.size);

//     let auth_data_offset =
//         auth_certification_data_offset + std::mem::size_of::<sgx_ql_auth_data_t>();

//     // It should be [0,1,2,3...]
//     // defined at https://github.com/intel/SGXDataCenterAttestationPrimitives/blob/4605fae1c606de4ff1191719433f77f050f1c33c/QuoteGeneration/quote_wrapper/quote/qe_logic.cpp#L1452
//     //let auth_data_vec: Vec<u8> = quote_signature_data_vec[auth_data_offset..auth_data_offset + auth_data_header.size as usize].into();
//     //println!("Auth data:\n{:?}", auth_data_vec);

//     let temp_cert_data_offset = auth_data_offset + auth_data_header.size as usize;
//     let p_temp_cert_data: *const sgx_ql_certification_data_t =
//         quote_signature_data_vec[temp_cert_data_offset..].as_ptr() as _;
//     let temp_cert_data: sgx_ql_certification_data_t = unsafe { *p_temp_cert_data };

//     //println!("certification data offset = {}", temp_cert_data_offset);
//     //println!("certification data size = {}", temp_cert_data.size);

//     let cert_info_offset =
//         temp_cert_data_offset + std::mem::size_of::<sgx_ql_certification_data_t>();

//     //println!("cert info offset = {}", cert_info_offset);
//     // this should be the last structure
//     assert_eq!(
//         quote_signature_data_vec.len(),
//         cert_info_offset + temp_cert_data.size as usize
//     );

//     // let tail_content = quote_signature_data_vec[cert_info_offset..].to_vec();
//     // let enc_ppid_len = 384;
//     // let enc_ppid: &[u8] = &tail_content[0..enc_ppid_len];
//     // let pce_id: &[u8] = &tail_content[enc_ppid_len..enc_ppid_len + 2];
//     // let cpu_svn: &[u8] = &tail_content[enc_ppid_len + 2..enc_ppid_len + 2 + 16];
//     // let pce_isvsvn: &[u8] = &tail_content[enc_ppid_len + 2 + 16..enc_ppid_len + 2 + 18];
//     // println!("EncPPID:\n{:02x}", enc_ppid.iter().format(""));
//     // println!("PCE_ID:\n{:02x}", pce_id.iter().format(""));
//     // println!("TCBr - CPUSVN:\n{:02x}", cpu_svn.iter().format(""));
//     // println!("TCBr - PCE_ISVSVN:\n{:02x}", pce_isvsvn.iter().format(""));
//     // println!("QE_ID:\n{:02x}", quote3.header.user_data.iter().format(""));
// }

fn create_app_enclave_report(
    eid: sgx_enclave_id_t,
    qe_target_info: sgx_target_info_t,
    app_report: *mut sgx_report_t,
) -> bool {
    let mut retval = sgx_status_t::SGX_SUCCESS;
    let sgx_status = unsafe { ecall_create_report(eid, &mut retval, &qe_target_info, app_report) };
    if sgx_status != sgx_status_t::SGX_SUCCESS && retval != sgx_status_t::SGX_SUCCESS {
        println!("create_app_enclave_report failed");
        return false;
    }

    true
}

fn generate_quote(eid: sgx_enclave_id_t) -> SgxResult<Vec<u8>> {
    println!("[Enclave Wrapper] Step 1. Get target info");
    let mut qe_target_info = sgx_target_info_t::default();
    let qe3_ret = unsafe { sgx_qe_get_target_info(&mut qe_target_info) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!("[Enclave Wrapper] sgx_qe_get_target_info failed");
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
    }

    println!("[Enclave Wrapper] Step 2. Create app report");
    let mut app_report = sgx_report_t::default();
    if create_app_enclave_report(eid, qe_target_info, &mut app_report) != true {
        println!("FAIL");
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
    }

    println!("[Enclave Wrapper] Step 3. Call sgx_qe_get_quote_size");
    let mut quote_size = 0u32;
    let qe3_ret = unsafe { sgx_qe_get_quote_size(&mut quote_size) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!("[Enclave Wrapper] sgx_qe_get_quote_size failed");
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
    }

    println!("[Enclave Wrapper] Step 4. Call sgx_qe_get_quote");
    let mut quote_buffer = vec![0u8; quote_size as usize];
    let p_quote_buffer = quote_buffer.as_mut_ptr();
    let qe3_ret = unsafe { sgx_qe_get_quote(&app_report, quote_size, p_quote_buffer) };
    if qe3_ret != sgx_quote3_error_t::SGX_QL_SUCCESS {
        println!("[Enclave Wrapper] sgx_qe_get_quote failed");
        return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
    }

    Ok(quote_buffer)
}

fn verify_quote(eid: sgx_enclave_id_t, quote: Vec<u8>) -> Result<(), Error> {
    println!("Verifying quote");

    let rand_nonce = [42u8; 16];
    let mut qve_report_info = sgx_ql_qe_report_info_t::default();
    qve_report_info.nonce.rand = rand_nonce;
    
    println!("[Enclave Wrapper] Step 1. Get target_info from our enclave");
    let mut get_target_info_ret = sgx_status_t::SGX_SUCCESS;
    // Use separate variable to fix issue with packed fields
    let mut app_enclave_target_info = sgx_target_info_t::default();
    let sgx_ret = unsafe {
        ecall_get_target_info(eid, &mut get_target_info_ret, &mut app_enclave_target_info)
    };
    if sgx_ret != sgx_status_t::SGX_SUCCESS && get_target_info_ret != sgx_status_t::SGX_SUCCESS {
        println!("[Enclave Wrapper] ecall_get_target_info failed");
        return Err(Error::enclave_error(get_target_info_ret.as_str()));
    } 
    qve_report_info.app_enclave_target_info = app_enclave_target_info;




    println!("[Enclave Wrapper] Step 2. sgx_qv_set_enclave_load_policy");
    let res = unsafe { sgx_qv_set_enclave_load_policy(sgx_ql_request_policy_t::SGX_QL_EPHEMERAL) };
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

    // TODO: Maybe add check for supplemental data size

    println!("[Enclave Wrapper] Step 4. sgx_qv_verify_quote");
    let current_time = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs();
    let mut collateral_expiration_status = 0u32;
    let mut quote_verification_result = sgx_ql_qv_result_t::SGX_QL_QV_RESULT_UNSPECIFIED;
    let mut supplemental_data = vec![0u8; supplemental_data_size as usize];
    let res = unsafe {
        sgx_qv_verify_quote(
            quote.as_ptr(),
            quote.len() as u32,
            null(),
            current_time as i64,
            &mut collateral_expiration_status,
            &mut quote_verification_result,
            &mut qve_report_info,
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
            &qve_report_info,
            current_time as i64,
            collateral_expiration_status,
            quote_verification_result,
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

    println!("Quote verified. Result: {:?}", quote_verification_result);

    Ok(())
}

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
        let quote = generate_quote(eid).map_err(|err| Error::enclave_error(err.as_str()))?;
        verify_quote(eid, quote)?;
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
