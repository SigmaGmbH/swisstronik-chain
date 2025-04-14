use alloc::boxed::Box;
use alloc::collections::{BTreeMap, BTreeSet};
use alloc::string::ToString;
use alloc::vec::Vec;
use core::mem;
use evm::backend::{OverlayedChangeSet, RuntimeBackend, RuntimeBaseBackend, RuntimeEnvironment};
use evm::interpreter::error::{ExitError, ExitException, ExitResult, ExitSucceed};
use evm::interpreter::runtime::{Log, SetCodeOrigin};
use evm::{MergeStrategy, TransactionalBackend};
use primitive_types::{H160, H256, U256};
use crate::{coder, querier};
use crate::protobuf_generated::ffi;
use crate::vm::storage::StorageWithQuerier;

pub struct TxEnvironment {
    pub chain_id: U256,
    pub gas_price: U256,
    pub block_number: U256,
    pub timestamp: U256,
    pub block_gas_limit: U256,
    pub block_base_fee_per_gas: U256,
    pub block_coinbase: H160,
}

impl From<ffi::TransactionContext> for TxEnvironment {
    fn from(context: ffi::TransactionContext) -> Self {
        Self {
            chain_id: U256::from(context.chain_id),
            gas_price: U256::from_big_endian(&context.gas_price),
            block_number: U256::from(context.block_number),
            timestamp: U256::from(context.timestamp),
            block_gas_limit: U256::from(context.block_gas_limit),
            block_base_fee_per_gas: U256::from_big_endian(&context.block_base_fee_per_gas),
            block_coinbase: H160::from_slice(&context.block_coinbase),
        }
    }
}

pub struct Backend<'state> {
    querier: *mut querier::GoQuerier,
    storage: &'state StorageWithQuerier,
    environment: TxEnvironment,
    substate: Box<Substate>,
    accessed: BTreeSet<(H160, Option<H256>)>,
}

impl<'state> Backend<'state> {
    pub fn new(
        querier: *mut querier::GoQuerier,
        storage: &'state StorageWithQuerier,
        environment: TxEnvironment,
    ) -> Self {
        Self {
            querier,
            storage,
            substate: Box::new(Substate::new()),
            environment,
            accessed: BTreeSet::new(),
        }
    }

    pub fn apply_changeset(storage: &'state StorageWithQuerier, changeset: &OverlayedChangeSet) -> ExitResult {
        for (address, balance) in changeset.balances.clone() {
            storage.insert_account_balance(&address, &balance).map_err(|err| ExitException::Other(err.to_string().into()))?
        }

        for (address, nonce) in changeset.nonces.clone() {
            storage.insert_account_nonce(&address, &nonce).map_err(|err| ExitException::Other(err.to_string().into()))?
        }

        for (address, code) in changeset.codes.clone() {
            storage.insert_account_code(address, code).map_err(|err| ExitException::Other(err.to_string().into()))?
        }

        for ((address, key), value) in changeset.storages.clone() {
            storage.insert_storage_cell(address, key, value).map_err(|err| ExitException::Other(err.to_string().into()))?;
        }

        for address in changeset.deletes.clone() {
            storage.remove(&address).map_err(|err| ExitException::Other(err.to_string().into()))?;
        }

        Ok(ExitSucceed::Returned)
    }

    pub fn deconstruct(self) -> OverlayedChangeSet {
        OverlayedChangeSet {
                logs: self.substate.logs,
                balances: self.substate.balances,
                codes: self.substate.codes,
                nonces: self.substate.nonces,
                storage_resets: self.substate.storage_resets,
                storages: self.substate.storages,
                transient_storage: self.substate.transient_storage,
                deletes: self.substate.deletes,
        }
    }
}

impl<'state> RuntimeEnvironment for Backend<'state> {
    fn block_hash(&self, number: U256) -> H256 {
        let encoded_request = coder::encode_query_block_hash(&number);
        match querier::make_request(self.querier, encoded_request) {
            Some(result) => {
                // Decode protobuf
                let decoded_result = match protobuf::parse_from_bytes::<ffi::QueryBlockHashResponse>(
                    result.as_slice(),
                ) {
                    Ok(res) => res,
                    Err(err) => {
                        println!("Cannot decode protobuf response: {:?}", err);
                        return H256::default();
                    }
                };
                H256::from_slice(decoded_result.hash.as_slice())
            }
            None => {
                println!("Get block hash failed. Empty response");
                H256::default()
            }
        }
    }

