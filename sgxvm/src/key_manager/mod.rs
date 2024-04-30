use lazy_static::lazy_static;
use sgx_tstd::ffi::OsString;
use sgx_tstd::{env, sgxfs::SgxFile};
use sgx_types::{sgx_status_t, SgxResult};
use std::io::{Read, Write};
use std::string::String;
use std::vec::Vec;

use crate::encryption::{decrypt_deoxys, encrypt_deoxys};
use crate::error::Error;
use crate::key_manager::epoch_manager::EpochManager;
use crate::key_manager::keys::{StateEncryptionKey, TransactionEncryptionKey};

pub mod consts;
pub mod epoch_manager;
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

/// KeyManager handles keys sealing/unsealing and derivation.
pub struct KeyManager {
    pub latest_tx_key: TransactionEncryptionKey,
    pub latest_state_key: StateEncryptionKey,
    epoch_manager: EpochManager,
}

impl KeyManager {
    pub fn get_state_key(&self, epoch: u16) -> Option<StateEncryptionKey> {
        match self.epoch_manager.get_epoch(epoch) {
            Some(epoch) => Some(epoch.get_state_key()),
            None => None
        }
    }

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
        println!(
            "[KeyManager] Creating file for key manager. Location: {:?}",
            sealed_file_path
        );
        let mut sealed_file = SgxFile::create(sealed_file_path).map_err(|err| {
            println!(
                "[KeyManager] Cannot create file for key manager. Reason: {:?}",
                err
            );
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;
        println!("[KeyManager] File created");

        let encoded = self.epoch_manager.serialize()?;
        if let Err(err) = sealed_file.write(encoded.as_bytes()) {
            println!(
                "[KeyManager] Cannot write serialized epoch manager. Reason: {:?}",
                err
            );
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
            println!(
                "[KeyManager] Cannot open file with key manager. Reason: {:?}",
                err
            );
            sgx_status_t::SGX_ERROR_UNEXPECTED
        })?;

        let mut serialized_epoch_manager = String::new();
        if let Err(err) = sealed_file.read_to_string(&mut serialized_epoch_manager) {
            println!(
                "[KeyManager] Cannot read serialized epoch manager. Reason: {:?}",
                err
            );
            return Err(sgx_status_t::SGX_ERROR_UNEXPECTED);
        }

        let epoch_manager = EpochManager::deserialize(&serialized_epoch_manager)?;
        let latest_epoch = epoch_manager.get_latest_epoch()?;
        let latest_tx_key = latest_epoch.get_tx_key();
        let latest_state_key = latest_epoch.get_state_key();

        Ok(Self {
            epoch_manager,
            latest_tx_key,
            latest_state_key,
        })
    }

    /// Creates new KeyManager with signle random epoch key
    pub fn random() -> SgxResult<Self> {
        let random_epoch_manager = EpochManager::random_with_single_epoch()?;
        let latest_epoch = random_epoch_manager.get_latest_epoch()?;
        let latest_tx_key = latest_epoch.get_tx_key();
        let latest_state_key = latest_epoch.get_state_key();

        Ok(Self {
            epoch_manager: random_epoch_manager,
            latest_tx_key,
            latest_state_key,
        })
    }

    #[cfg(feature = "attestation_server")]
    /// Encrypts epoch data using shared key
    pub fn encrypt_epoch_data(
        &self,
        reg_key: &keys::RegistrationKey,
        public_key: Vec<u8>,
    ) -> SgxResult<Vec<u8>> {
        self.epoch_manager.encrypt(reg_key, public_key)
    }

    /// Recovers encrypted epoch data, obtained from attestation server
    pub fn decrypt_epoch_data(
        reg_key: &keys::RegistrationKey,
        public_key: Vec<u8>,
        encrypted_epoch_data: Vec<u8>,
    ) -> SgxResult<Self> {
        let epoch_manager = EpochManager::decrypt(reg_key, public_key, encrypted_epoch_data)?;
        let latest_epoch = epoch_manager.get_latest_epoch()?;
        let latest_tx_key = latest_epoch.get_tx_key();
        let latest_state_key = latest_epoch.get_state_key();

        Ok(Self {
            epoch_manager,
            latest_tx_key,
            latest_state_key,
        })
    }

    /// Return x25519 public key for transaction encryption.
    pub fn get_public_key(&self, block_number: Option<u64>) -> Vec<u8> {
        match block_number {
            Some(block_number) => Vec::default(), // TODO: Implement
            None => self.latest_tx_key.public_key(),
        }
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
