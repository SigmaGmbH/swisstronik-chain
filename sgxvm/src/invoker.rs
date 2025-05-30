use alloc::{rc::Rc, vec::Vec};
use core::{cmp::min, convert::Infallible};
use core::cell::RefCell;
use evm::interpreter::error::{ExitError, ExitResult, ExitException};
use evm::{MergeStrategy, TransactionalBackend};
use evm::standard::{routines, Config, Resolver, TransactArgs, TransactValue, InvokerState, SubstackInvoke};
use evm::{Invoker, InvokerControl};
use evm::interpreter::{
    error::{
        CallCreateTrap, CallCreateTrapData, CallTrapData,
        Capture, CreateScheme, TrapConsume,
    },
    opcode::Opcode,
    runtime::{
        Context, GasState, RuntimeBackend, RuntimeEnvironment, RuntimeState, SetCodeOrigin,
        TransactionContext, Transfer,
    },
    Interpreter,
};
use primitive_types::{H160, H256, U256};
use sha3::{Digest, Keccak256};

/// A trap that can be turned into either a call/create trap (where we push new
/// call stack), or an interrupt (an external signal).
pub trait IntoCallCreateTrap {
    /// An external signal.
    type Interrupt;

    /// Turn the current trap into either a call/create trap or an interrupt.
    fn into_call_create_trap(self) -> Result<Opcode, Self::Interrupt>;
}

impl IntoCallCreateTrap for Opcode {
    type Interrupt = Infallible;

    fn into_call_create_trap(self) -> Result<Opcode, Infallible> {
        Ok(self)
    }
}

/// The invoke used in a top-layer transaction stack.
pub struct TransactInvoke {
    pub create_address: Option<H160>,
    pub gas_limit: U256,
    pub gas_price: U256,
    pub caller: H160,
}

pub struct DataContainer {
    pub gas_used: U256,
    pub return_value: Vec<u8>,
}

impl Default for DataContainer {
    fn default() -> Self {
        Self {
            gas_used: U256::from(21000),
            return_value: vec![],
        }
    }
}

/// Overlayed Invoker.
///
/// The generic parameters are as follows:
/// * `S`: The runtime state, usually [RuntimeState] but can be customized.
/// * `H`: Backend type.
/// * `R`: Code resolver type, also handle precompiles. Usually
///   [EtableResolver] but can be customized.
/// * `Tr`: Trap type, usually [crate::Opcode] but can be customized.
pub struct OverlayedInvoker<'config, 'resolver, R> {
    container: RefCell<Option<DataContainer>>,
    config: &'config Config,
    resolver: &'resolver R,
}

impl<'config, 'resolver, R> OverlayedInvoker<'config, 'resolver, R> {
    /// Create a new standard invoker with the given config and resolver.
    pub fn new(config: &'config Config, resolver: &'resolver R) -> Self {
        Self { config, resolver, container: RefCell::new(None) }
    }

    pub fn get_gas_used(&self) -> Option<U256> {
        self.container.borrow().as_ref().map(|data| data.gas_used)
    }

    pub fn get_return_value(&self) -> Option<Vec<u8>> {
        self.container.borrow().as_ref().map(|data| data.return_value.clone())
    }
}

