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
use option::OptionTrait;
use hash::LegacyHash;

fn hash_span<T, impl THash: LegacyHash<T>, impl TCopy: Copy<T>>(
    state: felt252, mut value: Span<T>
) -> felt252 {
    let item = value.pop_front();
    match item {
        Option::Some(x) => {
            let s = LegacyHash::hash(state, *x);
            gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
            hash_span(s, value)
        },
        Option::None(_) => state,
    }
}

// TODO: consider switching to lookups
fn pow(n: u128, m: u128) -> u128 {
    if m == 0_u128 {
        return 1_u128;
    }
    gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
    let half = pow(n, m / 2_u128);
    let total = half * half;
    // TODO: check if (& 1) is cheaper
    if (m % 2_u128) == 1_u128 {
        total * n
    } else {
        total
    }
}

// TODO: wrap hash_span
impl SpanLegacyHash<T, impl THash: LegacyHash<T>, impl TCopy: Copy<T>> of LegacyHash<Span<T>> {
    fn hash(state: felt252, mut value: Span<T>) -> felt252 {
        hash_span(state, value)
    }
}

#[contract]
mod Aggregator {
    use super::IAggregator;
    use super::Round;
    use super::SpanLegacyHash;
    use super::pow;

    use array::ArrayTrait;
    use array::SpanTrait;
    use box::BoxTrait;
    use hash::LegacyHash;
    use integer::U8IntoFelt252;
    use integer::U32IntoFelt252;
    use integer::U128IntoFelt252;
    use integer::u128s_from_felt252;
    use integer::U128sFromFelt252Result;
    use zeroable::Zeroable;
    use traits::Into;
    use traits::TryInto;
    use option::OptionTrait;

    use starknet::ContractAddress;
    use starknet::get_caller_address;
    use starknet::contract_address_const;
    use starknet::StorageAccess;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;
    use starknet::class_hash::ClassHash;

    use chainlink::utils::split_felt;
    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::access_control::AccessControl;
    use chainlink::libraries::upgradeable::Upgradeable;

    // NOTE: remove duplication once we can directly use the trait
    #[abi]
    trait IERC20 {
        fn balance_of(account: ContractAddress) -> u256;
        fn transfer(recipient: ContractAddress, amount: u256) -> bool;
    // fn transfer_from(sender: ContractAddress, recipient: ContractAddress, amount: u256) -> bool;
    }

    // NOTE: remove duplication once we can directly use the trait
    #[abi]
    trait IAccessController {
        fn has_access(user: ContractAddress, data: Array<felt252>) -> bool;
        fn add_access(user: ContractAddress);
        fn remove_access(user: ContractAddress);
        fn enable_access_check();
        fn disable_access_check();
    }

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

    #[derive(Copy, Drop, Serde)]
    struct Oracle {
        index: usize,
        // entire supply of LINK always fits into u96, so u128 is safe to use
        payment_juels: u128,
    }

