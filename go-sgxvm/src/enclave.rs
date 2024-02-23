use crate::dcap;
use crate::errors::{handle_c_error_default, Error};
use crate::memory::{ByteSliceView, UnmanagedVector};
use crate::protobuf_generated::node;
use crate::types::{Allocation, AllocationWithResult, GoQuerier};

use lazy_static::lazy_static;
use parking_lot::{Condvar, Mutex};
use protobuf::Message;
use sgx_types::*;
use sgx_urts::SgxEnclave;
use std::env;
use std::ops::Deref;
use std::panic::catch_unwind;
use std::time::Duration;

static ENCLAVE_FILE: &'static str = "enclave.signed.so";
const ENCLAVE_LOCK_TIMEOUT: u64 = 6 * 5;

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

    pub fn ecall_init_master_key(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        reset_flag: i32,
    ) -> sgx_status_t;

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

    pub fn ecall_status(eid: sgx_enclave_id_t, retval: *mut sgx_status_t) -> sgx_status_t;

    pub fn ecall_dcap_attestation(
        eid: sgx_enclave_id_t,
        retval: *mut sgx_status_t,
        hostname: *const u8,
        data_len: usize,
        socket_fd: c_int,
        qe_target_info: &sgx_target_info_t,
        quote_size: u32,
    ) -> sgx_status_t;
}

pub struct EnclaveApi;

impl EnclaveApi {
    pub fn check_node_status(eid: sgx_enclave_id_t) -> Result<(), Error> {
        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe { ecall_status(eid, &mut ret_val) };

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
        let res = unsafe { ecall_init_master_key(eid, &mut ret_val, reset as i32) };

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
        let res = unsafe { ecall_share_seed(eid, &mut ret_val, fd) };

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
            ecall_request_seed(
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
        let target_info = match unsafe { dcap::get_target_info() } {
            Ok(target_info) => target_info,
            Err(err) => return Err(Error::enclave_error(err.as_str())),
        };
        // Prepare quote size for DCAP attestation
        let quote_size = match unsafe { dcap::get_quote_size() } {
            Ok(quote_size) => quote_size,
            Err(err) => return Err(Error::enclave_error(err.as_str())),
        };

        let mut ret_val = sgx_status_t::SGX_SUCCESS;
        let res = unsafe {
            ecall_dcap_attestation(
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
        let res = unsafe { ecall_is_initialized(eid, &mut ret_val) };

        match res {
            sgx_status_t::SGX_SUCCESS => Ok(ret_val != 0),
            _ => Err(Error::enclave_error(res.as_str())),
        }
    }
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

    let enclave_home = match env::var("ENCLAVE_HOME") {
        Ok(home) => home,
        Err(_) => {
            let dir_path = String::from(
                std::env::home_dir()
                    .expect("Please specify ENCLAVE_HOME env variable explicitly")
                    .to_str()
                    .unwrap(),
            );
            format!("{}/.swisstronik-enclave", dir_path)
        }
    };
    let enclave_path = format!("{}/{}", enclave_home, ENCLAVE_FILE);

    println!(
        "[DEBUG] Initialize enclave. Enclave location: {:?}",
        enclave_path
    );

    SgxEnclave::create(
        enclave_path,
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
                    node::SetupRequest_oneof_req::nodeStatus(_req) => {
                        EnclaveApi::check_node_status(evm_enclave.geteid())?;
                        let response = node::NodeStatusResponse::new();
                        let response_bytes: Vec<u8> = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::initializeMasterKey(req) => {
                        EnclaveApi::initialize_master_key(evm_enclave.geteid(), req.shouldReset)?;
                        let response = node::InitializeMasterKeyResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::startBootstrapServer(req) => {
                        EnclaveApi::start_bootstrap_server(evm_enclave.geteid(), req.fd)?;
                        let response = node::StartBootstrapServerResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::epidAttestationRequest(req) => {
                        EnclaveApi::perform_epid_attestation(evm_enclave.geteid(), req.hostname, req.fd)?;
                        let response = node::EPIDAttestationResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::dcapAttestationRequest(req) => {
                        EnclaveApi::perform_dcap_attestation(evm_enclave.geteid(), req.hostname, req.fd)?;
                        let response = node::DCAPAttestationResponse::new();
                        let response_bytes = response.write_to_bytes()?;
                        Ok(response_bytes)
                    }
                    node::SetupRequest_oneof_req::isInitialized(_) => {
                        let is_initialized = EnclaveApi::is_enclave_initialized(evm_enclave.geteid())?;
                        let mut response = node::IsInitializedResponse::new();
                        response.isInitialized = is_initialized;
                        let response_bytes = response.write_to_bytes()?;
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
        println!("[Enclave Doorbell] Setting up enclave doorbell");
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
