use crate::errors::{handle_c_error_default, Error};
use crate::memory::{ByteSliceView, UnmanagedVector};
use crate::protobuf_generated::node;
use crate::types::{Allocation, AllocationWithResult, GoQuerier};

use protobuf::Message;
use sgx_types::*;
use sgx_urts::SgxEnclave;
use std::panic::catch_unwind;
use std::env;
use std::ops::Deref;
use std::time::Duration;
use std::path::Path;
use lazy_static::lazy_static;
use parking_lot::{Condvar, Mutex};

static ENCLAVE_FILE: &'static str = "enclave.signed.so";
static ENCLAVE_HOME: &'static str = env!("ENCLAVE_HOME", "please specify ENCLAVE_HOME env variable");
const ENCLAVE_LOCK_TIMEOUT: u64 = 6*5;

lazy_static! {
    pub static ref ENCLAVE_DOORBELL: EnclaveDoorbell = EnclaveDoorbell::new();
}

#[allow(dead_code)]
extern "C" {
    pub fn handle_request(
        eid: sgx_enclave_id_t,
        retval: *mut AllocationWithResult,
        querier: *mut GoQuerier,
        request: *const u8,
        len: usize,
    ) -> sgx_status_t;

    pub fn ecall_allocate(
        eid: sgx_enclave_id_t,
        retval: *mut Allocation,
        data: *const u8,
        len: usize,
    ) -> sgx_status_t;

    pub fn ecall_init_master_key(eid: sgx_enclave_id_t, retval: *mut sgx_status_t, reset_flag: i32) -> sgx_status_t;

    pub fn ecall_is_initialized(eid: sgx_enclave_id_t, retval: *mut i32) -> sgx_status_t;

    pub fn ecall_share_seed(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        socket_fd: c_int,
    ) -> sgx_status_t;

    pub fn ecall_request_seed(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        hostname: *const u8,
        data_len: usize,
        socket_fd: c_int,
    ) -> sgx_status_t;
}

pub fn init_enclave() -> SgxResult<SgxEnclave> {
    let mut launch_token: sgx_launch_token_t = [0; 1024];
    // call sgx_create_enclave to initialize an enclave instance
    let mut launch_token_updated: i32 = 0;
    // Debug Support: set 2nd parameter to 1
    let debug = 1;
    let mut misc_attr = sgx_misc_attribute_t {
        secs_attr: sgx_attributes_t { flags: 0, xfrm: 0 },
        misc_select: 0,
    };

    println!("[DEBUG] Initialize enclave");

    SgxEnclave::create(
        format!("{}/{}", ENCLAVE_HOME, ENCLAVE_FILE),
        debug,
        &mut launch_token,
        &mut launch_token_updated,
        &mut misc_attr,
    )
}

