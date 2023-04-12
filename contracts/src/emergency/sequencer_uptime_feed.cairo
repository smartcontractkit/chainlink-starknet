impl Felt252TryIntoBool of TryInto::<felt252, bool> {
    fn try_into(self: felt252) -> Option<bool> {
        if self == 0 {
            return Option::Some(false);
        }
        if self == 1 {
            return Option::Some(true);
        }
        Option::None(())
    }
}

impl BoolIntoFelt252 of Into::<bool, felt252> {
    fn into(self: bool) -> felt252 {
        if (self) {
            return 1;
        }
        return 0;
    }
}

#[contract]
mod SequencerUptimeFeed {
    use super::Felt252TryIntoBool;
    use super::BoolIntoFelt252;
    use starknet::ContractAddress;
    use starknet::StorageAccess;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;

    use box::BoxTrait;
    use traits::Into;
    use traits::TryInto;
    use option::OptionTrait;

    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::access_controller::AccessController;
    use chainlink::libraries::simple_read_access_controller::SimpleReadAccessController;
    use chainlink::libraries::simple_write_access_controller::SimpleWriteAccessController;
    use chainlink::ocr2::aggregator::Round;
    use chainlink::ocr2::aggregator::IAggregator;

    #[derive(Copy, Drop, Serde, PartialEq)]
    struct RoundTransmission {
        answer: bool,
        block_num: u64,
        started_at: u64,
        updated_at: u64        
    }

    impl RoundTransmissionAccess of StorageAccess::<RoundTransmission> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<RoundTransmission> {
            Result::Ok(
                RoundTransmission {
                    answer: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 0_u8)
                    )?.try_into().unwrap(),
                    block_num: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 1_u8)
                    )?.try_into().unwrap(),
                    started_at: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 2_u8)
                    )?.try_into().unwrap(),
                    updated_at: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 3_u8)
                    )?.try_into().unwrap(),
                }
            )
        }

        fn write(
            address_domain: u32, base: StorageBaseAddress, value: RoundTransmission
        ) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 0_u8),
                value.answer.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 1_u8),
                value.block_num.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 2_u8),
                value.started_at.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 3_u8),
                value.updated_at.into()
            )
        }
    }


    struct Storage {
        // l1 sender is an ethereum address
        _l1_sender: u256,
        // maps round id to round transmission
        _round_transmissions: LegacyMap<u128, RoundTransmission>,
        _latest_round_id: u128,
    }

    #[event]
    fn RoundUpdated(status: bool, updated_at: u64) {}

    #[event]
    fn NewRound(round_id: u128, started_by: ContractAddress, started_at: u64) {}

    #[event]
    fn AnswerUpdated(current: bool, round_id: u128, timestamp: u64) {}

    #[event]
    fn UpdateIgnored(latest_status: bool, latest_timestamp: u64, incoming_status: bool, incoming_timestamp: u64) {}

    #[event]
    fn L1SenderTransferred(from_address: u256, to_address: u256) {}

    impl Aggregator of IAggregator {
        fn latest_round_data() -> Round {
            require_access();
            let latest_round_id = _latest_round_id::read();
            let round_transmission = _round_transmissions::read(latest_round_id);
            // bool -> felt252 -> u128
            let answer: u128 = round_transmission.answer.into().try_into().unwrap();
            Round {
                round_id: latest_round_id.into(),
                answer: answer,
                block_num: round_transmission.block_num,
                started_at: round_transmission.started_at,
                updated_at: round_transmission.updated_at,   
            }
        }

        fn round_data(round_id: u128) -> Round {
            require_access();
            assert(round_id < _latest_round_id::read(), 'invalid round id');
            let round_transmission = _round_transmissions::read(round_id);
            // bool -> felt252 -> u128
            let answer: u128 = round_transmission.answer.into().try_into().unwrap();
            Round {
                round_id: round_id.into(),
                answer: answer,
                block_num: round_transmission.block_num,
                started_at: round_transmission.started_at,
                updated_at: round_transmission.updated_at,   
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
    fn constructor(initial_status: bool, owner_address: ContractAddress) {
        initializer(initial_status, owner_address);
    }

    #[l1_handler]
    fn update_status(from_address: felt252, status: bool, timestamp: u64) {
        assert(_l1_sender::read() == from_address.into(), 'EXPECTED_FROM_BRIDGE_ONLY');

        let latest_round_id = _latest_round_id::read();
        let latest_round = _round_transmissions::read(latest_round_id);

        if (timestamp <= latest_round.started_at) {
            UpdateIgnored(latest_round.answer, latest_round.started_at, status, timestamp);
            return ();
        }

        if (latest_round.answer == status) {
            _update_round(latest_round_id, latest_round);
        } else {
            // only increment round when status flips
            let round_id = latest_round_id + 1_u128;
            _record_round(round_id, status, timestamp);
        }
    }

    #[external]
    fn set_l1_sender(address: u256) {
        Ownable::assert_only_owner();
        // because u256 literals are not supported
        let ETH_ADDRESS_BOUND = u256 { high: 0x100000000_u128, low: 0_u128 }; // 2 ** 160
        let ZERO = u256 { high: 0_u128, low: 0_u128 };

        assert(address < ETH_ADDRESS_BOUND, 'invalid eth address');
        assert(address != ZERO, '0 address not allowed');

        let old_address = _l1_sender::read();

        if (old_address != address) {
            _l1_sender::write(address);
            L1SenderTransferred(old_address, address);
        }

    }

    #[view]
    fn l1_sender() -> u256 {
        _l1_sender::read()
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
    fn transfer_owernship(new_owner: ContractAddress) {
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
    /// SimpleReadAccessController
    ///

    #[external]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        SimpleReadAccessController::has_access(user, data)
    }

     #[external]
    fn check_access(user: ContractAddress) {
        SimpleReadAccessController::check_access(user)
    }

    ///
    /// SimpleWriteAccessController
    ///

    #[external]
    fn add_access(user: ContractAddress) {
        SimpleWriteAccessController::add_access(user)
    }

    #[external]
    fn remove_access(user: ContractAddress) {
        SimpleWriteAccessController::remove_access(user)
    }

     #[external]
    fn enable_access_check() {
        SimpleWriteAccessController::enable_access_check()
    }

    #[external]
    fn disable_access_check() {
        SimpleWriteAccessController::disable_access_check()
    }


    ///
    /// Internals
    ///

    fn require_access() {
        let sender = starknet::info::get_caller_address();
         SimpleReadAccessController::check_access(sender);
    }

    fn initializer(initial_status: bool, owner_address: ContractAddress) {
        SimpleReadAccessController::initializer(owner_address);
        let round_id = 1_u128;
        let timestamp = starknet::info::get_block_timestamp();
        _record_round(round_id, initial_status, timestamp);
    }

    fn _record_round(round_id: u128, status: bool, timestamp: u64) {
        _latest_round_id::write(round_id);
        let block_info = starknet::info::get_block_info().unbox();
        let block_number = block_info.block_number;
        let block_timestamp = block_info.block_timestamp;

        let round = RoundTransmission {
            answer: status,
            block_num: block_number,
            started_at: timestamp,
            updated_at: block_timestamp,
        };
        _round_transmissions::write(round_id, round);

        let sender = starknet::info::get_caller_address();

        NewRound(round_id, sender, timestamp);
        AnswerUpdated(status, round_id, timestamp);
    }

    fn _update_round(round_id: u128, mut round: RoundTransmission) {
        round.updated_at = starknet::info::get_block_timestamp();
        _round_transmissions::write(round_id, round);

        RoundUpdated(round.answer, round.updated_at);
    }
}
