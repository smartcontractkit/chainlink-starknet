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

use array::ArrayTrait;
use array::SpanTrait;
use hash::LegacyHash;

fn hash_span<
    T,
    impl THash: LegacyHash::<T>,
    impl TCopy: Copy::<T>
>(state: felt252, mut value: Span<T>) -> felt252 {
    let item = value.pop_front();
    match item {
        Option::Some(x) => {
            let s = LegacyHash::hash(state, *x);
            hash_span(s, value)
        },
        Option::None(_) => state,
    }
}
// TODO: wrap hash_span
impl SpanLegacyHash<
    T,
    impl THash: LegacyHash::<T>,
    impl TCopy: Copy::<T>
> of LegacyHash::<Span<T>> {
    fn hash(state: felt252, mut value: Span<T>) -> felt252 {
        hash_span(state, value)
    }
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
    use array::SpanTrait;
    use box::BoxTrait;
    use hash::LegacyHash;
    use super::SpanLegacyHash;

    const GIGA: u128 = 1000000000_u128;

    const MAX_ORACLES: u32 = 31_u32;

    #[event]
    fn NewTransmission(
        round_id: u128,
        answer: u128,
        transmitter: ContractAddress,
        observation_timestamp: u64,
        observers: felt252,
        observations: Array<u128>,
        juels_per_fee_coin: u128,
        gas_price: u128,
        config_digest: felt252,
        epoch_and_round: u64,
        reimbursement: u128
    ) {}

    // TODO: should we pack into (index, payment) = split_felt()? index is u8, payment is u128
    #[derive(Copy, Drop, Serde)]
    struct Oracle {
        index: usize,

        // entire supply of LINK always fits into u96, so u128 is safe to use
        payment_juels: u128,
    }

    impl OracleStorageAccess of StorageAccess::<Oracle> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Oracle> {
            Result::Ok(
                Oracle {
                    index: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 0_u8)
                    )?.try_into().unwrap(),
                    payment_juels: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 1_u8)
                    )?.try_into().unwrap(),
                }
            )
        }

        fn write(address_domain: u32, base: StorageBaseAddress, value: Oracle) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8), value.index.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 1_u8), value.payment_juels.into()
            )
        }
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
        _latest_epoch_and_round: u64, // (u32, u32)
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
        _transmitters: LegacyMap<ContractAddress, Oracle>, // <pkey, Oracle>
        _signers: LegacyMap<felt252, usize>, // <pkey, index>
        _signers_list: LegacyMap<usize, felt252>,
        _transmitters_list: LegacyMap<usize, ContractAddress>,
        _reward_from_aggregator_round_id: LegacyMap<usize, u128>, // <index, round_id>
        _transmissions: LegacyMap<u128, Transmission>,

        // link token
        _link_token: ContractAddress,

        // billing
        _billing_access_controller: ContractAddress,
        _billing: Billing,

        // payee management
        _payees: LegacyMap<ContractAddress, ContractAddress>, // <transmitter, payment_address>
        _proposed_payees: LegacyMap<ContractAddress, ContractAddress> // <transmitter, payment_address>
    }

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
            // TODO: require_access()
            _description::read()
        }

        fn decimals() -> u8 {
            // TODO: require_access()
            _decimals::read()
        }

        fn type_and_version() -> felt252 {
            0 // TODO
        }

    }

    // ---

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

    // --- Validation ---

    // NOTE: Currently unimplemented:

    // --- Configuration

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

    #[derive(Copy, Drop, Serde)]
    struct OracleConfig {
        signer: felt252,
        transmitter: ContractAddress,    
    }

    impl OracleConfigLegacyHash of LegacyHash::<OracleConfig> {
        fn hash(mut state: felt252, value: OracleConfig) -> felt252 {
            state = LegacyHash::hash(state, value.signer);
            state = LegacyHash::hash(state, value.transmitter);
            state
        }
    }

    #[derive(Copy, Drop, Serde)]
    struct OnchainConfig {
        version: u8,
        min_answer: u128,
        max_answer: u128,
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

        let min_answer = _min_answer::read();
        let max_answer = _max_answer::read();

        let computed_onchain_config = OnchainConfig {
            version: 1_u8,
            min_answer,
            max_answer,
        };

        // TODO: validate answer range

        pay_oracles();

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
            @computed_onchain_config,
            offchain_config_version,
            @offchain_config,
        );

        _latest_config_digest::write(digest);

        // reset epoch & round
        _latest_epoch_and_round::write(0_u64);

        ConfigSet(
            prev_block_num,
            digest,            
            config_count,
            oracles,
            f,
            onchain_config,
            offchain_config_version,
            offchain_config
        );

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
        _transmitters::write(transmitter, Oracle { index: 0_usize, payment_juels: 0_u128 });

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
        assert(existing_transmitter.index == 0_usize, 'repeated transmitter');

        _signers::write(*oracle.signer, index);
        _signers_list::write(index, *oracle.signer);

        _transmitters::write(*oracle.transmitter, Oracle { index, payment_juels: 0_u128 });
        _transmitters_list::write(index, *oracle.transmitter);

        _reward_from_aggregator_round_id::write(index, latest_round_id);

        add_oracles(oracles, index, len - 1_usize, latest_round_id)
    }

    // const DIGEST_MASK: felt252 = 2 ** (252 - 12) - 1;
    // const PREFIX: felt252 = 4 * 2 ** (252 - 12);
    const DIGEST_MASK: felt252 = 0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF;
    const PREFIX: felt252 =  0x4000000000000000000000000000000000000000000000000000000000000;

    fn config_digest_from_data(
        chain_id: felt252,
        contract_address: ContractAddress,
        config_count: u64,
        oracles: @Array<OracleConfig>,
        f: u8,
        onchain_config: @OnchainConfig,
        offchain_config_version: u64,
        offchain_config: @Array<felt252>,
    ) -> felt252 {
        let mut state = 0;
        state = LegacyHash::hash(state, chain_id);
        state = LegacyHash::hash(state, contract_address);
        state = LegacyHash::hash(state, config_count);
        state = LegacyHash::hash(state, oracles.len());
        // TODO: for oracle in oracles, hash each
        state = LegacyHash::hash(state, oracles.span());
        state = LegacyHash::hash(state, f);
        state = LegacyHash::hash(state, 3); // onchain_config.len() = 3
        state = LegacyHash::hash(state, *onchain_config.version);
        state = LegacyHash::hash(state, *onchain_config.min_answer);
        state = LegacyHash::hash(state, *onchain_config.max_answer);
        state = LegacyHash::hash(state, offchain_config_version);
        state = LegacyHash::hash(state, offchain_config.len());
        state = LegacyHash::hash(state, offchain_config.span());

        state
        // TODO: clamp first two bytes with the config digest prefix
        // NOTE: Cairo 1.0 missing bitwise operations on felt252
        // let masked = state & DIGEST_MASK;
        // masked + PREFIX
    }

    #[view]
    fn latest_config_details() -> (u64, u64, felt252) {
        let config_count = _config_count::read();
        let block_number = _latest_config_block_number::read();
        let config_digest = _latest_config_digest::read();

        (config_count, block_number, config_digest)
    }

    // #[view]
    // fn transmitters() {
    //
    // }

    // --- Transmission ---

    #[derive(Copy, Drop, Serde)]
    struct Signature {
        r: felt252,
        s: felt252,    
        public_key: felt252,
    }

    #[derive(Copy, Drop, Serde)]
    struct ReportContext {
        config_digest: felt252,
        epoch_and_round: u64,
        extra_hash: felt252,
    }

    #[external]
    fn transmit(
        report_context: ReportContext,
        observation_timestamp: u64,
        observers: felt252, // TODO:
        observations: Array<u128>,
        juels_per_fee_coin: u128,
        gas_price: u128,// TODO:
        signatures: Array<Signature>,
    ) {
    
        let epoch_and_round = _latest_epoch_and_round::read();
        assert(epoch_and_round < epoch_and_round, 'stale report');

        // validate transmitter
        let caller = starknet::info::get_caller_address();
        let oracle = _transmitters::read(caller);
        assert(oracle.index != 0_usize, 'unknown sender'); // 0 index = uninitialized

        // Validate config digest matches latest_config_digest
        let config_digest = _latest_config_digest::read();
        assert(report_context.config_digest == config_digest, 'config digest mismatch');

        let f = _f::read();
        assert(signatures.len().into() == (f + 1_u8).into(), 'wrong number of signatures');

        let msg = hash_report(
            @report_context,
            observation_timestamp,
            observers,
            @observations,
            juels_per_fee_coin,
            gas_price
        );

        // TODO: assert error here
        let signatures_len = signatures.len();
        verify_signatures(msg, @signatures, signatures_len, integer::u256_from_felt252(0));

        // report():

        let observations_len = observations.len();
        assert(observations_len <= MAX_ORACLES, '');
        assert(f.into().try_into().unwrap() < observations_len, '');

        _latest_epoch_and_round::write(report_context.epoch_and_round);

        let median_idx = observations_len / 2_usize;
        let median = observations[median_idx];

        // TODO: Validate median in min-max range

        let prev_round_id = _latest_aggregator_round_id::read();
        let round_id = prev_round_id + 1_u128;
        _latest_aggregator_round_id::write(round_id);

        let block_info = starknet::info::get_block_info().unbox();

        _transmissions::write(
            round_id,
            Transmission {
                answer: *median,
                block_num: block_info.block_number,
                observation_timestamp,
                transmission_timestamp: block_info.block_timestamp,
            }
        );

        // NOTE: Usually validating via validator would happen here, currently disabled

        let billing = _billing::read();
    
        // TODO: calculate_reimbursement
        let reimbursement_juels = 1_u128;

        // end report()

        NewTransmission(
            round_id,
            *median,
             caller,
            observation_timestamp,
            observers,
             observations,
            juels_per_fee_coin,
            gas_price,
            report_context.config_digest,
            report_context.epoch_and_round,
            reimbursement_juels,
        );

        // TODO: pay transmitter
        let payment = 1_u128;
        // TODO: check overflow

        _transmitters::write(
            caller,
            Oracle { index: oracle.index, payment_juels: oracle.payment_juels + payment } // TODO: modify oracle via oracle.payment_juels += payment?
        );

    }

    fn hash_report(
        report_context: @ReportContext,
        observation_timestamp: u64,
        observers: felt252,
        observations: @Array<u128>,
        juels_per_fee_coin: u128,
        gas_price: u128
    ) -> felt252 {
        let mut state = 0;
        state = LegacyHash::hash(state, *report_context.config_digest);
        state = LegacyHash::hash(state, *report_context.epoch_and_round);
        state = LegacyHash::hash(state, *report_context.extra_hash);
        state = LegacyHash::hash(state, observation_timestamp);
        state = LegacyHash::hash(state, observers);
        state = LegacyHash::hash(state, observations.len());
        state = LegacyHash::hash(state, observations.span());
        state = LegacyHash::hash(state, juels_per_fee_coin);
        state = LegacyHash::hash(state, gas_price);
        state
    }

    // TODO: signed_count feels more inefficient as u256
    fn verify_signatures(msg: felt252, signatures: @Array<Signature>, len: usize, signed_count: u256) {
        if len == 0_usize {
            // Check all signatures are unique (we only saw each pubkey once)
            // NOTE: This relies on protocol-level design constraints (MAX_ORACLES = 31, f = 10) which
            // ensures 31 bytes is enough to store a count for each oracle. Whenever the MAX_ORACLES
            // is updated the mask below should also be updated.
            assert(MAX_ORACLES == 31_u32, '');
            // TODO: 
            return ();
        }

        // TODO:
        let signature = signatures[0_usize];

        let index = _signers::read(*signature.public_key);
        assert(index != 0_usize, 'invalid signer'); // 0 index == uninitialized

        let is_valid = ecdsa::check_ecdsa_signature(
            msg,
            *signature.public_key,
            *signature.r,
            *signature.s
        );

        assert(is_valid, '');

        // signed_count + 1 << (8 * index)
        // TODO:
        // let (shift) = pow(2, 8 * index);
        // let signed_count = signed_count + shift;

        verify_signatures(msg, signatures, len - 1_usize, signed_count)
    }

    #[view]
    fn latest_transmission_details() {}

    // --- RequestNewRound

    // --- Queries

    #[view]
    fn description() -> felt252 {
        Aggregator::description()
    }

    #[view]
    fn decimals() -> u8 {
        Aggregator::decimals()
    }

    #[view]
    fn latest_round_data() -> Round {
        Aggregator::latest_round_data()
    }

    #[view]
    fn round_data(round_id: u128) -> Round {
        Aggregator::round_data(round_id)
    }

    // --- Set LINK Token

    #[event]
    fn LinkTokenSet(
        old_link_token: ContractAddress,
        new_link_token: ContractAddress,
    ) {}

    #[external]
    fn set_link_token(link_token: ContractAddress, recipient: ContractAddress) {
        // TODO: Ownable.assert_only_owner()

        let old_token = _link_token::read();

        if link_token == old_token {
            return ();
        }

        let contract_address = starknet::info::get_contract_address();

        // TODO
        // call balanceOf as a sanity check to confirm we're talking to a token
        // IERC20.balanceOf(contract_address=link_token, account=contract_address);

        pay_oracles();

        // TODO: 
        // transfer remaining balance to recipient
        // let (amount: Uint256) = IERC20.balanceOf(contract_address=link_token, account=contract_address);
        // IERC20.transfer(contract_address=old_token, recipient=recipient, amount=amount);

        _link_token::write(link_token);

        LinkTokenSet(old_token, link_token);
    }

    // --- Billing Access Controller

    #[event]
    fn BillingAccessControllerSet(
        old_controller: ContractAddress,
        new_controller: ContractAddress,
    ) {}

    #[external]
    fn set_billing_access_controller(access_controller: ContractAddress) {
        // TODO: Ownable.assert_only_owner()

        let old_controller = _billing_access_controller::read();
        if access_controller == old_controller {
            return ();
        }

        _billing_access_controller::write(access_controller);
        BillingAccessControllerSet(old_controller, access_controller);
    }

    // --- Billing Config

    #[derive(Copy, Drop, Serde)]
    struct Billing {
        // TODO: use a single felt via (observation_payment, transmission_payment) = split_felt()?
        observation_payment_gjuels: u32,
        transmission_payment_gjuels: u32,
        gas_base: u32,
        gas_per_signature: u32,
    }

    impl BillingStorageAccess of StorageAccess::<Billing> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Billing> {
            Result::Ok(
                Billing {
                    observation_payment_gjuels: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 0_u8)
                    )?.try_into().unwrap(),
                    transmission_payment_gjuels: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 1_u8)
                    )?.try_into().unwrap(),
                    gas_base: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 2_u8)
                    )?.try_into().unwrap(),
                    gas_per_signature: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 3_u8)
                    )?.try_into().unwrap(),
                }
            )
        }

        fn write(address_domain: u32, base: StorageBaseAddress, value: Billing) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8), value.observation_payment_gjuels.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 1_u8), value.transmission_payment_gjuels.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 2_u8), value.gas_base.into()
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 3_u8), value.gas_per_signature.into()
            )
        }
    }

    #[event]
    fn BillingSet(config: Billing) {}

    #[external]
    fn set_billing(config: Billing) {
        // TODO: has_billing_access();

        pay_oracles();

        _billing::write(config);

        BillingSet(config);
    }

    // --- Payments and Withdrawals

    #[event]
    fn OraclePaid(
        transmitter: ContractAddress,
        payee: ContractAddress,
        amount: u256,
        link_token: ContractAddress,
    ) {}

    #[external]
    fn withdraw_payment(transmitter: ContractAddress) {

    }

    fn _owed_payment(oracle: @Oracle) -> u128 {
        if *oracle.index == 0_usize {
            return 0_u128;
        }

        let billing = _billing::read();

        let latest_round_id = _latest_aggregator_round_id::read();
        let from_round_id = _reward_from_aggregator_round_id::read(*oracle.index);
        let rounds = latest_round_id - from_round_id;

        (rounds * billing.observation_payment_gjuels * GIGA) + *oracle.payment_juels
    }

    fn pay_oracles() {}

    // --- Transmitter Payment

    // --- Payee Management

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

}