    impl OracleStorageAccess of StorageAccess<Oracle> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Oracle> {
            let value = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8)
            )?;
            let (index, payment_juels) = split_felt(value);
            Result::Ok(Oracle { index: index.into().try_into().unwrap(), payment_juels,  })
        }

        fn write(
            address_domain: u32, base: StorageBaseAddress, value: Oracle
        ) -> SyscallResult::<()> {
            let value = value.index.into() * SHIFT_128 + value.payment_juels.into();
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8), value
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

    impl TransmissionStorageAccess of StorageAccess<Transmission> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Transmission> {
            let block_num_and_answer = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8)
            )?;
            let (block_num, answer) = split_felt(block_num_and_answer);
            let timestamps = storage_read_syscall(
                address_domain, storage_address_from_base_and_offset(base, 1_u8)
            )?;
            let (observation_timestamp, transmission_timestamp) = split_felt(timestamps);

            Result::Ok(
                Transmission {
                    answer,
                    block_num: block_num.into().try_into().unwrap(),
                    observation_timestamp: observation_timestamp.into().try_into().unwrap(),
                    transmission_timestamp: transmission_timestamp.into().try_into().unwrap(),
                }
            )
        }

        fn write(
            address_domain: u32, base: StorageBaseAddress, value: Transmission
        ) -> SyscallResult::<()> {
            let block_num_and_answer = value.block_num.into() * SHIFT_128 + value.answer.into();
            let timestamps = value.observation_timestamp.into() * SHIFT_128
                + value.transmission_timestamp.into();
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 0_u8),
                block_num_and_answer
            )?;
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 1_u8), timestamps
            )
        }
    }

    struct Storage {
        /// Maximum number of faulty oracles
        _f: u8,
        _latest_epoch_and_round: u64, // (u32, u32)
        _latest_aggregator_round_id: u128,
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
        _proposed_payees: LegacyMap<ContractAddress,
        ContractAddress> // <transmitter, payment_address>
    }

    fn _require_read_access() {
        let caller = starknet::info::get_caller_address();
        AccessControl::check_read_access(caller);
    }

    impl Aggregator of IAggregator {
        fn latest_round_data() -> Round {
            _require_read_access();
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
            _require_read_access();
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
            _require_read_access();
            _description::read()
        }

        fn decimals() -> u8 {
            _require_read_access();
            _decimals::read()
        }

        fn type_and_version() -> felt252 {
            'ocr2/aggregator.cairo 1.0.0'
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
        Ownable::initializer(owner);
        AccessControl::initializer();
        _link_token::write(link);
        _billing_access_controller::write(billing_access_controller);

        assert(min_answer < max_answer, 'min >= max');
        _min_answer::write(min_answer);
        _max_answer::write(max_answer);

        _decimals::write(decimals);
        _description::write(description);
    }

    // --- Upgradeable ---

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl)
    }

    // --- Ownership ---

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

    // -- Access Control --

    #[view]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        AccessControl::has_access(user, data)
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

    impl OracleConfigLegacyHash of LegacyHash<OracleConfig> {
        fn hash(mut state: felt252, value: OracleConfig) -> felt252 {
            state = LegacyHash::hash(state, value.signer);
            state = LegacyHash::hash(state, value.transmitter);
            state
        }
    }

    #[external]
    fn set_config(
        oracles: Array<OracleConfig>,
        f: u8,
        onchain_config: Array<felt252>,
        offchain_config_version: u64,
        offchain_config: Array<felt252>,
    ) -> felt252 { // digest
        Ownable::assert_only_owner();
        assert(oracles.len() <= MAX_ORACLES, 'too many oracles');
        assert(U8IntoFelt252::into(3_u8 * f).try_into().unwrap() < oracles.len(), 'faulty-oracle f too high');
        assert(f > 0_u8, 'f must be positive');

        assert(onchain_config.len() == 0_u32, 'onchain_config must be empty');

        let min_answer = _min_answer::read();
        let max_answer = _max_answer::read();

        let mut computed_onchain_config = ArrayTrait::new();
        computed_onchain_config.append(1); // version
        computed_onchain_config.append(min_answer.into());
        computed_onchain_config.append(max_answer.into());

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
            computed_onchain_config,
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

        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        remove_oracles(n - 1_usize)
    }

    // TODO: measure gas, then rewrite (add_oracles and remove_oracles) using pop_front to see gas costs
    fn add_oracles(oracles: @Array<OracleConfig>, index: usize, len: usize, latest_round_id: u128) {
        if len == 0_usize {
            _oracles_len::write(index);
            return ();
        }

        let oracle = oracles.at(index);

        // NOTE: index should start with 1 here because storage is 0-initialized.
        // That way signers(pkey) => 0 indicates "not present"
        let index = index + 1_usize;

        // check for duplicates
        let existing_signer = _signers::read(*oracle.signer);
        assert(existing_signer == 0_usize, 'repeated signer');

        let existing_transmitter = _transmitters::read(*oracle.transmitter);
        assert(existing_transmitter.index == 0_usize, 'repeated transmitter');

        _signers::write(*oracle.signer, index);
        _signers_list::write(index, *oracle.signer);

        _transmitters::write(*oracle.transmitter, Oracle { index, payment_juels: 0_u128 });
        _transmitters_list::write(index, *oracle.transmitter);

        _reward_from_aggregator_round_id::write(index, latest_round_id);

        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        add_oracles(oracles, index, len - 1_usize, latest_round_id)
    }

    const SHIFT_128: felt252 = 0x100000000000000000000000000000000;
    // 4 * 2 ** (124 - 12)
    const HALF_PREFIX: u128 = 0x40000000000000000000000000000_u128;
    // 2 ** (124 - 12) - 1
    const HALF_DIGEST_MASK: u128 = 0xffffffffffffffffffffffffffff_u128;

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
        state = LegacyHash::hash(state, oracles.span()); // for oracle in oracles, hash each
        state = LegacyHash::hash(state, f);
        state = LegacyHash::hash(state, onchain_config.len());
        state = LegacyHash::hash(state, onchain_config.span());
        state = LegacyHash::hash(state, offchain_config_version);
        state = LegacyHash::hash(state, offchain_config.len());
        state = LegacyHash::hash(state, offchain_config.span());
        let len: usize = 3
            + 1
            + (oracles.len() * 2)
            + 1
            + 1
            + onchain_config.len()
            + 1
            + 1
            + offchain_config.len();
        state = LegacyHash::hash(state, len);

        // since there's no bitwise ops on felt252, we split into two u128s and recombine.
        // we only need to clamp and prefix the top bits.
        let (top, bottom) = split_felt(state);
        let masked_top = (top & HALF_DIGEST_MASK) | HALF_PREFIX;
        let masked = (masked_top.into() * SHIFT_128) + bottom.into();

        masked
    }

    #[view]
    fn latest_config_details() -> (u64, u64, felt252) {
        let config_count = _config_count::read();
        let block_number = _latest_config_block_number::read();
        let config_digest = _latest_config_digest::read();

        (config_count, block_number, config_digest)
    }

    #[view]
    fn transmitters() -> Array<ContractAddress> {
        let len = _oracles_len::read();
        let result = ArrayTrait::new();
        transmitters_(len, 0_usize, result)
    }

    fn transmitters_(
        len: usize, index: usize, mut result: Array<ContractAddress>
    ) -> Array<ContractAddress> {
        if len == 0_usize {
            return result;
        }
        let index = index + 1_usize;
        let transmitter = _transmitters_list::read(index);
        result.append(transmitter);
        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        transmitters_(len - 1_usize, index, result)
    }

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
        observers: felt252,
        observations: Array<u128>,
        juels_per_fee_coin: u128,
        gas_price: u128,
        mut signatures: Array<Signature>,
    ) {
        let signatures_len = signatures.len();

        let epoch_and_round = _latest_epoch_and_round::read();
        assert(epoch_and_round < report_context.epoch_and_round, 'stale report');

        // validate transmitter
        let caller = starknet::info::get_caller_address();
        let mut oracle = _transmitters::read(caller);
        assert(oracle.index != 0_usize, 'unknown sender'); // 0 index = uninitialized

        // Validate config digest matches latest_config_digest
        let config_digest = _latest_config_digest::read();
        assert(report_context.config_digest == config_digest, 'config digest mismatch');

        let f = _f::read();
        assert(signatures_len.into() == U8IntoFelt252::into(f + 1_u8), 'wrong number of signatures');

        let msg = hash_report(
            @report_context,
            observation_timestamp,
            observers,
            @observations,
            juels_per_fee_coin,
            gas_price
        );

        // Check all signatures are unique (we only saw each pubkey once)
        // NOTE: This relies on protocol-level design constraints (MAX_ORACLES = 31, f = 10) which
        // ensures we have enough bits to store a count for each oracle. Whenever the MAX_ORACLES
        // is updated, the signed_count parameter should be reconsidered.
        //
        // Although 31 bits is enough, we use a u128 here for simplicity because BitAnd and BitOr
        // operators are defined only for u128 and u256.
        assert(MAX_ORACLES == 31_u32, '');
        verify_signatures(msg, ref signatures, 0_u128);

        // report():

        let observations_len = observations.len();
        assert(observations_len <= MAX_ORACLES, '');
        assert(U8IntoFelt252::into(f).try_into().unwrap() < observations_len, '');

        _latest_epoch_and_round::write(report_context.epoch_and_round);

        let median_idx = observations_len / 2_usize;
        let median = *observations[median_idx];

        // Validate median in min-max range
        let min_answer = _min_answer::read();
        let max_answer = _max_answer::read();
        assert(min_answer <= median & median <= max_answer, 'median is out of min-max range');

        let prev_round_id = _latest_aggregator_round_id::read();
        let round_id = prev_round_id + 1_u128;
        _latest_aggregator_round_id::write(round_id);

        let block_info = starknet::info::get_block_info().unbox();

        _transmissions::write(
            round_id,
            Transmission {
                answer: median,
                block_num: block_info.block_number,
                observation_timestamp,
                transmission_timestamp: block_info.block_timestamp,
            }
        );

        // NOTE: Usually validating via validator would happen here, currently disabled

        let billing = _billing::read();
        let reimbursement_juels = calculate_reimbursement(
            juels_per_fee_coin, signatures_len, gas_price, billing
        );

        // end report()

        NewTransmission(
            round_id,
            median,
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

        // pay transmitter
        let payment = reimbursement_juels
            + (U32IntoFelt252::into(billing.transmission_payment_gjuels).try_into().unwrap() * GIGA);
        // TODO: check overflow

        oracle.payment_juels += payment;
        _transmitters::write(caller, oracle);
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
        let len: usize = 5 + 1 + observations.len() + 2;
        state = LegacyHash::hash(state, len);
        state
    }

    fn verify_signatures(msg: felt252, ref signatures: Array<Signature>, mut signed_count: u128) {
        let signature = match signatures.pop_front() {
            Option::Some(signature) => signature,
            Option::None(_) => {
                // No more signatures left!
                return ();
            }
        };

        let index = _signers::read(signature.public_key);
        assert(index != 0_usize, 'invalid signer'); // 0 index == uninitialized

        let indexed_bit = pow(2_u128, U32IntoFelt252::into(index).try_into().unwrap() - 1_u128);
        let prev_signed_count = signed_count;
        signed_count = signed_count | indexed_bit;
        assert(prev_signed_count != signed_count, 'duplicate signer');

        let is_valid = ecdsa::check_ecdsa_signature(
            msg, signature.public_key, signature.r, signature.s
        );

        assert(is_valid, '');

        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        verify_signatures(msg, ref signatures, signed_count)
    }

    #[view]
    fn latest_transmission_details() -> (felt252, u64, u128, u64) {
        let config_digest = _latest_config_digest::read();
        let latest_round_id = _latest_aggregator_round_id::read();
        let epoch_and_round = _latest_epoch_and_round::read();
        let transmission = _transmissions::read(latest_round_id);
        (config_digest, epoch_and_round, transmission.answer, transmission.transmission_timestamp)
    }

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

    #[view]
    fn type_and_version() -> felt252 {
        Aggregator::type_and_version()
    }

    // --- Set LINK Token

    #[event]
    fn LinkTokenSet(old_link_token: ContractAddress, new_link_token: ContractAddress, ) {}

    #[external]
    fn set_link_token(link_token: ContractAddress, recipient: ContractAddress) {
        Ownable::assert_only_owner();

        let old_token = _link_token::read();

        if link_token == old_token {
            return ();
        }

        let contract_address = starknet::info::get_contract_address();

        // call balanceOf as a sanity check to confirm we're talking to a token
        let token = IERC20Dispatcher { contract_address: link_token };
        token.balance_of(account: contract_address);

        pay_oracles();

        // transfer remaining balance of old token to recipient
        let old_token_dispatcher = IERC20Dispatcher { contract_address: old_token };
        let amount = old_token_dispatcher.balance_of(account: contract_address);
        old_token_dispatcher.transfer(recipient, amount);

        _link_token::write(link_token);

        LinkTokenSet(old_token, link_token);
    }

    // --- Billing Access Controller

    #[event]
    fn BillingAccessControllerSet(
        old_controller: ContractAddress, new_controller: ContractAddress, 
    ) {}

    #[external]
    fn set_billing_access_controller(access_controller: ContractAddress) {
        Ownable::assert_only_owner();

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
        observation_payment_gjuels: u32,
        transmission_payment_gjuels: u32,
        gas_base: u32,
        gas_per_signature: u32,
    }

    impl BillingStorageAccess of StorageAccess<Billing> {
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

        fn write(
            address_domain: u32, base: StorageBaseAddress, value: Billing
        ) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 0_u8),
                value.observation_payment_gjuels.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 1_u8),
                value.transmission_payment_gjuels.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 2_u8),
                value.gas_base.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 3_u8),
                value.gas_per_signature.into()
            )
        }
    }

    #[event]
    fn BillingSet(config: Billing) {}

    #[external]
    fn set_billing(config: Billing) {
        has_billing_access();

        pay_oracles();

        _billing::write(config);

        BillingSet(config);
    }

    #[view]
    fn billing() -> Billing {
        _billing::read()
    }

    fn has_billing_access() {
        let caller = starknet::info::get_caller_address();
        let owner = Ownable::owner();

        // owner always has access
        if caller == owner {
            return ();
        }

        let access_controller = _billing_access_controller::read();
        let access_controller = IAccessControllerDispatcher { contract_address: access_controller };
        assert(
            access_controller.has_access(caller, ArrayTrait::new()), 'caller does not have access'
        );
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
        let caller = starknet::info::get_caller_address();
        let payee = _payees::read(transmitter);
        assert(caller == payee, 'only payee can withdraw');

        let latest_round_id = _latest_aggregator_round_id::read();
        let link_token = _link_token::read();
        pay_oracle(transmitter, latest_round_id, link_token)
    }

    fn _owed_payment(oracle: @Oracle) -> u128 {
        if *oracle.index == 0_usize {
            return 0_u128;
        }

        let billing = _billing::read();

        let latest_round_id = _latest_aggregator_round_id::read();
        let from_round_id = _reward_from_aggregator_round_id::read(*oracle.index);
        let rounds = latest_round_id - from_round_id;

        (rounds * U32IntoFelt252::into(billing.observation_payment_gjuels).try_into().unwrap() * GIGA)
            + *oracle.payment_juels
    }

    #[view]
    fn owed_payment(transmitter: ContractAddress) -> u128 {
        let oracle = _transmitters::read(transmitter);
        _owed_payment(@oracle)
    }

    fn pay_oracle(
        transmitter: ContractAddress, latest_round_id: u128, link_token: ContractAddress
    ) {
        let oracle = _transmitters::read(transmitter);
        if oracle.index == 0_usize {
            return ();
        }

        let amount = _owed_payment(@oracle);
        // if zero, fastpath return to avoid empty transfers
        if amount == 0_u128 {
            return ();
        }

        let payee = _payees::read(transmitter);

        // NOTE: equivalent to converting u128 to u256
        let amount = u256 { high: 0_u128, low: amount };

        let token = IERC20Dispatcher { contract_address: link_token };
        token.transfer(recipient: payee, amount: amount);

        // Reset payment
        _reward_from_aggregator_round_id::write(oracle.index, latest_round_id);
        _transmitters::write(transmitter, Oracle { index: oracle.index, payment_juels: 0_u128 });

        OraclePaid(transmitter, payee, amount, link_token);
    }

    fn pay_oracles() {
        let len = _oracles_len::read();
        let latest_round_id = _latest_aggregator_round_id::read();
        let link_token = _link_token::read();
        pay_oracles_(len, latest_round_id, link_token)
    }

    fn pay_oracles_(index: usize, latest_round_id: u128, link_token: ContractAddress) {
        if index == 0_usize {
            return ();
        }

        let transmitter = _transmitters_list::read(index);
        pay_oracle(transmitter, latest_round_id, link_token);

        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        pay_oracles_(index - 1_usize, latest_round_id, link_token)
    }

    #[external]
    fn withdraw_funds(recipient: ContractAddress, amount: u256) {
        has_billing_access();

        let link_token = _link_token::read();
        let contract_address = starknet::info::get_contract_address();

        let due = total_link_due();
        // NOTE: equivalent to converting u128 to u256
        let due = u256 { high: 0_u128, low: due };

        let token = IERC20Dispatcher { contract_address: link_token };
        let balance = token.balance_of(account: contract_address);

        assert(due <= balance, 'amount due exceeds balance');
        let available = balance - due;

        // Transfer as much as there is available
        let amount = if available < amount {
            available
        } else {
            amount
        };
        token.transfer(recipient, amount);
    }

    fn total_link_due() -> u128 {
        let len = _oracles_len::read();
        let latest_round_id = _latest_aggregator_round_id::read();

        total_link_due_(len, latest_round_id, 0_u128, 0_u128)
    }
    fn total_link_due_(
        index: usize, latest_round_id: u128, total_rounds: u128, payments_juels: u128
    ) -> u128 {
        if index == 0_usize {
            let billing = _billing::read();
            return (total_rounds
                * U32IntoFelt252::into(billing.observation_payment_gjuels).try_into().unwrap()
                * GIGA)
                + payments_juels;
        }

        let transmitter = _transmitters_list::read(index);
        let oracle = _transmitters::read(transmitter);
        assert(oracle.index != 0_usize, ''); // 0 == undefined

        let from_round_id = _reward_from_aggregator_round_id::read(oracle.index);
        let rounds = latest_round_id - from_round_id;

        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        total_link_due_(
            index - 1_usize,
            latest_round_id,
            total_rounds + rounds,
            payments_juels + oracle.payment_juels
        )
    }

    #[view]
    fn link_available_for_payment() -> (bool, u128) { // (is negative, absolute difference)
        let link_token = _link_token::read();
        let contract_address = starknet::info::get_contract_address();

        let token = IERC20Dispatcher { contract_address: link_token };
        let balance = token.balance_of(account: contract_address);
        // entire link supply fits into u96 so this should not fail
        assert(balance.high == 0_u128, 'balance too high');
        let balance: u128 = balance.low;

        let due = total_link_due();
        if balance > due {
            (false, balance - due)
        } else {
            (true, due - balance)
        }
    }

    // --- Transmitter Payment

    const MARGIN: u128 = 115_u128;

    fn calculate_reimbursement(
        juels_per_fee_coin: u128, signature_count: usize, gas_price: u128, config: Billing
    ) -> u128 {
        // TODO: determine new values for these constants
        // Based on estimateFee (f=1 14977, f=2 14989, f=3 15002 f=4 15014 f=5 15027, count = f+1)
        // gas_base = 14951, gas_per_signature = 13
        let signature_count_u128: u128 = U32IntoFelt252::into(signature_count).try_into().unwrap();
        let gas_base_u128: u128 = U32IntoFelt252::into(config.gas_base).try_into().unwrap();
        let gas_per_signature_u128: u128 = U32IntoFelt252::into(config.gas_per_signature).try_into().unwrap();

        let exact_gas = gas_base_u128 + (signature_count_u128 * gas_per_signature_u128);
        let gas = exact_gas * MARGIN / 100_u128; // scale to 115% for some margin
        let amount = gas * gas_price;
        amount * juels_per_fee_coin
    }

    // --- Payee Management

    #[event]
    fn PayeeshipTransferRequested(
        transmitter: ContractAddress, current: ContractAddress, proposed: ContractAddress, 
    ) {}

    #[event]
    fn PayeeshipTransferred(
        transmitter: ContractAddress, previous: ContractAddress, current: ContractAddress, 
    ) {}

    #[derive(Copy, Drop, Serde)]
    struct PayeeConfig {
        transmitter: ContractAddress,
        payee: ContractAddress,
    }

    #[external]
    fn set_payees(payees: Array<PayeeConfig>) {
        Ownable::assert_only_owner();
        set_payee(payees)
    }

    fn set_payee(mut payees: Array<PayeeConfig>) {
        let payee = match payees.pop_front() {
            Option::Some(payee) => payee,
            Option::None(_) => {
                // No more payees left!
                return ();
            }
        };

        let current_payee = _payees::read(payee.transmitter);
        let is_unset = current_payee.is_zero();
        let is_same = current_payee == payee.payee;
        assert(is_unset | is_same, 'payee already set');

        _payees::write(payee.transmitter, payee.payee);

        PayeeshipTransferred(payee.transmitter, current_payee, payee.payee);

        gas::withdraw_gas_all(get_builtin_costs()).expect('Out of gas');
        set_payee(payees)
    }

    #[external]
    fn transfer_payeeship(transmitter: ContractAddress, proposed: ContractAddress) {
        assert(!proposed.is_zero(), 'cannot transfer to zero address');
        let caller = starknet::info::get_caller_address();
        let payee = _payees::read(transmitter);
        assert(caller == payee, 'only current payee can update');
        assert(caller != proposed, 'cannot transfer to self');

        _proposed_payees::write(transmitter, proposed);
        PayeeshipTransferRequested(transmitter, payee, proposed)
    }

    #[external]
    fn accept_payeeship(transmitter: ContractAddress) {
        let proposed = _proposed_payees::read(transmitter);
        let caller = starknet::info::get_caller_address();
        assert(caller == proposed, 'only proposed payee can accept');
        let previous = _payees::read(transmitter);

        _payees::write(transmitter, proposed);
        _proposed_payees::write(transmitter, Zeroable::zero());
        PayeeshipTransferred(transmitter, previous, caller)
    }
}
