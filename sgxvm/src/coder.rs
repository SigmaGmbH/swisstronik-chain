use ethabi::Address;
use primitive_types::{H160, U256, H256};
use protobuf::Message;
use crate::protobuf_generated::ffi;
use std::{
    vec::Vec,
    string::String
};

fn u256_to_vec(value: &U256) -> Vec<u8> {
    let mut buffer = [0u8; 32];
    value.to_big_endian(&mut buffer);
    buffer.to_vec()
}

pub fn encode_query_block_hash(number: &U256) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryBlockHash::new();
    request.set_number(u256_to_vec(number));
    cosmos_request.set_blockHash(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_account(account_address: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryGetAccount::new();
    request.set_address(account_address.as_bytes().to_vec());
    cosmos_request.set_getAccount(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_contains_key(account_address: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryContainsKey::new();
    request.set_key(account_address.as_bytes().to_vec());
    cosmos_request.set_containsKey(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_storage_cell(account_address: &H160, index: &H256) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryGetAccountStorageCell::new();
    request.set_address(account_address.as_bytes().to_vec());
    request.set_index(index.as_bytes().to_vec());
    cosmos_request.set_storageCell(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_account_code(account_address: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryGetAccountCode::new();
    request.set_address(account_address.as_bytes().to_vec());
    cosmos_request.set_accountCode(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_insert_account_code(account_address: H160, code: Vec<u8>) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryInsertAccountCode::new();
    request.set_address(account_address.as_bytes().to_vec());
    request.set_code(code);
    cosmos_request.set_insertAccountCode(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_insert_storage_cell(account_address: H160, index: H256, value: Vec<u8>) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryInsertStorageCell::new();
    request.set_address(account_address.as_bytes().to_vec());
    request.set_index(index.as_bytes().to_vec());
    request.set_value(value);
    cosmos_request.set_insertStorageCell(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_remove(account_address: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryRemove::new();
    request.set_address(account_address.as_bytes().to_vec());
    cosmos_request.set_remove(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_remove_storage_cell(account_address: &H160, index: &H256) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryRemoveStorageCell::new();
    request.set_address(account_address.as_bytes().to_vec());
    request.set_index(index.as_bytes().to_vec());
    cosmos_request.set_removeStorageCell(request);
    cosmos_request.write_to_bytes().unwrap()
}


pub fn encode_add_verification_details_request(
    user_address: Address,
    issuer_address: H160,
    origin_chain: String,
    verification_type: u32,
    issuance_timestamp: u32,
    expiration_timestamp: u32,
    proof_data: Vec<u8>,
    schema: String,
    issuer_verification_id: String,
    version: u32,
) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryAddVerificationDetails::new();

    request.set_userAddress(user_address.as_bytes().to_vec());
    request.set_issuerAddress(issuer_address.as_bytes().to_vec());
    request.set_originChain(origin_chain);
    request.set_verificationType(verification_type);
    request.set_issuanceTimestamp(issuance_timestamp);
    request.set_expirationTimestamp(expiration_timestamp);
    request.set_proofData(proof_data);
    request.set_schema(schema);
    request.set_issuerVerificationId(issuer_verification_id);
    request.set_version(version);

    cosmos_request.set_addVerificationDetails(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_add_verification_details_v2_request(
    user_address: Address,
    issuer_address: H160,
    origin_chain: String,
    verification_type: u32,
    issuance_timestamp: u32,
    expiration_timestamp: u32,
    proof_data: Vec<u8>,
    schema: String,
    issuer_verification_id: String,
    version: u32,
    user_public_key: Vec<u8>,
) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryAddVerificationDetailsV2::new();

    request.set_userAddress(user_address.as_bytes().to_vec());
    request.set_issuerAddress(issuer_address.as_bytes().to_vec());
    request.set_originChain(origin_chain);
    request.set_verificationType(verification_type);
    request.set_issuanceTimestamp(issuance_timestamp);
    request.set_expirationTimestamp(expiration_timestamp);
    request.set_proofData(proof_data);
    request.set_schema(schema);
    request.set_issuerVerificationId(issuer_verification_id);
    request.set_version(version);
    request.set_userPublicKey(user_public_key);

    cosmos_request.set_addVerificationDetailsV2(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_issuance_tree_root_request() -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let request = ffi::QueryIssuanceTreeRoot::new();
    cosmos_request.set_issuanceTreeRoot(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_revocation_tree_root_request() -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let request = ffi::QueryRevocationTreeRoot::new();
    cosmos_request.set_revocationTreeRoot(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_has_verification_request(
    user_address: H160,
    verification_type: u32,
    expiration_timestamp: u32,
    allowed_issuers: Vec<Address>,
) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryHasVerification::new();

    request.set_userAddress(user_address.as_bytes().to_vec());
    request.set_verificationType(verification_type);
    request.set_expirationTimestamp(expiration_timestamp);

    let issuers_vec: Vec<Vec<u8>> = allowed_issuers.into_iter().map(|issuer| issuer.as_bytes().to_vec()).collect();
    request.set_allowedIssuers(issuers_vec.into());

    cosmos_request.set_hasVerification(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_verification_data(user_address: Address, issuer_address: H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryGetVerificationData::new();

    request.set_userAddress(user_address.as_bytes().to_vec());
    request.set_issuerAddress(issuer_address.as_bytes().to_vec());

    cosmos_request.set_getVerificationData(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_insert_account_balance(address: &H160, balance: &U256) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryInsertAccountBalance::new();

    request.set_address(address.as_bytes().to_vec());
    request.set_balance(u256_to_vec(balance));

    cosmos_request.set_insertAccountBalance(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_insert_account_nonce(address: &H160, nonce: &U256) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryInsertAccountNonce::new();

    request.set_address(address.as_bytes().to_vec());
    request.set_nonce(nonce.as_u64());

    cosmos_request.set_insertAccountNonce(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_account_code_hash(address: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryGetAccountCodeHash::new();

    request.set_address(address.as_bytes().to_vec());

    cosmos_request.set_codeHash(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_get_account_code_size(address: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryGetAccountCodeSize::new();

    request.set_address(address.as_bytes().to_vec());

    cosmos_request.set_codeSize(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_revoke_verification(verification_id: Vec<u8>, issuer: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryRevokeVerification::new();

    request.set_verificationId(verification_id);
    request.set_issuer(issuer.as_bytes().to_vec());

    cosmos_request.set_revokeVerification(request);
    cosmos_request.write_to_bytes().unwrap()
}

pub fn encode_convert_credential(verification_id: Vec<u8>, holder_public_key: Vec<u8>, caller: &H160) -> Vec<u8> {
    let mut cosmos_request = ffi::CosmosRequest::new();
    let mut request = ffi::QueryConvertCredential::new();

    request.set_holderPublicKey(holder_public_key);
    request.set_verificationId(verification_id);
    request.set_caller(caller.as_bytes().to_vec());

    cosmos_request.set_convertCredential(request);
    cosmos_request.write_to_bytes().unwrap()
}