#[no_mangle]
/// Handles all incoming protobuf-encoded requests related to node setup
/// such as generating of attestation certificate, keys, etc.
pub unsafe extern "C" fn handle_initialization_request(
    request: ByteSliceView,
    error_msg: Option<&mut UnmanagedVector>,
) -> UnmanagedVector {
    let r = catch_unwind(|| {
        // Check if request is correct
        let req_bytes = request
            .read()
            .ok_or_else(|| Error::unset_arg(crate::cache::PB_REQUEST_ARG))?;

        let request = match protobuf::parse_from_bytes::<node::SetupRequest>(req_bytes) {
            Ok(request) => request,
            Err(e) => {
                return Err(Error::protobuf_decode(e.to_string()));
            }
        };

        let enclave_access_token = crate::enclave::ENCLAVE_DOORBELL
            .get_access(1) // This can never be recursive
            .ok_or(sgx_status_t::SGX_ERROR_BUSY)?;
        let evm_enclave = (*enclave_access_token)?;

        let result = match request.req {
            Some(req) => {
                match req {
                    node::SetupRequest_oneof_req::initializeMasterKey(req) => {
                        let mut retval = sgx_status_t::SGX_SUCCESS;
                        let should_reset = req.shouldReset as i32;
                        let res = ecall_init_master_key(evm_enclave.geteid(), &mut retval, should_reset);

                        match res {
                            sgx_status_t::SGX_SUCCESS => {}
                            _ => {
                                return Err(Error::enclave_error(res.as_str()));
                            }
                        };

                        match retval {
                            sgx_status_t::SGX_SUCCESS => {}
                            _ => {
                                return Err(Error::enclave_error(retval.as_str()));
                            }
                        }

                        // Create response, convert it to bytes and return
                        let response = node::InitializeMasterKeyResponse::new();
                        let response_bytes = match response.write_to_bytes() {
                            Ok(res) => res,
                            Err(_) => {
                                return Err(Error::protobuf_decode("Response encoding failed"));
                            }
                        };

                        Ok(response_bytes)
                    },
                    node::SetupRequest_oneof_req::startSeedServer(req) => {
                        let mut retval = sgx_status_t::SGX_SUCCESS;
                        let res = ecall_share_seed(evm_enclave.geteid(), &mut retval, req.fd);

                        match res {
                            sgx_status_t::SGX_SUCCESS => {}
                            _ => {
                                return Err(Error::enclave_error(res.as_str()));
                            }
                        };

                        match retval {
                            sgx_status_t::SGX_SUCCESS => {}
                            _ => {
                                return Err(Error::enclave_error(retval.as_str()));
                            }
                        }

                        // Create response, convert it to bytes and return
                        let response = node::StartSeedServerResponse::new();
                        let response_bytes = match response.write_to_bytes() {
                            Ok(res) => res,
                            Err(_) => {
                                return Err(Error::protobuf_decode("Response encoding failed"));
                            }
                        };

                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::nodeSeed(req) => {
                        if req.hostname.is_empty() {
                            return Err(Error::unset_arg("Hostname was not set"));
                        }

                        let mut retval = sgx_status_t::SGX_SUCCESS;
                        let res = ecall_request_seed(
                            evm_enclave.geteid(),
                            &mut retval,
                            req.hostname.as_ptr() as *const u8,
                            req.hostname.len(),
                            req.fd
                        );

                        match (res, retval) {
                            (sgx_status_t::SGX_SUCCESS, sgx_status_t::SGX_SUCCESS) => {}
                            (_, _) => {
                                return Err(Error::enclave_error(res.as_str()));
                            }
                        };

                        // Create response, convert it to bytes and return
                        let response = node::NodeSeedResponse::new();
                        let response_bytes = match response.write_to_bytes() {
                            Ok(res) => res,
                            Err(_) => {
                                return Err(Error::protobuf_decode("Response encoding failed"));
                            }
                        };

                        Ok(response_bytes)
                    },
                    node::SetupRequest_oneof_req::isInitialized(_) => {
                        println!("[SGX_WRAPPER] checking if node is initialized");
                        let mut retval = 0i32;
                        let res = ecall_is_initialized(evm_enclave.geteid(), &mut retval);

                        match res {
                            sgx_status_t::SGX_SUCCESS => {}
                            _ => {
                                return Err(Error::enclave_error(res.as_str()));
                            }
                        };

                        // Create response, convert it to bytes and return
                        let mut response = node::IsInitializedResponse::new();
                        response.isInitialized = retval != 0;
                        let response_bytes = match response.write_to_bytes() {
                            Ok(res) => res,
                            Err(_) => {
                                return Err(Error::protobuf_decode("[SGX_WRAPPER] Response encoding failed"));
                            }
                        };

                        Ok(response_bytes)
                    }
                }
            }
            None => Err(Error::protobuf_decode("Request unwrapping failed")),
        };

        result
    })
    .unwrap_or_else(|_| Err(Error::panic()));

    let data = handle_c_error_default(r, error_msg);
    UnmanagedVector::new(Some(data))
}

pub struct EnclaveDoorbell {
    enclave: SgxResult<SgxEnclave>,
    condvar: Condvar,
    /// Amount of tasks allowed to use the enclave at the same time.
    count: Mutex<u8>,
}

impl EnclaveDoorbell {
    fn new() -> Self {
        println!("Setting up enclave doorbell");
        Self {
            enclave: init_enclave(),
            condvar: Condvar::new(),
            count: Mutex::new(8),
        }
    }

    fn wait_for(&'static self, duration: Duration, query_depth: u32) -> Option<EnclaveAccessToken> {
        if query_depth == 1 {
            let mut count = self.count.lock();
            if *count == 0 {
                // try to wait for other tasks to complete
                let wait = self.condvar.wait_for(&mut count, duration);
                // double check that the count is nonzero, so there's an available slot in the enclave.
                if wait.timed_out() || *count == 0 {
                    return None;
                }
            }
            *count -= 1;
        }
        Some(EnclaveAccessToken::new(self, query_depth))
    }

    pub fn get_access(&'static self, query_depth: u32) -> Option<EnclaveAccessToken> {
        self.wait_for(Duration::from_secs(ENCLAVE_LOCK_TIMEOUT), query_depth)
    }
}

// NEVER add Clone or Copy
pub struct EnclaveAccessToken {
    doorbell: &'static EnclaveDoorbell,
    enclave: SgxResult<&'static SgxEnclave>,
    query_depth: u32,
}

impl EnclaveAccessToken {
    fn new(doorbell: &'static EnclaveDoorbell, query_depth: u32) -> Self {
        let enclave = doorbell.enclave.as_ref().map_err(|status| *status);
        Self {
            doorbell,
            enclave,
            query_depth,
        }
    }
}

impl Deref for EnclaveAccessToken {
    type Target = SgxResult<&'static SgxEnclave>;

    fn deref(&self) -> &Self::Target {
        &self.enclave
    }
}

impl Drop for EnclaveAccessToken {
    fn drop(&mut self) {
        if self.query_depth == 1 {
            let mut count = self.doorbell.count.lock();
            *count += 1;
            drop(count);
            self.doorbell.condvar.notify_one();
        }
    }
}