    fn block_number(&self) -> U256 {
        self.environment.block_number
    }

    fn block_coinbase(&self) -> H160 {
        self.environment.block_coinbase
    }

    fn block_timestamp(&self) -> U256 {
        self.environment.timestamp
    }

    fn block_difficulty(&self) -> U256 {
        U256::zero()
    }

    fn block_randomness(&self) -> Option<H256> {
        None
    }

    fn block_gas_limit(&self) -> U256 {
        self.environment.block_gas_limit
    }

    fn block_base_fee_per_gas(&self) -> U256 {
        self.environment.block_base_fee_per_gas
    }

    fn chain_id(&self) -> U256 {
        self.environment.chain_id
    }
}

impl<'state> RuntimeBaseBackend for Backend<'state> {
    fn balance(&self, address: H160) -> U256 {
        if let Some(balance) = self.substate.known_balance(address) {
            balance
        } else {
            self.storage.get_account(&address).0
        }
    }

    fn code_size(&self, address: H160) -> U256 {
        self.storage.get_account_code_size(&address).unwrap_or(U256::zero())
    }

    fn code_hash(&self, address: H160) -> H256 {
        self.storage.get_account_code_hash(&address).unwrap_or(H256::default())
    }

    fn code(&self, address: H160) -> Vec<u8> {
        if let Some(code) = self.substate.known_code(address) {
            code
        } else {
            self.storage.get_account_code(&address).unwrap_or(Vec::new())
        }
    }

    fn storage(&self, address: H160, index: H256) -> H256 {
        if let Some(value) = self.substate.known_storage(address, index) {
            value
        } else {
            self.storage.get_account_storage_cell(&address, &index).unwrap_or(H256::default())
        }
    }

    fn transient_storage(&self, address: H160, index: H256) -> H256 {
        if let Some(value) = self.substate.known_transient_storage(address, index) {
            value
        } else {
            H256::default()
        }
    }

    fn exists(&self, address: H160) -> bool {
        if let Some(exists) = self.substate.known_exists(address) {
            exists
        } else {
            self.storage.contains_key(&address)
        }
    }

    fn nonce(&self, address: H160) -> U256 {
        self.storage.get_account(&address).1
    }
}

impl<'state> RuntimeBackend for Backend<'state> {
    fn original_storage(&self, address: H160, index: H256) -> H256 {
        self.storage(address, index)
    }

    fn deleted(&self, address: H160) -> bool {
        self.substate.deleted(address)
    }

    fn is_cold(&self, address: H160, index: Option<H256>) -> bool {
        !self.accessed.contains(&(address, index))
    }

    fn mark_hot(&mut self, address: H160, index: Option<H256>) {
        self.accessed.insert((address, index));
    }

    fn set_storage(&mut self, address: H160, index: H256, value: H256) -> Result<(), ExitError> {
        self.substate.storages.insert((address, index), value);
        Ok(())
    }

    fn set_transient_storage(
        &mut self,
        address: H160,
        index: H256,
        value: H256,
    ) -> Result<(), ExitError> {
        self.substate
            .transient_storage
            .insert((address, index), value);
        Ok(())
    }

    fn log(&mut self, log: Log) -> Result<(), ExitError> {
        self.substate.logs.push(log);
        Ok(())
    }

    fn mark_delete(&mut self, address: H160) {
        self.substate.deletes.insert(address);
    }

    fn reset_storage(&mut self, address: H160) {
        self.substate.storage_resets.insert(address);
    }

    fn set_code(
        &mut self,
        address: H160,
        code: Vec<u8>,
        _origin: SetCodeOrigin,
    ) -> Result<(), ExitError> {
        self.substate.codes.insert(address, code);
        Ok(())
    }

    fn reset_balance(&mut self, address: H160) {
        self.substate.balances.insert(address, U256::zero());
    }

    fn deposit(&mut self, target: H160, value: U256) {
        if value == U256::zero() {
            return;
        }

        let current_balance = self.balance(target);
        self.substate
            .balances
            .insert(target, current_balance.saturating_add(value));
    }

    fn withdrawal(&mut self, source: H160, value: U256) -> Result<(), ExitError> {
        if value == U256::zero() {
            return Ok(());
        }

        let current_balance = self.balance(source);
        if current_balance < value {
            return Err(ExitException::OutOfFund.into());
        }
        let new_balance = current_balance - value;
        self.substate.balances.insert(source, new_balance);
        Ok(())
    }

