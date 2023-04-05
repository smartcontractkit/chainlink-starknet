
// TODO: round_id should probably be u128 then felt252 when prefixed

#[derive(Copy, Drop, Serde, PartialEq)]
struct Round {
    round_id: u128,
    answer: u128,
    block_num: u64,
    started_at: u64,
    updated_at: u64,
}

trait IAggregator {
    fn latest_round_data() -> Round;
    fn round_data(round_id: u128) -> Round;
    fn description() -> felt252;
    fn decimals() -> u8;
    fn type_and_version() -> felt252;
}

#[contract]
mod Aggregator {
    use super::IAggregator;
    use super::Round;
    use starknet::get_caller_address;
    use starknet::contract_address_const;
    use starknet::ContractAddressZeroable;
    use zeroable::Zeroable;

    use starknet::ContractAddress;

    use starknet::StorageAccess;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;
    use traits::Into;
    use traits::TryInto;
    use option::OptionTrait;

    #[derive(Copy, Drop, Serde)]
    struct Oracle {
        test: u128,
    }

    #[derive(Copy, Drop, Serde)]
    struct Transmission {
        answer: u128,
        block_num: u64,
        observation_timestamp: u64,
        transmission_timestamp: u64,
    }

    impl TransmissionStorageAccess of StorageAccess::<Transmission> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Transmission> {
            Result::Ok(
                Transmission {
                    answer: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 0_u8)
                    )?.try_into().unwrap(),
                    block_num: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 1_u8)
                    )?.try_into().unwrap(),
                    observation_timestamp: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 2_u8)
                    )?.try_into().unwrap(),
                    transmission_timestamp: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 3_u8)
                    )?.try_into().unwrap(),
                }
            )
        }

        fn write(address_domain: u32, base: StorageBaseAddress, value: Transmission) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8), value.answer.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 1_u8), value.block_num.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 2_u8), value.observation_timestamp.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 3_u8), value.transmission_timestamp.into()
            )
        }
    }

    struct Storage {
        /// Maximum number of faulty oracles
        _f: u8,
        _epoch_and_round: felt252, // TODO
        _latest_aggregator_round_id: u128, // TODO:
        _answer_range: bool, // TODO
        _decimals: u8,
        _description: felt252,
        _latest_config_block_number: felt252,
        _config_count: u128,
        _latest_config_digest: felt252,
        // _oracles: Array<Oracle>, // NOTE: array can't be used in storage
        // _transmitters, _signers, _signers_list, _transmitters_list
        _reward_from_aggregator_round_id_: LegacyMap<u8, u128>, // <index, round_id>
        _transmissions: LegacyMap<u128, Transmission>,

        // link token
        _link_token: ContractAddress,

        // billing
        _billing_access_controller: ContractAddress,
        _billing: bool, //TODO

        // payee management
        _payees: LegacyMap<ContractAddress, ContractAddress>, // <transmitter, payment_address>
        _proposed_payees: LegacyMap<ContractAddress, ContractAddress> // <transmitter, payment_address>
    }

    #[event]
    fn ConfigSet() {}

    #[event]
    fn LinkTokenSet() {}

    #[event]
    fn BillingAccessControllerSet() {}

    #[event]
    fn BillingSet() {}

    #[event]
    fn OraclePaid() {}

    #[event]
    fn PayeeshipTransferRequested() {}

    #[event]
    fn PayeeshipTransferred() {}

    impl Aggregator of IAggregator {
        fn latest_round_data() -> Round {
            // TODO: require_access()
            let latest_round_id = _latest_aggregator_round_id::read();
            let transmission = _transmissions::read(latest_round_id);
            Round {
                round_id: latest_round_id,
                answer: transmission.answer,
                block_num: transmission.block_num,
                started_at: transmission.observation_timestamp,
                updated_at: transmission.transmission_timestamp,
            }
        }

        fn round_data(round_id: u128) -> Round {
            // TODO: require_access()
            let transmission = _transmissions::read(round_id);
            Round {
                round_id,
                answer: transmission.answer,
                block_num: transmission.block_num,
                started_at: transmission.observation_timestamp,
                updated_at: transmission.transmission_timestamp,
            }
        }

        fn description() -> felt252 {
            _description::read()
        }

        fn decimals() -> u8 {
            _decimals::read()
        }

        fn type_and_version() -> felt252 {
            0 // TODO
        }

    }

    #[constructor]
    fn constructor(
        owner: ContractAddress,
        link: ContractAddress,
        min_answer: u128,
        max_answer: u128,
        billing_access_controller: ContractAddress,
        decimals: u8,
        description: felt252
    ) {
        // Ownable.initialize
        // SimpleReadAccessController.initialize
        _link_token::write(link);
        _billing_access_controller::write(billing_access_controller);

        // assert_lt (min, max)
        // TODO: write range

        _decimals::write(decimals);
        _description::write(description);
    }
}
