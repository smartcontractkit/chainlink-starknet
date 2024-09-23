use starknet::ContractAddress;
use alexandria_bytes::{Bytes, BytesTrait};
use alexandria_encoding::sol_abi::sol_bytes::SolBytesTrait;
use alexandria_encoding::sol_abi::encode::SolAbiEncodeTrait;

#[derive(Copy, Drop, Serde)]
struct Call {
    target: ContractAddress,
    selector: felt252,
    data: Span<felt252>,
}

fn _hash_operation_batch(calls: Span<Call>, predecessor: u256, salt: u256) -> u256 {
    let mut encoded: Bytes = BytesTrait::new_empty();

    let mut i = 0;
    while i < calls
        .len() {
            let call = *calls.at(i);
            encoded = encoded.encode(call.target).encode(call.selector);
            let mut j = 0;
            while j < call.data.len() {
                encoded = encoded.encode(*call.data.at(j));
                j += 1;
            };
            i += 1;
        };

    encoded = encoded.encode(predecessor).encode(salt);
    encoded.keccak()
}

#[starknet::interface]
trait IRBACTimelock<TContractState> {
    fn schedule_batch(
        ref self: TContractState, calls: Span<Call>, predecessor: u256, salt: u256, delay: u256
    );
    fn cancel(ref self: TContractState, id: u256);
    fn execute_batch(ref self: TContractState, calls: Span<Call>, predecessor: u256, salt: u256);
    fn bypasser_execute_batch(ref self: TContractState, calls: Span<Call>);
    fn update_delay(ref self: TContractState, new_delay: u256);
    fn block_function_selector(ref self: TContractState, selector: felt252);
    fn unblock_function_selector(ref self: TContractState, selector: felt252);
    fn get_blocked_function_selector_count(self: @TContractState) -> u256;
    fn get_blocked_function_selector_at(self: @TContractState, index: u256) -> felt252;
    fn is_operation(self: @TContractState, id: u256) -> bool;
    fn is_operation_pending(self: @TContractState, id: u256) -> bool;
    fn is_operation_ready(self: @TContractState, id: u256) -> bool;
    fn is_operation_done(self: @TContractState, id: u256) -> bool;
    fn get_timestamp(self: @TContractState, id: u256) -> u256;
    fn get_min_delay(self: @TContractState) -> u256;
    fn hash_operation_batch(
        self: @TContractState, calls: Span<Call>, predecessor: u256, salt: u256
    ) -> u256;
}

// todo: add the erc receiver stuff + supports interface (register it for coin safe transfers)

#[starknet::contract]
mod RBACTimelock {
    use core::traits::TryInto;
    use core::starknet::SyscallResultTrait;
    use starknet::{ContractAddress, call_contract_syscall};
    use openzeppelin::{
        access::accesscontrol::AccessControlComponent, introspection::src5::SRC5Component,
        token::erc1155::erc1155_receiver::ERC1155ReceiverComponent,
        token::erc721::erc721_receiver::ERC721ReceiverComponent,
    };
    use chainlink::libraries::enumerable_set::EnumerableSetComponent;
    use super::{Call, _hash_operation_batch};
    use alexandria_bytes::{Bytes, BytesTrait};
    use alexandria_encoding::sol_abi::sol_bytes::SolBytesTrait;
    use alexandria_encoding::sol_abi::encode::SolAbiEncodeTrait;

    component!(path: SRC5Component, storage: src5, event: SRC5Event);
    component!(path: AccessControlComponent, storage: access_control, event: AccessControlEvent);
    component!(path: EnumerableSetComponent, storage: set, event: EnumerableSetEvent);
    component!(
        path: ERC1155ReceiverComponent, storage: erc1155_receiver, event: ERC1155ReceiverEvent
    );
    component!(path: ERC721ReceiverComponent, storage: erc721_receiver, event: ERC721ReceiverEvent);

    // SRC5
    #[abi(embed_v0)]
    impl SRC5Impl = SRC5Component::SRC5Impl<ContractState>;
    impl SRC5InternalImpl = SRC5Component::InternalImpl<ContractState>;

    // AccessControl
    #[abi(embed_v0)]
    impl AccessControlImpl =
        AccessControlComponent::AccessControlImpl<ContractState>;
    impl AccessControlInternalImpl = AccessControlComponent::InternalImpl<ContractState>;

