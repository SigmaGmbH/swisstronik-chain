extern crate sgx_types;
extern crate sgx_urts;
extern crate errno;
extern crate thiserror;
extern crate protobuf;
extern crate lazy_static;
extern crate parking_lot;

pub mod enclave;
pub mod memory;
pub mod version;
pub mod errors;
pub mod types;
pub mod ocall;
pub mod protobuf_generated;

// We only interact with this crate via `extern "C"` interfaces, not those public
// exports. There are no guarantees those exports are stable.
// We keep them here such that we can access them in the docs (`cargo doc`).
pub use memory::{
    destroy_unmanaged_vector, new_unmanaged_vector, ByteSliceView, U8SliceView, UnmanagedVector,
};


pub fn main() {
    let enclave_access_token = enclave::ENCLAVE_DOORBELL
        .get_access(1) // This can never be recursive
        .ok_or(sgx_types::sgx_status_t::SGX_ERROR_BUSY).unwrap();

    let evm_enclave = (*enclave_access_token).unwrap();

    let is_initialized = enclave::enclave_api::EnclaveApi::is_enclave_initialized(evm_enclave.geteid());
    println!("Finish")
}