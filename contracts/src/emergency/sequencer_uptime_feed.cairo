#[contract]
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

    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::access_control::AccessControl;
    use chainlink::ocr2::aggregator::Round;
    use chainlink::ocr2::aggregator::IAggregator;
    use chainlink::ocr2::aggregator::Aggregator::Transmission;
    use chainlink::ocr2::aggregator::Aggregator::TransmissionStorageAccess;
    use chainlink::libraries::upgradeable::Upgradeable;

    struct Storage {
        // l1 sender is an starknet validator ethereum address
        _l1_sender: felt252,
        // maps round id to round transmission
        _round_transmissions: LegacyMap<u128, Transmission>,
        _latest_round_id: u128,
    }

    #[event]
    fn RoundUpdated(status: u128, updated_at: u64) {}

    #[event]
    fn NewRound(round_id: u128, started_by: ContractAddress, started_at: u64) {}

    #[event]
    fn AnswerUpdated(current: u128, round_id: u128, timestamp: u64) {}

    #[event]
    fn UpdateIgnored(
        latest_status: u128, latest_timestamp: u64, incoming_status: u128, incoming_timestamp: u64
    ) {}

    #[event]
    fn L1SenderTransferred(from_address: EthAddress, to_address: EthAddress) {}

    impl Aggregator of IAggregator {
        fn latest_round_data() -> Round {
            _require_read_access();
            let latest_round_id = _latest_round_id::read();
            let round_transmission = _round_transmissions::read(latest_round_id);
            Round {
                round_id: latest_round_id.into(),
                answer: round_transmission.answer,
                block_num: round_transmission.block_num,
                started_at: round_transmission.observation_timestamp,
                updated_at: round_transmission.transmission_timestamp,
            }
        }

        fn round_data(round_id: u128) -> Round {
            _require_read_access();
            assert(round_id < _latest_round_id::read(), 'invalid round id');
            let round_transmission = _round_transmissions::read(round_id);
            Round {
                round_id: round_id.into(),
                answer: round_transmission.answer,
                block_num: round_transmission.block_num,
                started_at: round_transmission.observation_timestamp,
                updated_at: round_transmission.transmission_timestamp,
            }
        }

        fn description() -> felt252 {
            'L2 Sequencer Uptime Status Feed'
        }

        fn decimals() -> u8 {
            0_u8
        }

        fn type_and_version() -> felt252 {
            'SequencerUptimeFeed 1.0.0'
        }
    }

    #[constructor]
    fn constructor(initial_status: u128, owner_address: ContractAddress) {
        _initializer(initial_status, owner_address);
    }

    #[l1_handler]
    fn update_status(from_address: felt252, status: u128, timestamp: u64) {
        assert(_l1_sender::read() == from_address, 'EXPECTED_FROM_BRIDGE_ONLY');

        let latest_round_id = _latest_round_id::read();
        let latest_round = _round_transmissions::read(latest_round_id);

        if timestamp <= latest_round.observation_timestamp {
            UpdateIgnored(
                latest_round.answer, latest_round.transmission_timestamp, status, timestamp
            );
            return ();
        }

        if latest_round.answer == status {
            _update_round(latest_round_id, latest_round);
        } else {
            // only increment round when status flips
            let round_id = latest_round_id + 1_u128;
            _record_round(round_id, status, timestamp);
        }
    }

    #[external]
    fn set_l1_sender(address: EthAddress) {
        Ownable::assert_only_owner();

        assert(!address.is_zero(), '0 address not allowed');

        let old_address = _l1_sender::read();

        if old_address != address.into() {
            _l1_sender::write(address.into());
            L1SenderTransferred(
                old_address.try_into().unwrap(), 
                address
            );
        }
    }

    #[view]
    fn l1_sender() -> EthAddress {
        _l1_sender::read().try_into().unwrap()
    }

    ///
    /// Upgradeable
    ///

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl)
    }

    ///
    /// Aggregator
    ///

    #[view]
    fn latest_round_data() -> Round {
        Aggregator::latest_round_data()
    }

    #[view]
    fn round_data(round_id: u128) -> Round {
        Aggregator::round_data(round_id)
    }

    #[view]
    fn description() -> felt252 {
        Aggregator::description()
    }

    #[view]
    fn decimals() -> u8 {
        Aggregator::decimals()
    }

    #[view]
    fn type_and_version() -> felt252 {
        Aggregator::type_and_version()
    }

    ///
    /// Ownership
    ///

    #[view]
    fn owner() -> ContractAddress {
        Ownable::owner()
    }

    #[view]
    fn proposed_owner() -> ContractAddress {
        Ownable::proposed_owner()
    }

    #[external]
    fn transfer_ownership(new_owner: ContractAddress) {
        Ownable::transfer_ownership(new_owner)
    }

    #[external]
    fn accept_ownership() {
        Ownable::accept_ownership()
    }

    #[external]
    fn renounce_ownership() {
        Ownable::renounce_ownership()
    }

    ///
    /// Access Control
    ///

    #[view]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        AccessControl::has_access(user, data)
    }

    #[view]
    fn check_access(user: ContractAddress) {
        AccessControl::check_access(user)
    }

    #[external]
    fn add_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        AccessControl::add_access(user);
    }

    #[external]
    fn remove_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        AccessControl::remove_access(user);
    }

    #[external]
    fn enable_access_check() {
        Ownable::assert_only_owner();
        AccessControl::enable_access_check();
    }

    #[external]
    fn disable_access_check() {
        Ownable::assert_only_owner();
        AccessControl::disable_access_check();
    }


    ///
    /// Internals
    ///

    fn _require_read_access() {
        let sender = starknet::info::get_caller_address();
        AccessControl::check_read_access(sender);
    }

    fn _initializer(initial_status: u128, owner_address: ContractAddress) {
        Ownable::initializer(owner_address);
        AccessControl::initializer();
        let round_id = 1_u128;
        let timestamp = starknet::info::get_block_timestamp();
        _record_round(round_id, initial_status, timestamp);
    }

    fn _record_round(round_id: u128, status: u128, timestamp: u64) {
        _latest_round_id::write(round_id);
        let block_info = starknet::info::get_block_info().unbox();
        let block_number = block_info.block_number;
        let block_timestamp = block_info.block_timestamp;

        let round = Transmission {
            answer: status,
            block_num: block_number,
            observation_timestamp: timestamp,
            transmission_timestamp: block_timestamp,
        };
        _round_transmissions::write(round_id, round);

        let sender = starknet::info::get_caller_address();

        NewRound(round_id, sender, timestamp);
        AnswerUpdated(status, round_id, timestamp);
    }

    fn _update_round(round_id: u128, mut round: Transmission) {
        round.transmission_timestamp = starknet::info::get_block_timestamp();
        _round_transmissions::write(round_id, round);

        RoundUpdated(round.answer, round.transmission_timestamp);
    }
}
