use protobuf::Message;
use crate::protobuf_generated::ffi::NodePublicKeyResponse;
use crate::AllocationWithResult;
use crate::key_manager::KeyManager;

/// Handles incoming request for node public key
pub fn handle_public_key_request() -> AllocationWithResult {
    let key_manager = match KeyManager::unseal() {
        Ok(manager) => manager,
        Err(err) => {
            return AllocationWithResult::default()
        }
    };

    let public_key = key_manager.get_public_key();

    let mut response = NodePublicKeyResponse::new();
    response.set_publicKey(public_key);

    let encoded_response = match response.write_to_bytes() {
        Ok(res) => res,
        Err(err) => {
            println!("Cannot encode protobuf result. Reason: {:?}", err);
            return AllocationWithResult::default();
        }
    };
    
    super::allocate_inner(encoded_response)
}