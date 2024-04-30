use lazy_static::lazy_static;
use sgx_tstd::ffi::OsString;
use sgx_tstd::{env, sgxfs::SgxFile};
use sgx_types::{sgx_status_t, SgxResult};
use std::io::{Read, Write};
use std::vec::Vec;
use std::string::String;

use crate::error::Error;
use crate::key_manager::keys::{StateEncryptionKey, TransactionEncryptionKey};
use crate::encryption::{decrypt_deoxys, encrypt_deoxys};
use serde::{Deserialize, Deserializer, Serialize, Serializer};
use serde_json::Value;

pub mod keys;
pub mod utils;

pub const SEED_SIZE: usize = 32;
pub const SEED_FILENAME: &str = ".swtr_seed";
pub const PUBLIC_KEY_SIZE: usize = 32;
pub const PRIVATE_KEY_SIZE: usize = 32;

lazy_static! {
    pub static ref UNSEALED_KEY_MANAGER: Option<KeyManager> = KeyManager::unseal().ok();
    pub static ref SEED_HOME: OsString =
        env::var_os("SEED_HOME").unwrap_or_else(get_default_seed_home);
}

/// Handles initialization of a new seed node by creating and sealing master key to seed file
/// If `reset_flag` was set to `true`, it will rewrite existing seed file
pub fn init_master_key_inner(reset_flag: i32) -> sgx_status_t {
    // Check if master key exists
    let master_key_exists = match KeyManager::exists() {
        Ok(exists) => exists,
        Err(err) => {
            return err;
        }
    };

    // If master key does not exist or reset flag was set, generate random master key and seal it
    if !master_key_exists || reset_flag != 0 {
        // Generate random master key
        let key_manager = match KeyManager::random() {
            Ok(manager) => manager,
            Err(err) => return err,
        };

        // Seal master key
        match key_manager.seal() {
            Ok(_) => return sgx_status_t::SGX_SUCCESS,
            Err(err) => return err,
        };
    }

    sgx_status_t::SGX_SUCCESS
}

#[derive(Serialize, Deserialize)]
pub struct Epoch {
    epoch_key: [u8; 32],
    starting_block: u64
}

#[derive(Serialize, Deserialize)]
/// KeyManager handles keys sealing/unsealing and derivation.
/// * master_key - This key is used to derive keys for transaction and state encryption/decryption
pub struct KeyManager {
    epochs: Vec<Epoch>
}

impl KeyManager {
    /// Checks if file with sealed master key exists
    pub fn exists() -> SgxResult<bool> {
        match SgxFile::open(format!("{}/{}", SEED_HOME.to_str().unwrap(), SEED_FILENAME)) {
            Ok(_) => Ok(true),
            Err(ref err) if err.kind() == std::io::ErrorKind::NotFound => Ok(false),
            Err(err) => {
                println!(
                    "[KeyManager] Cannot check if sealed file exists. Reason: {:?}",
                    err
                );
                Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
            }
        }
    }

