use alloc::string::ToString;
use std::vec::Vec;
use sha3::{Keccak256, Digest};
use k256::{
    ecdsa::recoverable,
    elliptic_curve::sec1::ToEncodedPoint,
};
use evm::interpreter::runtime::Log as RuntimeLog;
use ethereum::Log;
use primitive_types::{H160, H256};
use protobuf::RepeatedField;
use crate::protobuf_generated::ffi::{AccessListItem, Topic};

pub fn recover_sender(msg: &H256, sig: &Vec<u8>) -> Option<[u8; 20]> {
    if sig.len() != 65 {
        return None;
    }

    let mut sig_buf = [0u8; 65];
    sig_buf.copy_from_slice(&sig);

    if sig_buf[64] > 26 {
        sig_buf[64] = sig[64] - 27
    }

    let signature = match recoverable::Signature::try_from(&sig_buf[..]) {
        Ok(signature) => signature,
        Err(err) => {
            println!("DEBUG failed to construct recoverable signature: {:?}", err.to_string());
            return None
        },
    };

    let recovered_key = match signature.recover_verifying_key_from_digest_bytes(msg.as_bytes().into()) {
        Ok(key) => key,
        Err(err) => {
            println!("DEBUG failed to recover verification key: {:?}", err.to_string());
            return None
        },
    };

    let public_key = recovered_key.to_encoded_point(false);
    let mut hasher = Keccak256::new();
    hasher.update(&public_key.as_bytes()[1..]); // Skip the compression byte
    let hash = hasher.finalize();

    let mut address = [0u8; 20];
    address.copy_from_slice(&hash[12..32]);
    Some(address)
}

pub fn parse_access_list(data: RepeatedField<AccessListItem>) -> Vec<(H160, Vec<H256>)> {
    let mut access_list = Vec::default();
    for access_list_item in data.to_vec() {
        let address = H160::from_slice(&access_list_item.address);
        let slots = access_list_item
            .storageSlot
            .to_vec()
            .into_iter()
            .map(|item| H256::from_slice(&item))
            .collect();

        access_list.push((address, slots));
    }

    access_list
}

pub fn convert_logs(input: Vec<RuntimeLog>) -> Vec<Log> {
    input
        .into_iter()
        .map(|rl| Log {
            address: rl.address,
            topics: rl.topics,
            data: rl.data,
        })
        .collect()
}

/// Converts EVM topic into protobuf-generated `Topic`
pub fn convert_topic_to_proto(topic: H256) -> Topic {
    let mut protobuf_topic = Topic::new();
    protobuf_topic.set_inner(topic.as_fixed_bytes().to_vec());

    protobuf_topic
}