impl<'config, 'resolver, H, R, Tr> Invoker<H, Tr> for OverlayedInvoker<'config, 'resolver, R>
where
    R::State: InvokerState<'config> + AsRef<RuntimeState> + AsMut<RuntimeState>,
    H: RuntimeEnvironment + RuntimeBackend + TransactionalBackend,
    R: Resolver<H>,
    Tr: TrapConsume<CallCreateTrap>,
{
    type State = R::State;
    type Interpreter = R::Interpreter;
    type Interrupt = Tr::Rest;
    type TransactArgs = TransactArgs;
    type TransactInvoke = TransactInvoke;
    type TransactValue = TransactValue;
    type SubstackInvoke = SubstackInvoke;

    fn new_transact(
        &self,
        args: TransactArgs,
        handler: &mut H,
    ) -> Result<
        (
            TransactInvoke,
            InvokerControl<Self::Interpreter, (ExitResult, (R::State, Vec<u8>))>,
        ),
        ExitError,
    > {
        let caller = args.caller();
        let gas_price = args.gas_price();

        handler.inc_nonce(caller)?;

        let address = match &args {
            TransactArgs::Call { address, .. } => *address,
            TransactArgs::Create {
                caller,
                salt,
                init_code,
                ..
            } => match salt {
                Some(salt) => {
                    let scheme = CreateScheme::Create2 {
                        caller: *caller,
                        code_hash: H256::from_slice(Keccak256::digest(init_code).as_slice()),
                        salt: *salt,
                    };
                    scheme.address(handler)
                }
                None => {
                    let scheme = CreateScheme::Legacy { caller: *caller };
                    scheme.address(handler)
                }
            },
        };
        let value = args.value();

        let invoke = TransactInvoke {
            gas_limit: args.gas_limit(),
            gas_price: args.gas_price(),
            caller: args.caller(),
            create_address: match &args {
                TransactArgs::Call { .. } => None,
                TransactArgs::Create { .. } => Some(address),
            },
        };

        handler.push_substate();

        let context = Context {
            caller,
            address,
            apparent_value: value,
        };
        let transaction_context = TransactionContext {
            origin: caller,
            gas_price,
        };
        let transfer = Transfer {
            source: caller,
            target: address,
            value,
        };
        let runtime_state = RuntimeState {
            context,
            transaction_context: Rc::new(transaction_context),
            retbuf: Vec::new(),
        };

        let work = || -> Result<(TransactInvoke, _), ExitError> {
            match args {
                TransactArgs::Call {
                    caller,
                    address,
                    data,
                    gas_limit,
                    access_list,
                    ..
                } => {
                    for (address, keys) in &access_list {
                        handler.mark_hot(*address, None);
                        for key in keys {
                            handler.mark_hot(*address, Some(*key));
                        }
                    }

                    let state = <R::State>::new_transact_call(
                        runtime_state,
                        gas_limit,
                        &data,
                        &access_list,
                        self.config,
                    )?;

                    let machine = routines::make_enter_call_machine(
                        self.config,
                        self.resolver,
                        address,
                        data,
                        Some(transfer),
                        state,
                        handler,
                    )?;

                    if self.config.increase_state_access_gas {
                        if self.config.warm_coinbase_address {
                            let coinbase = handler.block_coinbase();
                            handler.mark_hot(coinbase, None);
                        }
                        handler.mark_hot(caller, None);
                        handler.mark_hot(address, None);
                    }

                    Ok((invoke, machine))
                }
                TransactArgs::Create {
                    caller,
                    init_code,
                    gas_limit,
                    access_list,
                    ..
                } => {
                    let state = <R::State>::new_transact_create(
                        runtime_state,
                        gas_limit,
                        &init_code,
                        &access_list,
                        self.config,
                    )?;

                    let machine = routines::make_enter_create_machine(
                        self.config,
                        self.resolver,
                        caller,
                        init_code,
                        transfer,
                        state,
                        handler,
                    )?;

                    Ok((invoke, machine))
                }
            }
        };

        work().map_err(|err| {
            handler.pop_substate(MergeStrategy::Discard);
            err
        })
    }

    fn finalize_transact(
        &self,
        invoke: &TransactInvoke,
        result: ExitResult,
        (mut substate, retval): (R::State, Vec<u8>),
        handler: &mut H,
    ) -> Result<TransactValue, ExitError> {
        // Since retval is moved into closure, we clone it here
        let retval_copy = retval.clone();

        let work = || -> Result<TransactValue, ExitError> {
            match result {
                Ok(result) => {
                    if let Some(address) = invoke.create_address {
                        let retbuf = retval;

                        routines::deploy_create_code(
                            self.config,
                            address,
                            retbuf,
                            &mut substate,
                            handler,
                            SetCodeOrigin::Transaction,
                        )?;

                        Ok(TransactValue::Create {
                            succeed: result,
                            address,
                        })
                    } else {
                        Ok(TransactValue::Call {
                            succeed: result,
                            retval,
                        })
                    }
                }
                Err(result) => Err(result),
            }
        };

        let result = work();

        match &result {
            Ok(_) => {
                handler.pop_substate(MergeStrategy::Commit);
            }
            Err(_) => {
                handler.pop_substate(MergeStrategy::Discard);
            }
        }

        let used_gas = invoke.gas_limit.saturating_sub(substate.effective_gas());
        *self.container.borrow_mut() = Some(DataContainer{
            gas_used: used_gas,
            return_value: retval_copy,
        });

        result
    }

    fn enter_substack(
        &self,
        trap: Tr,
        machine: &mut Self::Interpreter,
        handler: &mut H,
        depth: usize,
    ) -> Capture<
        Result<
            (
                SubstackInvoke,
                InvokerControl<Self::Interpreter, (ExitResult, (R::State, Vec<u8>))>,
            ),
            ExitError,
        >,
        Self::Interrupt,
    > {
        fn l64(gas: U256) -> U256 {
            gas - gas / U256::from(64)
        }

        let opcode = match trap.consume() {
            Ok(opcode) => opcode,
            Err(interrupt) => return Capture::Trap(interrupt),
        };

        if depth >= self.config.call_stack_limit {
            return Capture::Exit(Err(ExitException::CallTooDeep.into()));
        }

        let trap_data = match CallCreateTrapData::new_from(opcode, machine.machine_mut()) {
            Ok(trap_data) => trap_data,
            Err(err) => return Capture::Exit(Err(err)),
        };

        let after_gas = if self.config.call_l64_after_gas {
            l64(machine.machine().state.gas())
        } else {
            machine.machine().state.gas()
        };
        let target_gas = trap_data.target_gas().unwrap_or(after_gas);
        let gas_limit = min(after_gas, target_gas);

        let call_has_value =
            matches!(&trap_data, CallCreateTrapData::Call(call) if call.has_value());

        let is_static = if machine.machine().state.is_static() {
            true
        } else {
            match &trap_data {
                CallCreateTrapData::Call(CallTrapData { is_static, .. }) => *is_static,
                _ => false,
            }
        };

        let transaction_context = machine.machine().state.as_ref().transaction_context.clone();

        match trap_data {
            CallCreateTrapData::Call(call_trap_data) => {
                let substate = match machine.machine_mut().state.substate(
                    RuntimeState {
                        context: call_trap_data.context.clone(),
                        transaction_context,
                        retbuf: Vec::new(),
                    },
                    gas_limit,
                    is_static,
                    call_has_value,
                ) {
                    Ok(submeter) => submeter,
                    Err(err) => return Capture::Exit(Err(err)),
                };

                let target = call_trap_data.target;

                Capture::Exit(routines::enter_call_substack(
                    self.config,
                    self.resolver,
                    call_trap_data,
                    target,
                    substate,
                    handler,
                ))
            }
            CallCreateTrapData::Create(create_trap_data) => {
                let caller = create_trap_data.scheme.caller();
                let address = create_trap_data.scheme.address(handler);
                let code = create_trap_data.code.clone();

                let substate = match machine.machine_mut().state.substate(
                    RuntimeState {
                        context: Context {
                            address,
                            caller,
                            apparent_value: create_trap_data.value,
                        },
                        transaction_context,
                        retbuf: Vec::new(),
                    },
                    gas_limit,
                    is_static,
                    call_has_value,
                ) {
                    Ok(submeter) => submeter,
                    Err(err) => return Capture::Exit(Err(err)),
                };

                Capture::Exit(routines::enter_create_substack(
                    self.config,
                    self.resolver,
                    code,
                    create_trap_data,
                    substate,
                    handler,
                ))
            }
        }
    }

    fn exit_substack(
        &self,
        result: ExitResult,
        (mut substate, retval): (R::State, Vec<u8>),
        trap_data: Self::SubstackInvoke,
        parent: &mut Self::Interpreter,
        handler: &mut H,
    ) -> Result<(), ExitError> {
        let strategy = match &result {
            Ok(_) => MergeStrategy::Commit,
            Err(ExitError::Reverted) => MergeStrategy::Revert,
            Err(_) => MergeStrategy::Discard,
        };

        match trap_data {
            SubstackInvoke::Create { address, trap } => {
                let retbuf = retval;
                let caller = trap.scheme.caller();

                let result = result.and_then(|_| {
                    routines::deploy_create_code(
                        self.config,
                        address,
                        retbuf.clone(),
                        &mut substate,
                        handler,
                        SetCodeOrigin::Subcall(caller),
                    )?;

                    Ok(address)
                });

                parent.machine_mut().state.merge(substate, strategy);
                handler.pop_substate(strategy);

                trap.feedback(result, retbuf, parent)?;

                Ok(())
            }
            SubstackInvoke::Call { trap } => {
                let retbuf = retval;

                parent.machine_mut().state.merge(substate, strategy);
                handler.pop_substate(strategy);

                trap.feedback(result, retbuf, parent)?;

                Ok(())
            }
        }
    }
}