    // ERC1155Receiver
    #[abi(embed_v0)]
    impl ERC1155ReceiverImpl =
        ERC1155ReceiverComponent::ERC1155ReceiverImpl<ContractState>;
    impl ERC1155ReceiverInternalImpl = ERC1155ReceiverComponent::InternalImpl<ContractState>;

    // ERC721Receiver
    #[abi(embed_v0)]
    impl ERC721ReceiverImpl =
        ERC721ReceiverComponent::ERC721ReceiverImpl<ContractState>;
    impl ERC721ReceiverInternalImpl = ERC721ReceiverComponent::InternalImpl<ContractState>;

    // EnumerableSet
    impl EnumerableSetInternalImpl = EnumerableSetComponent::InternalImpl<ContractState>;

    // we use sn_keccak intead of keccak256
    const ADMIN_ROLE: felt252 = selector!("ADMIN_ROLE");
    const PROPOSER_ROLE: felt252 = selector!("PROPOSER_ROLE");
    const EXECUTOR_ROLE: felt252 = selector!("EXECUTOR_ROLE");
    const CANCELLER_ROLE: felt252 = selector!("CANCELLER_ROLE");
    const BYPASSER_ROLE: felt252 = selector!("BYPASSER_ROLE");
    const _DONE_TIMESTAMP: u256 = 0x1;

    const BLOCKED_FUNCTIONS: u256 = 'BLOCKED_FUNCTION_SELECTORS';

    #[storage]
    struct Storage {
        #[substorage(v0)]
        erc721_receiver: ERC721ReceiverComponent::Storage,
        #[substorage(v0)]
        erc1155_receiver: ERC1155ReceiverComponent::Storage,
        #[substorage(v0)]
        set: EnumerableSetComponent::Storage,
        #[substorage(v0)]
        src5: SRC5Component::Storage,
        #[substorage(v0)]
        access_control: AccessControlComponent::Storage,
        // id -> timestamp
        _timestamps: LegacyMap<u256, u256>, // timestamp at which operation is ready to be executed
        _min_delay: u256
    }

    #[derive(Drop, starknet::Event)]
    struct MinDelayChange {
        old_duration: u256,
        new_duration: u256
    }

    #[derive(Drop, starknet::Event)]
    struct CallScheduled {
        #[key]
        id: u256,
        #[key]
        index: u256,
        predecessor: u256,
        salt: u256,
        delay: u256,
        target: ContractAddress,
        selector: felt252,
        data: Span<felt252>,
    }

    #[derive(Drop, starknet::Event)]
    struct Cancelled {
        #[key]
        id: u256
    }

    #[derive(Drop, starknet::Event)]
    struct CallExecuted {
        #[key]
        id: u256,
        #[key]
        index: u256,
        target: ContractAddress,
        selector: felt252,
        data: Span<felt252>,
    }

    #[derive(Drop, starknet::Event)]
    struct BypasserCallExecuted {
        #[key]
        index: u256,
        target: ContractAddress,
        selector: felt252,
        data: Span<felt252>,
    }

    #[derive(Drop, starknet::Event)]
    struct FunctionSelectorBlocked {
        #[key]
        selector: felt252
    }

    #[derive(Drop, starknet::Event)]
    struct FunctionSelectorUnblocked {
        #[key]
        selector: felt252
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        ERC721ReceiverEvent: ERC721ReceiverComponent::Event,
        #[flat]
        ERC1155ReceiverEvent: ERC1155ReceiverComponent::Event,
        #[flat]
        SRC5Event: SRC5Component::Event,
        #[flat]
        AccessControlEvent: AccessControlComponent::Event,
        #[flat]
        EnumerableSetEvent: EnumerableSetComponent::Event,
        MinDelayChange: MinDelayChange,
        CallScheduled: CallScheduled,
        Cancelled: Cancelled,
        CallExecuted: CallExecuted,
        BypasserCallExecuted: BypasserCallExecuted,
        FunctionSelectorBlocked: FunctionSelectorBlocked,
        FunctionSelectorUnblocked: FunctionSelectorUnblocked
    }