    /// Seals Key Manager to protected file, so it will be accessible only for enclave.
    /// For now, enclaves with same MRSIGNER will be able to recover that file, but
    /// we'll use MRENCLAVE since Upgradeability Protocol will be implemented
    pub fn seal(&self) -> SgxResult<()> {
        // Prepare file to write serialized key manager
        let seed_home_path = match SEED_HOME.to_str() {
            Some(path) => path,
            None => {
                println!("[KeyManager] Cannot get SEED_HOME env");
                return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
            }
        };

        let sealed_file_path = format!("{}/{}", seed_home_path, SEED_FILENAME);
        println!("[KeyManager] Creating file for key manager. Location: {:?}", sealed_file_path);
        let mut sealed_file = SgxFile::create(sealed_file_path).map_err(|err| {
            println!(
                "[KeyManager] Cannot create file for key manager. Reason: {:?}",
                err
            );
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        println!("[KeyManager] File created");

        // Write serialized key manager to the file
        let encoded = serde_json::to_string(&self).map_err(|err| {
            println!("[KeyManager] Cannot encode key manager to JSON. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;
        if let Err(err) = sealed_file.write(encoded.as_bytes()) {
            println!("[KeyManager] Cannot write serialized key manager. Reason: {:?}", err);
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }

        Ok(())
    }

    /// Unseals key manager from protected file. If file was not found or unaccessible,
    /// will return SGX_ERROR_UNEXPECTED
    pub fn unseal() -> SgxResult<Self> {
        println!(
            "[KeyManager] Sealed file location: {}/{}",
            SEED_HOME.to_str().unwrap(),
            SEED_FILENAME
        );

        // Unseal file with key manager
        let sealed_file_path = format!("{}/{}", SEED_HOME.to_str().unwrap(), SEED_FILENAME);
        let mut sealed_file = SgxFile::open(sealed_file_path).map_err(|err| {
            println!("[KeyManager] Cannot open file with key manager. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        let mut serialized_key_manager = String::new();
        if let Err(err) = sealed_file.read_to_string(&mut serialized_key_manager) {
            println!("[KeyManager] Cannot read serialized key manager. Reason: {:?}", err);
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        }

        let key_manager: KeyManager = serde_json::from_str(&serialized_key_manager).map_err(|err| {
            println!("[KeyManager] Cannot deserialize. Reason: {:?}", err);
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        Ok(key_manager)

        // Prepare buffer for key manager and read it from file
        // let mut master_key = [0u8; SEED_SIZE];
        // if let Err(err) = sealed_file.read(&mut master_key) {
        //     println!("[KeyManager] Cannot read master key file. Reason: {:?}", err);
        //     return Err(sgx_status_t::SGX_ERROR_UNEXPECTED)
        // }

        // TODO: Read and unseal from key manager

        // // Derive keys for transaction and state encryption
        // let tx_key_bytes = utils::derive_key(&master_key, b"TransactionEncryptionKeyV1");
        // let state_key_bytes = utils::derive_key(&master_key, b"StateEncryptionKeyV1");

        // Ok(Self {
        //     master_key,
        //     tx_key: TransactionEncryptionKey::from(tx_key_bytes),
        //     state_key: StateEncryptionKey::from(state_key_bytes),
        // })
    }

    /// Creates new KeyManager with random master key
    pub fn random() -> SgxResult<Self> {
        let master_key = utils::random_bytes32().map_err(|err| {
            println!("[KeyManager] Cannot create random master key. Reason: {:?}", err);
            err
        })?;

        // Derive keys for transaction and state encryption
        let tx_key_bytes = utils::derive_key(&master_key, b"TransactionEncryptionKeyV1");
        let state_key_bytes = utils::derive_key(&master_key, b"StateEncryptionKeyV1");

        Ok(Self {
            master_key,
            tx_key: TransactionEncryptionKey::from(tx_key_bytes),
            state_key: StateEncryptionKey::from(state_key_bytes),
        })
    }

    #[cfg(feature = "attestation_server")]
    /// Encrypts master key using shared key
    pub fn to_encrypted_master_key(
        &self,
        reg_key: &keys::RegistrationKey,
        public_key: Vec<u8>,
    ) -> Result<Vec<u8>, Error> {
        // Convert public key to appropriate format
        let public_key: [u8; 32] = match public_key.try_into() {
            Ok(public_key) => public_key,
            Err(_) => return Err(Error::decryption_err("Public key has wrong length")),
        };
        let public_key = x25519_dalek::PublicKey::from(public_key);

        // Derive shared secret
        let shared_secret = reg_key.diffie_hellman(public_key);

        // Encrypted master key
        let encrypted_value = encrypt_deoxys(shared_secret.as_bytes(), self.master_key.to_vec(), None)?;

        // Add public key as prefix
        let reg_public_key = reg_key.public_key();
        Ok([reg_public_key.as_bytes(), encrypted_value.as_slice()].concat())
    }

    /// Recovers encrypted master key obtained from seed exchange server
    pub fn from_encrypted_master_key(
        reg_key: &keys::RegistrationKey,
        public_key: Vec<u8>,
        encrypted_master_key: Vec<u8>,
    ) -> Result<Self, Error> {
        // Convert public key to appropriate format
        let public_key: [u8; 32] = match public_key.try_into() {
            Ok(public_key) => public_key,
            Err(_) => return Err(Error::encryption_err("Public key has wrong length")),
        };
        let public_key = x25519_dalek::PublicKey::from(public_key);

        // Derive shared secret
        let shared_secret = reg_key.diffie_hellman(public_key);

        // Decrypt master key
        let master_key = decrypt_deoxys(shared_secret.as_bytes(), encrypted_master_key)?;

        // Convert master key to appropriate format
        let master_key: [u8; 32] = match master_key.try_into() {
            Ok(master_key) => master_key,
            Err(_) => {
                return Err(Error::decryption_err("Master key has wrong length"));
            }
        };

        // Derive keys for transaction and state encryption
        let tx_key_bytes = utils::derive_key(&master_key, b"TransactionEncryptionKeyV1");
        let state_key_bytes = utils::derive_key(&master_key, b"StateEncryptionKeyV1");

        Ok(Self {
            master_key,
            tx_key: TransactionEncryptionKey::from(tx_key_bytes),
            state_key: StateEncryptionKey::from(state_key_bytes),
        })
    }

    /// Return x25519 public key for transaction encryption
    pub fn get_public_key(&self) -> Vec<u8> {
        self.tx_key.public_key()
    }


}

/// Tries to return path to $HOME/.swisstronik-enclave directory.
/// If it cannot find home directory, panics with error
fn get_default_seed_home() -> OsString {
    let home_dir = env::home_dir().expect("[KeyManager] Cannot find home directory");
    let default_seed_home = home_dir
        .to_str()
        .expect("[KeyManager] Cannot decode home directory path");
    OsString::from(format!("{}/.swisstronik-enclave", default_seed_home))
}
