/// This file contains implementations of various OCALLs for SGX-enclave

use crate::errors::GoError;
use crate::memory::{U8SliceView, UnmanagedVector};
use crate::types::{Allocation, AllocationWithResult, GoQuerier};

use sgx_types::*;
use std::net::{SocketAddr, TcpStream};
use std::os::unix::io::IntoRawFd;
use std::slice;

#[no_mangle]
pub extern "C" fn ocall_get_quote(
    p_sigrl: *const u8,
    sigrl_len: u32,
    p_report: *const sgx_report_t,
    quote_type: sgx_quote_sign_type_t,
    p_spid: *const sgx_spid_t,
    p_nonce: *const sgx_quote_nonce_t,
    p_qe_report: *mut sgx_report_t,
    p_quote: *mut u8,
    _maxlen: u32,
    p_quote_len: *mut u32,
) -> sgx_status_t {
    let mut real_quote_len: u32 = 0;

    let ret = unsafe { sgx_calc_quote_size(p_sigrl, sigrl_len, &mut real_quote_len as *mut u32) };

    if ret != sgx_status_t::SGX_SUCCESS {
        println!("sgx_calc_quote_size returned {}", ret);
        return ret;
    }

    unsafe {
        *p_quote_len = real_quote_len;
    }

    let ret = unsafe {
        sgx_get_quote(
            p_report,
            quote_type,
            p_spid,
            p_nonce,
            p_sigrl,
            sigrl_len,
            p_qe_report,
            p_quote as *mut sgx_quote_t,
            real_quote_len,
        )
    };

    if ret != sgx_status_t::SGX_SUCCESS {
        println!("sgx_calc_quote_size returned {}", ret);
        return ret;
    }

    ret
}

#[no_mangle]
pub extern "C" fn ocall_get_update_info(
    platform_blob: *const sgx_platform_info_t,
    enclave_trusted: i32,
    update_info: *mut sgx_update_info_bit_t,
) -> sgx_status_t {
    unsafe { sgx_report_attestation_status(platform_blob, enclave_trusted, update_info) }
}

#[no_mangle]
pub extern "C" fn ocall_allocate(data: *const u8, len: usize) -> Allocation {
    let slice = unsafe { slice::from_raw_parts(data, len) };
    let mut vector_copy = slice.to_vec();

    let ptr = vector_copy.as_mut_ptr();
    let len = vector_copy.len();
    std::mem::forget(vector_copy);

    Allocation {
        result_ptr: ptr,
        result_len: len,
    }
}

#[no_mangle]
pub extern "C" fn ocall_sgx_init_quote(
    ret_ti: *mut sgx_target_info_t,
    ret_gid: *mut sgx_epid_group_id_t,
) -> sgx_status_t {
    unsafe { sgx_init_quote(ret_ti, ret_gid) }
}

#[no_mangle]
pub extern "C" fn ocall_get_ias_socket(ret_fd: *mut c_int) -> sgx_status_t {
    let port = 443;
    let hostname = "api.trustedservices.intel.com";
    let addr = lookup_ipv4(hostname, port);
    let sock = TcpStream::connect(&addr).expect("[-] Connect tls server failed!");

    unsafe {
        *ret_fd = sock.into_raw_fd();
    }

    sgx_status_t::SGX_SUCCESS
}

pub fn lookup_ipv4(host: &str, port: u16) -> SocketAddr {
    use std::net::ToSocketAddrs;

    let addrs = (host, port).to_socket_addrs().unwrap();
    for addr in addrs {
        if let SocketAddr::V4(_) = addr {
            return addr;
        }
    }

    unreachable!("Cannot lookup address");
}

#[no_mangle]
pub extern "C" fn ocall_query_raw(
    querier: *mut GoQuerier,
    request: *const u8,
    request_len: usize,
) -> AllocationWithResult {
    // Recover request and querier
    let request = unsafe { slice::from_raw_parts(request, request_len) };
    let querier = unsafe { &*querier };

    // Prepare vectors for output and error
    let mut output = UnmanagedVector::default();
    let mut error_msg = UnmanagedVector::default();

    // Make request to GoQuerier (Connector)
    let go_result: GoError = (querier.vtable.query_external)(
        querier.state,
        U8SliceView::new(Some(&request)),
        &mut output as *mut UnmanagedVector,
        &mut error_msg as *mut UnmanagedVector,
    )
    .into();

    // Consume vectors to destroy them
    let output = output.consume();
    let error_msg = error_msg.consume();

    match go_result {
        GoError::None => {
            let output = output.unwrap_or_default();

            // Bind the token to a local variable to ensure its
            // destructor runs in the end of the function
            let enclave_access_token = crate::enclave::ENCLAVE_DOORBELL
                // This is always called from an ocall contxt, so we don't want to wait for
                // an new TCS. To do that, we say that our query depth is >1, e.g. 2
                .get_access(2)
                .ok_or(sgx_status_t::SGX_ERROR_BUSY);

            let enclave_access_token = match enclave_access_token {
                Ok(token) => token,
                Err(status) => {
                    return AllocationWithResult {
                        result_ptr: std::ptr::null_mut(),
                        result_size: 0usize,
                        status,
                    };
                }
            };

            let enclave_id = enclave_access_token
                .expect("If we got here, surely the enclave has been loaded")
                .geteid();

            let mut allocation_result = std::mem::MaybeUninit::<Allocation>::uninit();

            let res = unsafe {
                crate::enclave::ecall_allocate(
                    enclave_id,
                    allocation_result.as_mut_ptr(),
                    output.as_ptr(),
                    output.len(),
                )
            };

            match res {
                sgx_status_t::SGX_SUCCESS => {
                    let allocation_result = unsafe { allocation_result.assume_init() };
                    return AllocationWithResult {
                        result_ptr: allocation_result.result_ptr,
                        result_size: output.len(),
                        status: sgx_status_t::SGX_SUCCESS,
                    };
                }
                _ => {
                    println!("ecall_allocate failed. Reason: {:?}", res.as_str());
                    return AllocationWithResult {
                        result_ptr: std::ptr::null_mut(),
                        result_size: 0usize,
                        status: res,
                    };
                }
            };
        }
        _ => {
            let err_msg = error_msg.unwrap_or_default();
            println!(
                "[OCALL] query_raw: got error: {:?} with message: {:?}",
                go_result,
                String::from_utf8_lossy(&err_msg)
            );
            return AllocationWithResult::default();
        }
    };
}