    #[constructor]
    fn constructor(
        ref self: ContractState,
        min_delay: u256,
        admin: ContractAddress,
        proposers: Array<ContractAddress>,
        executors: Array<ContractAddress>,
        cancellers: Array<ContractAddress>,
        bypassers: Array<ContractAddress>
    ) {
        self.access_control.initializer();
        self.erc1155_receiver.initializer();
        self.erc721_receiver.initializer();
        self.access_control._set_role_admin(ADMIN_ROLE, ADMIN_ROLE);
        self.access_control._set_role_admin(PROPOSER_ROLE, ADMIN_ROLE);
        self.access_control._set_role_admin(EXECUTOR_ROLE, ADMIN_ROLE);
        self.access_control._set_role_admin(CANCELLER_ROLE, ADMIN_ROLE);
        self.access_control._set_role_admin(BYPASSER_ROLE, ADMIN_ROLE);
        self.access_control._grant_role(ADMIN_ROLE, admin);

        let mut i = 0;
        while i < proposers
            .len() {
                self.access_control._grant_role(PROPOSER_ROLE, *proposers.at(i));
                i += 1;
            };

        let mut i = 0;
        while i < executors
            .len() {
                self.access_control._grant_role(EXECUTOR_ROLE, *executors.at(i));
                i += 1;
            };

        let mut i = 0;
        while i < cancellers
            .len() {
                self.access_control._grant_role(CANCELLER_ROLE, *cancellers.at(i));
                i += 1
            };

        let mut i = 0;
        while i < bypassers
            .len() {
                self.access_control._grant_role(BYPASSER_ROLE, *bypassers.at(i));
                i += 1
            };

        self._min_delay.write(min_delay);

        self
            .emit(
                Event::MinDelayChange(MinDelayChange { old_duration: 0, new_duration: min_delay, })
            )
    }

    #[abi(embed_v0)]
    impl RBACTimelockImpl of super::IRBACTimelock<ContractState> {
        fn schedule_batch(
            ref self: ContractState, calls: Span<Call>, predecessor: u256, salt: u256, delay: u256
        ) {
            self._assert_only_role_or_admin_role(PROPOSER_ROLE);

            let id = self.hash_operation_batch(calls, predecessor, salt);
            self._schedule(id, delay);

            let mut i = 0;
            while i < calls
                .len() {
                    let call = *calls.at(i);
                    assert(
                        !self.set.contains(BLOCKED_FUNCTIONS, call.selector.into()),
                        'selector is blocked'
                    );

                    self
                        .emit(
                            Event::CallScheduled(
                                CallScheduled {
                                    id: id,
                                    index: i.into(),
                                    target: call.target,
                                    selector: call.selector,
                                    data: call.data,
                                    predecessor: predecessor,
                                    salt: salt,
                                    delay: delay
                                }
                            )
                        );

                    i += 1;
                }
        }

        fn cancel(ref self: ContractState, id: u256) {
            self._assert_only_role_or_admin_role(CANCELLER_ROLE);

            assert(self.is_operation_pending(id), 'rbact: cant cancel operation');

            self._timestamps.write(id, 0);

            self.emit(Event::Cancelled(Cancelled { id: id }));
        }

        fn execute_batch(
            ref self: ContractState, calls: Span<Call>, predecessor: u256, salt: u256
        ) {
            self._assert_only_role_or_admin_role(EXECUTOR_ROLE);

            let id = self.hash_operation_batch(calls, predecessor, salt);

            self._before_call(id, predecessor);

            let mut i = 0;
            while i < calls
                .len() {
                    let call = *(calls.at(i));
                    self._execute(call);
                    self
                        .emit(
                            Event::CallExecuted(
                                CallExecuted {
                                    id: id,
                                    index: i.into(),
                                    target: call.target,
                                    selector: call.selector,
                                    data: call.data
                                }
                            )
                        );
                    i += 1;
                };

            self._after_call(id);
        }

        fn bypasser_execute_batch(ref self: ContractState, calls: Span<Call>) {
            self._assert_only_role_or_admin_role(BYPASSER_ROLE);

            let mut i = 0;
            while i < calls
                .len() {
                    let call = *calls.at(i);
                    self._execute(call);
                    self
                        .emit(
                            Event::BypasserCallExecuted(
                                BypasserCallExecuted {
                                    index: i.into(),
                                    target: call.target,
                                    selector: call.selector,
                                    data: call.data
                                }
                            )
                        );

                    i += 1;
                }
        }