    fn inc_nonce(&mut self, address: H160) -> Result<(), ExitError> {
        let new_nonce = self.nonce(address).saturating_add(U256::from(1));
        self.substate.nonces.insert(address, new_nonce);
        Ok(())
    }
}

impl<'state> TransactionalBackend for Backend<'state> {
    fn push_substate(&mut self) {
        let mut parent = Box::new(Substate::new());
        mem::swap(&mut parent, &mut self.substate);
        self.substate.parent = Some(parent);
    }

    fn pop_substate(&mut self, strategy: MergeStrategy) {
        let mut child = self.substate.parent.take().expect("uneven substate pop");
        mem::swap(&mut child, &mut self.substate);
        let child = child;

        match strategy {
            MergeStrategy::Commit => {
                for log in child.logs {
                    self.substate.logs.push(log);
                }
                for (address, balance) in child.balances {
                    self.substate.balances.insert(address, balance);
                }
                for (address, code) in child.codes {
                    self.substate.codes.insert(address, code);
                }
                for (address, nonce) in child.nonces {
                    self.substate.nonces.insert(address, nonce);
                }
                for address in child.storage_resets {
                    self.substate.storage_resets.insert(address);
                }
                for ((address, key), value) in child.storages {
                    self.substate.storages.insert((address, key), value);
                }
                for ((address, key), value) in child.transient_storage {
                    self.substate
                        .transient_storage
                        .insert((address, key), value);
                }
                for address in child.deletes {
                    self.substate.deletes.insert(address);
                }
            }
            MergeStrategy::Revert | MergeStrategy::Discard => {}
        }
    }
}

struct Substate {
    parent: Option<Box<Substate>>,
    logs: Vec<Log>,
    balances: BTreeMap<H160, U256>,
    codes: BTreeMap<H160, Vec<u8>>,
    nonces: BTreeMap<H160, U256>,
    storage_resets: BTreeSet<H160>,
    storages: BTreeMap<(H160, H256), H256>,
    transient_storage: BTreeMap<(H160, H256), H256>,
    deletes: BTreeSet<H160>,
}

impl Substate {
    pub fn new() -> Self {
        Self {
            parent: None,
            logs: Vec::new(),
            balances: Default::default(),
            codes: Default::default(),
            nonces: Default::default(),
            storage_resets: Default::default(),
            storages: Default::default(),
            transient_storage: Default::default(),
            deletes: Default::default(),
        }
    }

    pub fn known_balance(&self, address: H160) -> Option<U256> {
        if let Some(balance) = self.balances.get(&address) {
            Some(*balance)
        } else if let Some(parent) = self.parent.as_ref() {
            parent.known_balance(address)
        } else {
            None
        }
    }

    pub fn known_code(&self, address: H160) -> Option<Vec<u8>> {
        if let Some(code) = self.codes.get(&address) {
            Some(code.clone())
        } else if let Some(parent) = self.parent.as_ref() {
            parent.known_code(address)
        } else {
            None
        }
    }

    pub fn known_storage(&self, address: H160, key: H256) -> Option<H256> {
        if let Some(value) = self.storages.get(&(address, key)) {
            Some(*value)
        } else if self.storage_resets.contains(&address) {
            Some(H256::default())
        } else if let Some(parent) = self.parent.as_ref() {
            parent.known_storage(address, key)
        } else {
            None
        }
    }

    pub fn known_transient_storage(&self, address: H160, key: H256) -> Option<H256> {
        if let Some(value) = self.transient_storage.get(&(address, key)) {
            Some(*value)
        } else if let Some(parent) = self.parent.as_ref() {
            parent.known_transient_storage(address, key)
        } else {
            None
        }
    }

    pub fn known_exists(&self, address: H160) -> Option<bool> {
        if self.balances.contains_key(&address)
            || self.nonces.contains_key(&address)
            || self.codes.contains_key(&address)
        {
            Some(true)
        } else if let Some(parent) = self.parent.as_ref() {
            parent.known_exists(address)
        } else {
            None
        }
    }

    pub fn deleted(&self, address: H160) -> bool {
        if self.deletes.contains(&address) {
            true
        } else if let Some(parent) = self.parent.as_ref() {
            parent.deleted(address)
        } else {
            false
        }
    }
}