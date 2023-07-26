use starknet::EthAddress;

#[starknet::interface]
trait ISequencerUptimeFeed<TContractState> {
    fn l1_sender(self: @TContractState) -> EthAddress;
    fn set_l1_sender(ref self: TContractState, address: EthAddress);
}

#[starknet::contract]
mod SequencerUptimeFeed {
    use starknet::EthAddress;
    use starknet::EthAddressSerde;
    use starknet::EthAddressIntoFelt252;
    use starknet::Felt252TryIntoEthAddress;
    use starknet::EthAddressZeroable;
    use starknet::ContractAddress;
    use starknet::StorageAccess;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;
    use starknet::class_hash::ClassHash;

    use box::BoxTrait;
    use traits::Into;
    use traits::TryInto;
    use option::OptionTrait;
    use zeroable::Zeroable;

    use chainlink::libraries::ownable::{Ownable, IOwnable};
    use chainlink::libraries::access_control::{AccessControl, IAccessController};
    use chainlink::ocr2::aggregator::Round;
    use chainlink::ocr2::aggregator::IAggregator;
    use chainlink::ocr2::aggregator::{Transmission};
    use chainlink::libraries::upgradeable::Upgradeable;

    #[storage]
    struct Storage {
        // l1 sender is an starknet validator ethereum address
        _l1_sender: felt252,
        // maps round id to round transmission
        _round_transmissions: LegacyMap<u128, Transmission>,
        _latest_round_id: u128,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        RoundUpdated: RoundUpdated,
        NewRound: NewRound,
        AnswerUpdated: AnswerUpdated,
        UpdateIgnored: UpdateIgnored,
        L1SenderTransferred: L1SenderTransferred,
    }

    #[derive(Drop, starknet::Event)]
    struct RoundUpdated {
        status: u128,
        updated_at: u64
    }

    #[derive(Drop, starknet::Event)]
    struct NewRound {
        round_id: u128,
        started_by: ContractAddress,
        started_at: u64
    }

    #[derive(Drop, starknet::Event)]
    struct AnswerUpdated {
        current: u128,
        round_id: u128,
        timestamp: u64
    }

    #[derive(Drop, starknet::Event)]
    struct UpdateIgnored {
        latest_status: u128,
        latest_timestamp: u64,
        incoming_status: u128,
        incoming_timestamp: u64
    }

    #[derive(Drop, starknet::Event)]
    struct L1SenderTransferred {
        from_address: EthAddress,
        to_address: EthAddress
    }

    #[external(v0)]
    impl AggregatorImpl of IAggregator<ContractState> {
        fn latest_round_data(self: @ContractState) -> Round {
            self._require_read_access();
            let latest_round_id = self._latest_round_id.read();
            let round_transmission = self._round_transmissions.read(latest_round_id);
            Round {
                round_id: latest_round_id.into(),
                answer: round_transmission.answer,
                block_num: round_transmission.block_num,
                started_at: round_transmission.observation_timestamp,
                updated_at: round_transmission.transmission_timestamp,
            }
        }

        fn round_data(self: @ContractState, round_id: u128) -> Round {
            self._require_read_access();
            assert(round_id < self._latest_round_id.read(), 'invalid round id');
            let round_transmission = self._round_transmissions.read(round_id);
            Round {
                round_id: round_id.into(),
                answer: round_transmission.answer,
                block_num: round_transmission.block_num,
                started_at: round_transmission.observation_timestamp,
                updated_at: round_transmission.transmission_timestamp,
            }
        }

        fn description(self: @ContractState) -> felt252 {
            'L2 Sequencer Uptime Status Feed'
        }

        fn decimals(self: @ContractState) -> u8 {
            0_u8
        }

        fn type_and_version(self: @ContractState) -> felt252 {
            'SequencerUptimeFeed 1.0.0'
        }
    }

    #[constructor]
    fn constructor(ref self: ContractState, initial_status: u128, owner_address: ContractAddress) {
        self._initializer(initial_status, owner_address);
    }

    #[l1_handler]
    fn update_status(ref self: ContractState, from_address: felt252, status: u128, timestamp: u64) {
        assert(self._l1_sender.read() == from_address, 'EXPECTED_FROM_BRIDGE_ONLY');

        let latest_round_id = self._latest_round_id.read();
        let latest_round = self._round_transmissions.read(latest_round_id);

        if timestamp <= latest_round.observation_timestamp {
            self
                .emit(
                    Event::UpdateIgnored(
                        UpdateIgnored {
                            latest_status: latest_round.answer,
                            latest_timestamp: latest_round.transmission_timestamp,
                            incoming_status: status,
                            incoming_timestamp: timestamp
                        }
                    )
                );
            return ();
        }

        if latest_round.answer == status {
            self._update_round(latest_round_id, latest_round);
        } else {
            // only increment round when status flips
            let round_id = latest_round_id + 1_u128;
            self._record_round(round_id, status, timestamp);
        }
    }