        fn update_delay(ref self: ContractState, new_delay: u256) {
            self.access_control.assert_only_role(ADMIN_ROLE);

            self
                .emit(
                    Event::MinDelayChange(
                        MinDelayChange {
                            old_duration: self._min_delay.read(), new_duration: new_delay,
                        }
                    )
                );
            self._min_delay.write(new_delay);
        }

        //
        // ONLY ADMIN
        //

        fn block_function_selector(ref self: ContractState, selector: felt252) {
            self.access_control.assert_only_role(ADMIN_ROLE);

            // cast to u256 because that's what set stores 
            if self.set.add(BLOCKED_FUNCTIONS, selector.into()) {
                self
                    .emit(
                        Event::FunctionSelectorBlocked(
                            FunctionSelectorBlocked { selector: selector }
                        )
                    );
            }
        }

        fn unblock_function_selector(ref self: ContractState, selector: felt252) {
            self.access_control.assert_only_role(ADMIN_ROLE);

            if self.set.remove(BLOCKED_FUNCTIONS, selector.into()) {
                self
                    .emit(
                        Event::FunctionSelectorUnblocked(
                            FunctionSelectorUnblocked { selector: selector }
                        )
                    );
            }
        }

        //
        // VIEW ONLY
        //

        fn get_blocked_function_selector_count(self: @ContractState) -> u256 {
            self.set.length(BLOCKED_FUNCTIONS)
        }

        fn get_blocked_function_selector_at(self: @ContractState, index: u256) -> felt252 {
            // cast from u256 to felt252 should never error
            self.set.at(BLOCKED_FUNCTIONS, index).try_into().unwrap()
        }

        fn is_operation(self: @ContractState, id: u256) -> bool {
            self.get_timestamp(id) > 0
        }

        fn is_operation_pending(self: @ContractState, id: u256) -> bool {
            self.get_timestamp(id) > _DONE_TIMESTAMP
        }

        fn is_operation_ready(self: @ContractState, id: u256) -> bool {
            let timestamp = self.get_timestamp(id);
            timestamp > _DONE_TIMESTAMP && timestamp <= starknet::info::get_block_timestamp().into()
        }

        fn is_operation_done(self: @ContractState, id: u256) -> bool {
            self.get_timestamp(id) == _DONE_TIMESTAMP
        }

        fn get_timestamp(self: @ContractState, id: u256) -> u256 {
            self._timestamps.read(id)
        }

        fn get_min_delay(self: @ContractState) -> u256 {
            self._min_delay.read()
        }

        fn hash_operation_batch(
            self: @ContractState, calls: Span<Call>, predecessor: u256, salt: u256
        ) -> u256 {
            _hash_operation_batch(calls, predecessor, salt)
        }
    }


    #[generate_trait]
    impl InternalFunctions of InternalFunctionsTrait {
        fn _assert_only_role_or_admin_role(ref self: ContractState, role: felt252) {
            let caller = starknet::info::get_caller_address();
            if !self.access_control.has_role(ADMIN_ROLE, caller) {
                self.access_control.assert_only_role(role);
            }
        }

        fn _schedule(ref self: ContractState, id: u256, delay: u256) {
            assert(!self.is_operation(id), 'operation already scheduled');
            assert(delay >= self.get_min_delay(), 'insufficient delay');

            self._timestamps.write(id, starknet::info::get_block_timestamp().into() + delay)
        }

        fn _before_call(self: @ContractState, id: u256, predecessor: u256) {
            assert(self.is_operation_ready(id), 'rbact: operation not ready');
            assert(
                predecessor == 0 || self.is_operation_done(predecessor), 'rbact: missing dependency'
            );
        }

        fn _after_call(ref self: ContractState, id: u256) {
            assert(self.is_operation_ready(id), 'rbact: operation not ready');
            self._timestamps.write(id, _DONE_TIMESTAMP);
        }

        fn _execute(ref self: ContractState, call: Call) {
            let _response = call_contract_syscall(call.target, call.selector, call.data)
                .unwrap_syscall();
        }
    }
}
