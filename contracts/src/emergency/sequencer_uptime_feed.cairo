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

    use openzeppelin::access::ownable::OwnableComponent;

    use chainlink::libraries::access_control::{AccessControlComponent, IAccessController};
    use chainlink::libraries::access_control::AccessControlComponent::InternalTrait as AccessControlInternalTrait;
    use chainlink::libraries::type_and_version::ITypeAndVersion;
    use chainlink::ocr2::aggregator::Round;
    use chainlink::ocr2::aggregator::IAggregator;
    use chainlink::ocr2::aggregator::{Transmission};
    use chainlink::libraries::upgradeable::Upgradeable;

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);
    component!(path: AccessControlComponent, storage: access_control, event: AccessControlEvent);

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl OwnableInternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl AccessControlImpl =
        AccessControlComponent::AccessControlImpl<ContractState>;
    impl AccessControlInternalImpl = AccessControlComponent::InternalImpl<ContractState>;

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
        #[substorage(v0)]
        access_control: AccessControlComponent::Storage,
        // l1 sender is an starknet validator ethereum address
        _l1_sender: EthAddress,
        // maps round id to round transmission
        _round_transmissions: LegacyMap<u128, Transmission>,
        _latest_round_id: u128,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event,
        #[flat]
        AccessControlEvent: AccessControlComponent::Event,
        RoundUpdated: RoundUpdated,
        NewRound: NewRound,
        AnswerUpdated: AnswerUpdated,
        UpdateIgnored: UpdateIgnored,
        L1SenderTransferred: L1SenderTransferred,
    }

    #[derive(Drop, starknet::Event)]
    struct RoundUpdated {
        status: u128,
        #[key]
        updated_at: u64
    }

    #[derive(Drop, starknet::Event)]
    struct NewRound {
        #[key]
        round_id: u128,
        #[key]
        started_by: EthAddress,
        started_at: u64
    }

    #[derive(Drop, starknet::Event)]
    struct AnswerUpdated {
        current: u128,
        #[key]
        round_id: u128,
        #[key]
        timestamp: u64
    }

    #[derive(Drop, starknet::Event)]
    struct UpdateIgnored {
        latest_status: u128,
        #[key]
        latest_timestamp: u64,
        incoming_status: u128,
        #[key]
        incoming_timestamp: u64
    }

    #[derive(Drop, starknet::Event)]
    struct L1SenderTransferred {
        #[key]
        from_address: EthAddress,
        #[key]
        to_address: EthAddress
    }

    #[abi(embed_v0)]
    impl TypeAndVersion of ITypeAndVersion<ContractState> {
        fn type_and_version(self: @ContractState) -> felt252 {
            'SequencerUptimeFeed 1.0.0'
        }
    }

    #[abi(embed_v0)]
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
            assert(round_id <= self._latest_round_id.read(), 'invalid round id');
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

        fn latest_answer(self: @ContractState) -> u128 {
            self._require_read_access();
            let latest_round_id = self._latest_round_id.read();
            let round_transmission = self._round_transmissions.read(latest_round_id);
            round_transmission.answer
        }
    }

    #[constructor]
    fn constructor(ref self: ContractState, initial_status: u128, owner_address: ContractAddress) {
        self._initializer(initial_status, owner_address);
    }

    #[l1_handler]
    fn update_status(ref self: ContractState, from_address: felt252, status: u128, timestamp: u64) {
        //  Cairo enforces from_address to be a felt252 on the method signature, but we can cast it right after
        let from_address: EthAddress = from_address.try_into().unwrap();
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
            self._record_round(from_address, round_id, status, timestamp);
        }
    }

    #[abi(embed_v0)]
    impl SequencerUptimeFeedImpl of super::ISequencerUptimeFeed<ContractState> {
        fn set_l1_sender(ref self: ContractState, address: EthAddress) {
            self.ownable.assert_only_owner();

            assert(!address.is_zero(), '0 address not allowed');

            let old_address = self._l1_sender.read();

            if old_address != address {
                self._l1_sender.write(address);
                self
                    .emit(
                        Event::L1SenderTransferred(
                            L1SenderTransferred { from_address: old_address, to_address: address }
                        )
                    );
            }
        }

        fn l1_sender(self: @ContractState) -> EthAddress {
            self._l1_sender.read()
        }
    }

    ///
    /// Upgradeable
    ///

    #[abi(embed_v0)]
    fn upgrade(ref self: ContractState, new_impl: ClassHash) {
        self.ownable.assert_only_owner();
        Upgradeable::upgrade(new_impl)
    }

    ///
    /// Internals
    ///

    #[generate_trait]
    impl Internals of InternalTrait {
        fn _require_read_access(self: @ContractState) {
            let sender = starknet::info::get_caller_address();
            self.access_control.check_read_access(sender);
        }

        fn _initializer(
            ref self: ContractState, initial_status: u128, owner_address: ContractAddress
        ) {
            self.ownable.initializer(owner_address);
            self.access_control.initializer();
            let round_id = 1_u128;
            let timestamp = starknet::info::get_block_timestamp();
            let from_address = EthAddress {
                address: 0
            }; // initial round is set by the constructor, not by an L1 sender
            self._record_round(from_address, round_id, initial_status, timestamp);
        }

        fn _record_round(
            ref self: ContractState,
            sender: EthAddress,
            round_id: u128,
            status: u128,
            timestamp: u64
        ) {
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