    #[external(v0)]
    impl SequencerUptimeFeedImpl of super::ISequencerUptimeFeed<ContractState> {
        fn set_l1_sender(ref self: ContractState, address: EthAddress) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);

            assert(!address.is_zero(), '0 address not allowed');

            let old_address = self._l1_sender.read();

            if old_address != address.into() {
                self._l1_sender.write(address.into());
                self
                    .emit(
                        Event::L1SenderTransferred(
                            L1SenderTransferred {
                                from_address: old_address.try_into().unwrap(), to_address: address
                            }
                        )
                    );
            }
        }

        fn l1_sender(self: @ContractState) -> EthAddress {
            self._l1_sender.read().try_into().unwrap()
        }
    }

    ///
    /// Upgradeable
    ///

    #[external(v0)]
    fn upgrade(ref self: ContractState, new_impl: ClassHash) {
        let ownable = Ownable::unsafe_new_contract_state();
        Ownable::assert_only_owner(@ownable);
        Upgradeable::upgrade(new_impl)
    }

    ///
    /// Ownership
    ///

    #[external(v0)]
    impl OwnableImpl of IOwnable<ContractState> {
        fn owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::owner(@state)
        }

        fn proposed_owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::proposed_owner(@state)
        }

        fn transfer_ownership(ref self: ContractState, new_owner: ContractAddress) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::transfer_ownership(ref state, new_owner)
        }

        fn accept_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::accept_ownership(ref state)
        }

        fn renounce_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::renounce_ownership(ref state)
        }
    }


    ///
    /// Access Control
    ///

    #[external(v0)]
    impl AccessControllerImpl of IAccessController<ContractState> {
        fn has_access(self: @ContractState, user: ContractAddress, data: Array<felt252>) -> bool {
            let state = AccessControl::unsafe_new_contract_state();
            AccessControl::has_access(@state, user, data)
        }

        fn add_access(ref self: ContractState, user: ContractAddress) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::add_access(ref state, user)
        }

        fn remove_access(ref self: ContractState, user: ContractAddress) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::remove_access(ref state, user)
        }

        fn enable_access_check(ref self: ContractState) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::enable_access_check(ref state)
        }

        fn disable_access_check(ref self: ContractState) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::disable_access_check(ref state)
        }
    }


    ///
    /// Internals
    ///

    #[generate_trait]
    impl Internals of InternalTrait {
        fn _require_read_access(self: @ContractState) {
            let sender = starknet::info::get_caller_address();
            let access_control = AccessControl::unsafe_new_contract_state();
            AccessControl::check_read_access(@access_control, sender);
        }

        fn _initializer(
            ref self: ContractState, initial_status: u128, owner_address: ContractAddress
        ) {
            let mut ownable = Ownable::unsafe_new_contract_state();
            Ownable::constructor(ref ownable, owner_address);
            let mut access_control = AccessControl::unsafe_new_contract_state();
            AccessControl::constructor(ref access_control);
            let round_id = 1_u128;
            let timestamp = starknet::info::get_block_timestamp();
            self._record_round(round_id, initial_status, timestamp);
        }

        fn _record_round(ref self: ContractState, round_id: u128, status: u128, timestamp: u64) {
            self._latest_round_id.write(round_id);
            let block_info = starknet::info::get_block_info().unbox();
            let block_number = block_info.block_number;
            let block_timestamp = block_info.block_timestamp;

            let round = Transmission {
                answer: status,
                block_num: block_number,
                observation_timestamp: timestamp,
                transmission_timestamp: block_timestamp,
            };
            self._round_transmissions.write(round_id, round);

            let sender = starknet::info::get_caller_address();

            self
                .emit(
                    Event::NewRound(
                        NewRound { round_id: round_id, started_by: sender, started_at: timestamp }
                    )
                );
            self
                .emit(
                    Event::AnswerUpdated(
                        AnswerUpdated { current: status, round_id: round_id, timestamp: timestamp }
                    )
                );
        }

        fn _update_round(ref self: ContractState, round_id: u128, mut round: Transmission) {
            round.transmission_timestamp = starknet::info::get_block_timestamp();
            self._round_transmissions.write(round_id, round);

            self
                .emit(
                    Event::RoundUpdated(
                        RoundUpdated {
                            status: round.answer, updated_at: round.transmission_timestamp
                        }
                    )
                );
        }
    }
}
