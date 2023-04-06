#[derive(Copy, Drop, Serde, PartialEq)]
struct Round {
    // used as u128 internally, but necessary for phase-prefixed round ids as returned by proxy
    round_id: felt252,
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
    use integer::U128IntoFelt252;
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

    use array::ArrayTrait;
    use box::BoxTrait;
    use hash::LegacyHash;

    const MAX_ORACLES: u32 = 31_u32;

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
        _latest_epoch_and_round: felt252,
        _latest_aggregator_round_id: u128,
        // _answer_range: (u128, u128), // TODO        
        _min_answer: u128,
        _max_answer: u128,
        _decimals: u8,
        _description: felt252,
        _latest_config_block_number: u64,
        _config_count: u64,
        _latest_config_digest: felt252,

        // _oracles: Array<Oracle>, // NOTE: array can't be used in storage
        _oracles_len: usize,
        _transmitters: LegacyMap<ContractAddress, usize>, // <pkey, Oracle>
        _signers: LegacyMap<felt252, usize>, // <pkey, index>
        _signers_list: LegacyMap<usize, felt252>,
        _transmitters_list: LegacyMap<usize, ContractAddress>,
        _reward_from_aggregator_round_id: LegacyMap<usize, u128>, // <index, round_id>
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
    fn ConfigSet(
        previous_config_block_number: u64,
        latest_config_digest: felt252,
        config_count: u64,
        oracles: Array<OracleConfig>,
        f: u8,
        onchain_config: Array<felt252>,
        offchain_config_version: u64,
        offchain_config: Array<felt252>,
    ) {}

    #[event]
    fn LinkTokenSet(
        old_link_token: ContractAddress,
        new_link_token: ContractAddress,
    ) {}

    #[event]
    fn BillingAccessControllerSet(
        old_controller: ContractAddress,
        new_controller: ContractAddress,
    ) {}

    #[event]
    fn BillingSet(
        config: bool // TODO:
    ) {}

    #[event]
    fn OraclePaid(
        transmitter: ContractAddress,
        payee: ContractAddress,
        amount: u256,
        link_token: ContractAddress,
    ) {}

    #[event]
    fn PayeeshipTransferRequested(
        transmitter: ContractAddress,
        current: ContractAddress,
        proposed: ContractAddress,
    ) {}

    #[event]
    fn PayeeshipTransferred(
        transmitter: ContractAddress,
        previous: ContractAddress,
        current: ContractAddress,
    ) {}

    impl Aggregator of IAggregator {
        fn latest_round_data() -> Round {
            // TODO: require_access()
            let latest_round_id = _latest_aggregator_round_id::read();
            let transmission = _transmissions::read(latest_round_id);
            Round {
                round_id: latest_round_id.into(),
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
                round_id: round_id.into(),
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
        // _answer_range::write((min_answer, max_answer));
        _min_answer::write(min_answer);
        _max_answer::write(max_answer);

        _decimals::write(decimals);
        _description::write(description);
    }

    #[derive(Copy, Drop, Serde)]
    struct OracleConfig {
        signer: felt252,
        transmitter: ContractAddress,    
    }

    #[external]
    fn set_config(
        oracles: Array<OracleConfig>,
        f: u8,
        onchain_config: Array<felt252>,
        offchain_config_version: u64,
        offchain_config: Array<felt252>,
    ) -> felt252 { // digest
        // TODO: Ownable.assert_only_owner()
        assert(oracles.len() <= MAX_ORACLES, 'too many oracles');
        assert((3_u8 * f).into().try_into().unwrap() < oracles.len(), 'faulty-oracle f too high'); // NOTE: messy cast 
        assert(f > 0_u8, 'f must be positive');

        assert(onchain_config.len() == 0_u32, 'onchain_config must be empty');

        // TODO: validate answer range

        // TODO: pay_oracles()

        // remove old signers & transmitters
        let len = _oracles_len::read();
        remove_oracles(len);

        let latest_round_id = _latest_aggregator_round_id::read();

        let oracles_len = oracles.len(); // work around variable move issue
        add_oracles(@oracles, 0_usize, oracles_len, latest_round_id);

        _f::write(f);
        let block_num = starknet::info::get_block_info().unbox().block_number;
        let prev_block_num = _latest_config_block_number::read();
        _latest_config_block_number::write(block_num);
        // update config count
        let mut config_count = _config_count::read();
        config_count += 1_u64;
        _config_count::write(config_count);
        let contract_address = starknet::info::get_contract_address();
        let chain_id = starknet::info::get_tx_info().unbox().chain_id;

        let digest = config_digest_from_data(
            chain_id,
            contract_address,
            config_count,
            @oracles,
            f,
            @onchain_config,
            offchain_config_version,
            @offchain_config,
        );

        _latest_config_digest::write(digest);

        // reset epoch & round
        _latest_epoch_and_round::write(0);

        // TODO: ConfigSet()

        digest
    }

    fn remove_oracles(n: usize) {
        if n == 0_usize {
            _oracles_len::write(0_usize);
            return ();
        }

        let signer = _signers_list::read(n);
        _signers::write(signer, 0_usize);

        let transmitter = _transmitters_list::read(n);
        _transmitters::write(transmitter, 0_usize);

        remove_oracles(n - 1_usize)
    }

    // TODO: explore using a slice/Span + pop_front rather than index
    fn add_oracles(oracles: @Array<OracleConfig>, index: usize, len: usize, latest_round_id: u128) {
        // NOTE: index should start with 1 here because storage is 0-initialized.
        // That way signers(pkey) => 0 indicates "not present"
        let index = index + 1_usize;

        if len == 0_usize {
            _oracles_len::write(len);
            return ();
        }

        let oracle = oracles[index];
        // TODO: check for duplicates
        let existing_signer = _signers::read(*oracle.signer);
        assert(existing_signer == 0_usize, 'repeated signer');

        let existing_transmitter = _transmitters::read(*oracle.transmitter);
        assert(existing_transmitter == 0_usize, 'repeated transmitter');

        _signers::write(*oracle.signer, index);
        _signers_list::write(index, *oracle.signer);

        _transmitters::write(*oracle.transmitter, index);
        _transmitters_list::write(index, *oracle.transmitter);

        _reward_from_aggregator_round_id::write(index, latest_round_id);

        add_oracles(oracles, index, len - 1_usize, latest_round_id)
    }

    // const DIGEST_MASK = 2 ** (252 - 12) - 1;
    // const PREFIX = 4 * 2 ** (252 - 12);

    fn config_digest_from_data(
        chain_id: felt252,
        contract_address: ContractAddress,
        config_count: u64,
        oracles: @Array<OracleConfig>,
        f: u8,
        onchain_config: @Array<felt252>,
        offchain_config_version: u64,
        offchain_config: @Array<felt252>,
    ) -> felt252 {
        let mut state = 0;
        state = LegacyHash::hash(state, chain_id);
        state = LegacyHash::hash(state, contract_address);
        state = LegacyHash::hash(state, config_count);
        state = LegacyHash::hash(state, oracles.len());
        // TODO: for oracle in oracles, hash each
        state = LegacyHash::hash(state, f);
        state = LegacyHash::hash(state, onchain_config.len());
        // TODO: onchain_config
        state = LegacyHash::hash(state, offchain_config_version);
        state = LegacyHash::hash(state, offchain_config.len());
        // TODO: offchain_config

        // TODO: clamp first two bytes with the config digest prefix
        state
    }
}
