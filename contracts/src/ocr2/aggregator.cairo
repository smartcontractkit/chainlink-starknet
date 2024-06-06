use starknet::ContractAddress;

#[derive(Copy, Drop, Serde, PartialEq, starknet::Store)]
struct Round {
    // used as u128 internally, but necessary for phase-prefixed round ids as returned by proxy
    round_id: felt252,
    answer: u128,
    block_num: u64,
    started_at: u64,
    updated_at: u64,
}

#[derive(Copy, Drop, Serde, starknet::Store)]
struct Transmission {
    answer: u128,
    block_num: u64,
    observation_timestamp: u64,
    transmission_timestamp: u64,
}

// TODO: reintroduce custom storage to save on space
// impl TransmissionStorageAccess of StorageAccess<Transmission> {
//     fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Transmission> {
//         let block_num_and_answer = storage_read_syscall(
//             address_domain, storage_address_from_base_and_offset(base, 0_u8)
//         )?;
//         let (block_num, answer) = split_felt(block_num_and_answer);
//         let timestamps = storage_read_syscall(
//             address_domain, storage_address_from_base_and_offset(base, 1_u8)
//         )?;
//         let (observation_timestamp, transmission_timestamp) = split_felt(timestamps);

//         Result::Ok(
//             Transmission {
//                 answer,
//                 block_num: block_num.try_into().unwrap(),
//                 observation_timestamp: observation_timestamp.try_into().unwrap(),
//                 transmission_timestamp: transmission_timestamp.try_into().unwrap(),
//             }
//         )
//     }
//
//     fn write(
//         address_domain: u32, base: StorageBaseAddress, value: Transmission
//     ) -> SyscallResult::<()> {
//         let block_num_and_answer = value.block_num.into() * SHIFT_128 + value.answer.into();
//         let timestamps = value.observation_timestamp.into() * SHIFT_128
//             + value.transmission_timestamp.into();
//         storage_write_syscall(
//             address_domain,
//             storage_address_from_base_and_offset(base, 0_u8),
//             block_num_and_answer
//         )?;
//         storage_write_syscall(
//             address_domain, storage_address_from_base_and_offset(base, 1_u8), timestamps
//         )
//     }
// }

#[starknet::interface]
trait IAggregator<TContractState> {
    fn latest_round_data(self: @TContractState) -> Round;
    fn round_data(self: @TContractState, round_id: u128) -> Round;
    fn description(self: @TContractState) -> felt252;
    fn decimals(self: @TContractState) -> u8;
    fn latest_answer(self: @TContractState) -> u128;
}

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

#[starknet::interface]
trait Configuration<TContractState> {
    fn set_config(
        ref self: TContractState,
        oracles: Array<OracleConfig>,
        f: u8,
        onchain_config: Array<felt252>,
        offchain_config_version: u64,
        offchain_config: Array<felt252>,
    ) -> felt252; // digest
    fn latest_config_details(self: @TContractState) -> (u64, u64, felt252);
    fn transmitters(self: @TContractState) -> Array<ContractAddress>;
}

use Aggregator::{BillingConfig, BillingConfigSerde};

#[starknet::interface]
trait Billing<TContractState> {
    fn set_billing_access_controller(ref self: TContractState, access_controller: ContractAddress);
    fn set_billing(ref self: TContractState, config: Aggregator::BillingConfig);
    fn billing(self: @TContractState) -> Aggregator::BillingConfig;
    //
    fn withdraw_payment(ref self: TContractState, transmitter: ContractAddress);
    fn owed_payment(self: @TContractState, transmitter: ContractAddress) -> u128;
    fn withdraw_funds(ref self: TContractState, recipient: ContractAddress, amount: u256);
    fn link_available_for_payment(
        self: @TContractState
    ) -> (bool, u128); // (is negative, absolute difference)
    fn set_link_token(
        ref self: TContractState, link_token: ContractAddress, recipient: ContractAddress
    );
}

#[derive(Copy, Drop, Serde)]
struct PayeeConfig {
    transmitter: ContractAddress,
    payee: ContractAddress,
}

#[starknet::interface]
trait PayeeManagement<TContractState> {
    fn set_payees(ref self: TContractState, payees: Array<PayeeConfig>);
    fn transfer_payeeship(
        ref self: TContractState, transmitter: ContractAddress, proposed: ContractAddress
    );
    fn accept_payeeship(ref self: TContractState, transmitter: ContractAddress);
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

#[starknet::contract]
mod Aggregator {
    use super::Round;
    use super::{Transmission};
    use super::SpanLegacyHash;
    use super::pow;

    use array::ArrayTrait;
    use array::SpanTrait;
    use box::BoxTrait;
    use hash::LegacyHash;
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
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;
    use starknet::class_hash::ClassHash;

    use openzeppelin::access::ownable::OwnableComponent;
    use openzeppelin::token::erc20::interface::{IERC20, IERC20Dispatcher, IERC20DispatcherTrait};

    use chainlink::utils::split_felt;
    use chainlink::libraries::access_control::{AccessControlComponent, IAccessController};
    use chainlink::libraries::access_control::AccessControlComponent::InternalTrait as AccessControlInternalTrait;
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    use chainlink::libraries::access_control::{
        IAccessControllerDispatcher, IAccessControllerDispatcherTrait
    };
    use chainlink::libraries::type_and_version::ITypeAndVersion;

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);
    component!(path: AccessControlComponent, storage: access_control, event: AccessControlEvent);

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableTwoStepImpl<ContractState>;
    impl OwnableInternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[abi(embed_v0)]
    impl AccessControlImpl =
        AccessControlComponent::AccessControlImpl<ContractState>;
    impl AccessControlInternalImpl = AccessControlComponent::InternalImpl<ContractState>;

    const GIGA: u128 = 1000000000_u128;

    const MAX_ORACLES: u32 = 31_u32;

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event,
        #[flat]
        AccessControlEvent: AccessControlComponent::Event,
        NewTransmission: NewTransmission,
        ConfigSet: ConfigSet,
        LinkTokenSet: LinkTokenSet,
        BillingAccessControllerSet: BillingAccessControllerSet,
        BillingSet: BillingSet,
        OraclePaid: OraclePaid,
        PayeeshipTransferRequested: PayeeshipTransferRequested,
        PayeeshipTransferred: PayeeshipTransferred,
    }

    #[derive(Drop, starknet::Event)]
    struct NewTransmission {
        #[key]
        round_id: u128,
        answer: u128,
        #[key]
        transmitter: ContractAddress,
        observation_timestamp: u64,
        observers: felt252,
        observations: Array<u128>,
        juels_per_fee_coin: u128,
        gas_price: u128,
        config_digest: felt252,
        epoch_and_round: u64,
        reimbursement: u128
    }

    #[derive(Copy, Drop, Serde, starknet::Store)]
    struct Oracle {
        index: usize,
        // entire supply of LINK always fits into u96, so u128 is safe to use
        payment_juels: u128,
    }

    // TODO: reintroduce custom storage to save on space
    // impl OracleStorageAccess of StorageAccess<Oracle> {
    //     fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Oracle> {
    //         let value = storage_read_syscall(
    //             address_domain, storage_address_from_base_and_offset(base, 0_u8)
    //         )?;
    //         let (index, payment_juels) = split_felt(value);
    //         Result::Ok(Oracle { index: index.try_into().unwrap(), payment_juels,  })
    //     }
    //
    //     fn write(
    //         address_domain: u32, base: StorageBaseAddress, value: Oracle
    //     ) -> SyscallResult::<()> {
    //         let value = value.index.into() * SHIFT_128 + value.payment_juels.into();
    //         storage_write_syscall(
    //             address_domain, storage_address_from_base_and_offset(base, 0_u8), value
    //         )
    //     }
    // }

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
        #[substorage(v0)]
        access_control: AccessControlComponent::Storage,
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
        _billing: BillingConfig,
        // payee management
        _payees: LegacyMap<ContractAddress, ContractAddress>, // <transmitter, payment_address>
        _proposed_payees: LegacyMap<
            ContractAddress, ContractAddress
        > // <transmitter, payment_address>
    }

    #[generate_trait]
    impl AccessHelperImpl of AccessHelperTrait {
        fn _require_read_access(self: @ContractState) {
            let caller = starknet::info::get_caller_address();
            self.access_control.check_read_access(caller);
        }
    }

    #[abi(embed_v0)]
    impl TypeAndVersionImpl of ITypeAndVersion<ContractState> {
        fn type_and_version(self: @ContractState) -> felt252 {
            'ocr2/aggregator.cairo 1.0.0'
        }
    }

    #[abi(embed_v0)]
    impl AggregatorImpl of super::IAggregator<ContractState> {
        fn latest_round_data(self: @ContractState) -> Round {
            self._require_read_access();
            let latest_round_id = self._latest_aggregator_round_id.read();
            let transmission = self._transmissions.read(latest_round_id);
            Round {
                round_id: latest_round_id.into(),
                answer: transmission.answer,
                block_num: transmission.block_num,
                started_at: transmission.observation_timestamp,
                updated_at: transmission.transmission_timestamp,
            }
        }

        fn round_data(self: @ContractState, round_id: u128) -> Round {
            self._require_read_access();
            let transmission = self._transmissions.read(round_id);
            Round {
                round_id: round_id.into(),
                answer: transmission.answer,
                block_num: transmission.block_num,
                started_at: transmission.observation_timestamp,
                updated_at: transmission.transmission_timestamp,
            }
        }

        fn description(self: @ContractState) -> felt252 {
            self._require_read_access();
            self._description.read()
        }

        fn decimals(self: @ContractState) -> u8 {
            self._require_read_access();
            self._decimals.read()
        }

        fn latest_answer(self: @ContractState) -> u128 {
            self._require_read_access();
            let latest_round_id = self._latest_aggregator_round_id.read();
            let transmission = self._transmissions.read(latest_round_id);
            transmission.answer
        }
    }

    // ---

    #[constructor]
    fn constructor(
        ref self: ContractState,
        owner: ContractAddress,
        link: ContractAddress,
        min_answer: u128,
        max_answer: u128,
        billing_access_controller: ContractAddress,
        decimals: u8,
        description: felt252
    ) {
        self.ownable.initializer(owner);
        self.access_control.initializer();
        self._link_token.write(link);
        self._billing_access_controller.write(billing_access_controller);

        assert(min_answer < max_answer, 'min >= max');
        self._min_answer.write(min_answer);
        self._max_answer.write(max_answer);

        self._decimals.write(decimals);
        self._description.write(description);
    }

    // --- Upgradeable ---

    #[abi(embed_v0)]
    impl UpgradeableImpl of IUpgradeable<ContractState> {
        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            self.ownable.assert_only_owner();
            Upgradeable::upgrade(new_impl)
        }
    }

    // --- Validation ---

    // NOTE: Currently unimplemented:

    // --- Configuration

    #[derive(Drop, starknet::Event)]
    struct ConfigSet {
        #[key]
        previous_config_block_number: u64,
        #[key]
        latest_config_digest: felt252,
        config_count: u64,
        oracles: Array<OracleConfig>,
        f: u8,
        onchain_config: Array<felt252>,
        offchain_config_version: u64,
        offchain_config: Array<felt252>,
    }

    use super::OracleConfig;

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

    #[abi(embed_v0)]
    impl ConfigurationImpl of super::Configuration<ContractState> {
        fn set_config(
            ref self: ContractState,
            oracles: Array<OracleConfig>,
            f: u8,
            onchain_config: Array<felt252>,
            offchain_config_version: u64,
            offchain_config: Array<felt252>,
        ) -> felt252 { // digest
            self.ownable.assert_only_owner();
            assert(oracles.len() <= MAX_ORACLES, 'too many oracles');
            assert((3_u8 * f).into() < oracles.len(), 'faulty-oracle f too high');
            assert(f > 0_u8, 'f must be positive');

            assert(onchain_config.len() == 0_u32, 'onchain_config must be empty');

            let min_answer = self._min_answer.read();
            let max_answer = self._max_answer.read();

            let mut computed_onchain_config = ArrayTrait::new();
            computed_onchain_config.append(1); // version
            computed_onchain_config.append(min_answer.into());
            computed_onchain_config.append(max_answer.into());

            self.pay_oracles();

            // remove old signers & transmitters
            self.remove_oracles();

            let latest_round_id = self._latest_aggregator_round_id.read();

            self.add_oracles(@oracles, latest_round_id);

            self._f.write(f);
            let block_num = starknet::info::get_block_info().unbox().block_number;
            let prev_block_num = self._latest_config_block_number.read();
            self._latest_config_block_number.write(block_num);
            // update config count
            let mut config_count = self._config_count.read();
            config_count += 1_u64;
            self._config_count.write(config_count);
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

            self._latest_config_digest.write(digest);

            // reset epoch & round
            self._latest_epoch_and_round.write(0_u64);

            self
                .emit(
                    Event::ConfigSet(
                        ConfigSet {
                            previous_config_block_number: prev_block_num,
                            latest_config_digest: digest,
                            config_count: config_count,
                            oracles: oracles,
                            f: f,
                            onchain_config: computed_onchain_config,
                            offchain_config_version: offchain_config_version,
                            offchain_config: offchain_config
                        }
                    )
                );

            digest
        }

        fn latest_config_details(self: @ContractState) -> (u64, u64, felt252) {
            let config_count = self._config_count.read();
            let block_number = self._latest_config_block_number.read();
            let config_digest = self._latest_config_digest.read();

            (config_count, block_number, config_digest)
        }

        fn transmitters(self: @ContractState) -> Array<ContractAddress> {
            let mut index = 1;
            let mut len = self._oracles_len.read();
            let mut result = ArrayTrait::new();
            while len > 0_usize {
                let transmitter = self._transmitters_list.read(index);
                result.append(transmitter);
                len -= 1;
                index += 1;
            };
            return result;
        }
    }

    #[generate_trait]
    impl ConfigurationHelperImpl of ConfigurationHelperTrait {
        fn remove_oracles(ref self: ContractState) {
            let mut index = self._oracles_len.read();
            while index > 0_usize {
                let signer = self._signers_list.read(index);
                self._signers.write(signer, 0_usize);

                let transmitter = self._transmitters_list.read(index);
                self
                    ._transmitters
                    .write(transmitter, Oracle { index: 0_usize, payment_juels: 0_u128 });

                index -= 1;
            };
            self._oracles_len.write(0_usize);
        }

        fn add_oracles(
            ref self: ContractState, oracles: @Array<OracleConfig>, latest_round_id: u128
        ) {
            let mut index = 0;
            let mut span = oracles.span();
            while let Option::Some(oracle) = span
                .pop_front() {
                    // NOTE: index should start with 1 here because storage is 0-initialized.
                    // That way signers(pkey) => 0 indicates "not present"
                    // This is why we increment first, before using the index
                    index += 1;

                    // check for duplicates
                    let existing_signer = self._signers.read(*oracle.signer);
                    assert(existing_signer == 0_usize, 'repeated signer');

                    let existing_transmitter = self._transmitters.read(*oracle.transmitter);
                    assert(existing_transmitter.index == 0_usize, 'repeated transmitter');

                    self._signers.write(*oracle.signer, index);
                    self._signers_list.write(index, *oracle.signer);

                    self
                        ._transmitters
                        .write(*oracle.transmitter, Oracle { index, payment_juels: 0_u128 });
                    self._transmitters_list.write(index, *oracle.transmitter);

                    self._reward_from_aggregator_round_id.write(index, latest_round_id);
                };
            self._oracles_len.write(index);
        }
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

    #[abi(per_item)]
    #[generate_trait]
    impl TransmissionHelperImpl of TransmissionHelperTrait {
        fn hash_report(
            self: @ContractState,
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

        #[external(v0)]
        fn latest_transmission_details(self: @ContractState) -> (felt252, u64, u128, u64) {
            let config_digest = self._latest_config_digest.read();
            let latest_round_id = self._latest_aggregator_round_id.read();
            let epoch_and_round = self._latest_epoch_and_round.read();
            let transmission = self._transmissions.read(latest_round_id);
            (
                config_digest,
                epoch_and_round,
                transmission.answer,
                transmission.transmission_timestamp
            )
        }

        #[external(v0)]
        fn transmit(
            ref self: ContractState,
            report_context: ReportContext,
            observation_timestamp: u64,
            observers: felt252,
            observations: Array<u128>,
            juels_per_fee_coin: u128,
            gas_price: u128,
            mut signatures: Array<Signature>,
        ) {
            let signatures_len = signatures.len();

            let epoch_and_round = self._latest_epoch_and_round.read();
            assert(epoch_and_round < report_context.epoch_and_round, 'stale report');

            // validate transmitter
            let caller = starknet::info::get_caller_address();
            let mut oracle = self._transmitters.read(caller);
            assert(oracle.index != 0_usize, 'unknown sender'); // 0 index = uninitialized

            // Validate config digest matches latest_config_digest
            let config_digest = self._latest_config_digest.read();
            assert(report_context.config_digest == config_digest, 'config digest mismatch');

            let f = self._f.read();
            assert(signatures_len == (f + 1_u8).into(), 'wrong number of signatures');

            let msg = self
                .hash_report(
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
            self.verify_signatures(msg, ref signatures, 0_u128);

            // report():

            let observations_len = observations.len();
            assert(observations_len <= MAX_ORACLES, '');
            assert(f.into() < observations_len, '');

            self._latest_epoch_and_round.write(report_context.epoch_and_round);

            let median_idx = observations_len / 2_usize;
            let median = *observations[median_idx];

            // Validate median in min-max range
            let min_answer = self._min_answer.read();
            let max_answer = self._max_answer.read();
            assert(
                (min_answer <= median) & (median <= max_answer), 'median is out of min-max range'
            );

            let prev_round_id = self._latest_aggregator_round_id.read();
            let round_id = prev_round_id + 1_u128;
            self._latest_aggregator_round_id.write(round_id);

            let block_info = starknet::info::get_block_info().unbox();

            self
                ._transmissions
                .write(
                    round_id,
                    Transmission {
                        answer: median,
                        block_num: block_info.block_number,
                        observation_timestamp,
                        transmission_timestamp: block_info.block_timestamp,
                    }
                );

            // NOTE: Usually validating via validator would happen here, currently disabled

            let billing = self._billing.read();
            let reimbursement_juels = calculate_reimbursement(
                juels_per_fee_coin, signatures_len, gas_price, billing
            );

            // end report()

            self
                .emit(
                    Event::NewTransmission(
                        NewTransmission {
                            round_id: round_id,
                            answer: median,
                            transmitter: caller,
                            observation_timestamp: observation_timestamp,
                            observers: observers,
                            observations: observations,
                            juels_per_fee_coin: juels_per_fee_coin,
                            gas_price: gas_price,
                            config_digest: report_context.config_digest,
                            epoch_and_round: report_context.epoch_and_round,
                            reimbursement: reimbursement_juels,
                        }
                    )
                );

            // pay transmitter
            let payment = reimbursement_juels + (billing.transmission_payment_gjuels.into() * GIGA);
            // TODO: check overflow

            oracle.payment_juels += payment;
            self._transmitters.write(caller, oracle);
        }

        fn verify_signatures(
            self: @ContractState,
            msg: felt252,
            ref signatures: Array<Signature>,
            mut signed_count: u128
        ) {
            let mut span = signatures.span();
            while let Option::Some(signature) = span
                .pop_front() {
                    let index = self._signers.read(*signature.public_key);
                    assert(index != 0_usize, 'invalid signer'); // 0 index == uninitialized

                    let indexed_bit = pow(2_u128, index.into() - 1_u128);
                    let prev_signed_count = signed_count;
                    signed_count = signed_count | indexed_bit;
                    assert(prev_signed_count != signed_count, 'duplicate signer');

                    let is_valid = ecdsa::check_ecdsa_signature(
                        msg, *signature.public_key, *signature.r, *signature.s
                    );

                    assert(is_valid, '');
                };
        }
    }

    // --- Billing Config

    #[derive(Copy, Drop, Serde, starknet::Store)]
    struct BillingConfig {
        observation_payment_gjuels: u32,
        transmission_payment_gjuels: u32,
        gas_base: u32,
        gas_per_signature: u32,
    }

    // --- Billing Access Controller

    #[derive(Drop, starknet::Event)]
    struct BillingAccessControllerSet {
        #[key]
        old_controller: ContractAddress,
        #[key]
        new_controller: ContractAddress,
    }

    #[derive(Drop, starknet::Event)]
    struct BillingSet {
        config: BillingConfig
    }

    #[derive(Drop, starknet::Event)]
    struct OraclePaid {
        #[key]
        transmitter: ContractAddress,
        payee: ContractAddress,
        amount: u256,
        link_token: ContractAddress,
    }

    #[derive(Drop, starknet::Event)]
    struct LinkTokenSet {
        #[key]
        old_link_token: ContractAddress,
        #[key]
        new_link_token: ContractAddress
    }

    #[abi(embed_v0)]
    impl BillingImpl of super::Billing<ContractState> {
        fn set_link_token(
            ref self: ContractState, link_token: ContractAddress, recipient: ContractAddress
        ) {
            self.ownable.assert_only_owner();

            let old_token = self._link_token.read();

            if link_token == old_token {
                return ();
            }

            let contract_address = starknet::info::get_contract_address();

            // call balanceOf as a sanity check to confirm we're talking to a token
            let token = IERC20Dispatcher { contract_address: link_token };
            token.balance_of(account: contract_address);

            self.pay_oracles();

            // transfer remaining balance of old token to recipient
            let old_token_dispatcher = IERC20Dispatcher { contract_address: old_token };
            let amount = old_token_dispatcher.balance_of(account: contract_address);
            old_token_dispatcher.transfer(recipient, amount);

            self._link_token.write(link_token);

            self
                .emit(
                    Event::LinkTokenSet(
                        LinkTokenSet { old_link_token: old_token, new_link_token: link_token }
                    )
                );
        }

        fn set_billing_access_controller(
            ref self: ContractState, access_controller: ContractAddress
        ) {
            self.ownable.assert_only_owner();

            let old_controller = self._billing_access_controller.read();
            if access_controller == old_controller {
                return ();
            }

            self._billing_access_controller.write(access_controller);
            self
                .emit(
                    Event::BillingAccessControllerSet(
                        BillingAccessControllerSet {
                            old_controller: old_controller, new_controller: access_controller
                        }
                    )
                );
        }

        fn set_billing(ref self: ContractState, config: BillingConfig) {
            self.has_billing_access();

            self.pay_oracles();

            self._billing.write(config);

            self.emit(Event::BillingSet(BillingSet { config: config }));
        }

        fn billing(self: @ContractState) -> BillingConfig {
            self._billing.read()
        }

        // Payments and Withdrawals

        fn withdraw_payment(ref self: ContractState, transmitter: ContractAddress) {
            let caller = starknet::info::get_caller_address();
            let payee = self._payees.read(transmitter);
            assert(caller == payee, 'only payee can withdraw');

            let latest_round_id = self._latest_aggregator_round_id.read();
            let link_token = self._link_token.read();
            self.pay_oracle(transmitter, latest_round_id, link_token)
        }

        fn owed_payment(self: @ContractState, transmitter: ContractAddress) -> u128 {
            let oracle = self._transmitters.read(transmitter);
            self._owed_payment(@oracle)
        }

        fn withdraw_funds(ref self: ContractState, recipient: ContractAddress, amount: u256) {
            self.has_billing_access();

            let link_token = self._link_token.read();
            let contract_address = starknet::info::get_contract_address();

            let due = self.total_link_due();
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

        fn link_available_for_payment(
            self: @ContractState
        ) -> (bool, u128) { // (is negative, absolute difference)
            let link_token = self._link_token.read();
            let contract_address = starknet::info::get_contract_address();

            let token = IERC20Dispatcher { contract_address: link_token };
            let balance = token.balance_of(account: contract_address);
            // entire link supply fits into u96 so this should not fail
            assert(balance.high == 0_u128, 'balance too high');
            let balance: u128 = balance.low;

            let due = self.total_link_due();
            if balance > due {
                (false, balance - due)
            } else {
                (true, due - balance)
            }
        }
    }

    #[generate_trait]
    impl BillingHelperImpl of BillingHelperTrait {
        fn has_billing_access(self: @ContractState) {
            let caller = starknet::info::get_caller_address();
            let owner = self.ownable.owner();

            // owner always has access
            if caller == owner {
                return ();
            }

            let access_controller = self._billing_access_controller.read();
            let access_controller = IAccessControllerDispatcher {
                contract_address: access_controller
            };
            assert(
                access_controller.has_access(caller, ArrayTrait::new()),
                'caller does not have access'
            );
        }

        // --- Payments and Withdrawals

        fn _owed_payment(self: @ContractState, oracle: @Oracle) -> u128 {
            if *oracle.index == 0_usize {
                return 0_u128;
            }

            let billing = self._billing.read();

            let latest_round_id = self._latest_aggregator_round_id.read();
            let from_round_id = self._reward_from_aggregator_round_id.read(*oracle.index);
            let rounds = latest_round_id - from_round_id;

            (rounds * billing.observation_payment_gjuels.into() * GIGA) + *oracle.payment_juels
        }

        fn pay_oracle(
            ref self: ContractState,
            transmitter: ContractAddress,
            latest_round_id: u128,
            link_token: ContractAddress
        ) {
            let oracle = self._transmitters.read(transmitter);
            if oracle.index == 0_usize {
                return ();
            }

            let amount = self._owed_payment(@oracle);
            // if zero, fastpath return to avoid empty transfers
            if amount == 0_u128 {
                return ();
            }

            let payee = self._payees.read(transmitter);

            // NOTE: equivalent to converting u128 to u256
            let amount = u256 { high: 0_u128, low: amount };

            let token = IERC20Dispatcher { contract_address: link_token };
            token.transfer(recipient: payee, amount: amount);

            // Reset payment
            self._reward_from_aggregator_round_id.write(oracle.index, latest_round_id);
            self
                ._transmitters
                .write(transmitter, Oracle { index: oracle.index, payment_juels: 0_u128 });

            self
                .emit(
                    Event::OraclePaid(
                        OraclePaid {
                            transmitter: transmitter,
                            payee: payee,
                            amount: amount,
                            link_token: link_token
                        }
                    )
                );
        }

        fn pay_oracles(ref self: ContractState) {
            let mut index = self._oracles_len.read();
            let latest_round_id = self._latest_aggregator_round_id.read();
            let link_token = self._link_token.read();
            while index > 0_usize {
                let transmitter = self._transmitters_list.read(index);
                self.pay_oracle(transmitter, latest_round_id, link_token);
                index -= 1;
            };
        }

        fn total_link_due(self: @ContractState) -> u128 {
            let mut index = self._oracles_len.read();
            let latest_round_id = self._latest_aggregator_round_id.read();
            let mut total_rounds = 0;
            let mut payments_juels = 0;

            loop {
                if index == 0_usize {
                    break ();
                }
                let transmitter = self._transmitters_list.read(index);
                let oracle = self._transmitters.read(transmitter);
                assert(oracle.index != 0_usize, index.into()); // 0 == undefined

                let from_round_id = self._reward_from_aggregator_round_id.read(oracle.index);
                let rounds = latest_round_id - from_round_id;
                total_rounds += rounds;
                payments_juels += oracle.payment_juels;
                index -= 1;
            };

            let billing = self._billing.read();
            return (total_rounds * billing.observation_payment_gjuels.into() * GIGA)
                + payments_juels;
        }
    }

    // --- Transmitter Payment

    const MARGIN: u128 = 115_u128;

    fn calculate_reimbursement(
        juels_per_fee_coin: u128, signature_count: usize, gas_price: u128, config: BillingConfig
    ) -> u128 {
        // TODO: determine new values for these constants
        // Based on estimateFee (f=1 14977, f=2 14989, f=3 15002 f=4 15014 f=5 15027, count = f+1)
        // gas_base = 14951, gas_per_signature = 13
        let signature_count_u128: u128 = signature_count.into();
        let gas_base_u128: u128 = config.gas_base.into();
        let gas_per_signature_u128: u128 = config.gas_per_signature.into();

        let exact_gas = gas_base_u128 + (signature_count_u128 * gas_per_signature_u128);
        let gas = exact_gas * MARGIN / 100_u128; // scale to 115% for some margin
        let amount = gas * gas_price;
        amount * juels_per_fee_coin
    }

    // --- Payee Management

    use super::PayeeConfig;

    #[derive(Drop, starknet::Event)]
    struct PayeeshipTransferRequested {
        #[key]
        transmitter: ContractAddress,
        #[key]
        current: ContractAddress,
        #[key]
        proposed: ContractAddress,
    }

    #[derive(Drop, starknet::Event)]
    struct PayeeshipTransferred {
        #[key]
        transmitter: ContractAddress,
        #[key]
        previous: ContractAddress,
        #[key]
        current: ContractAddress,
    }

    #[abi(embed_v0)]
    impl PayeeManagementImpl of super::PayeeManagement<ContractState> {
        fn set_payees(ref self: ContractState, mut payees: Array<PayeeConfig>) {
            self.ownable.assert_only_owner();
            while let Option::Some(payee) = payees
                .pop_front() {
                    let current_payee = self._payees.read(payee.transmitter);
                    let is_unset = current_payee.is_zero();
                    let is_same = current_payee == payee.payee;
                    assert(is_unset | is_same, 'payee already set');

                    self._payees.write(payee.transmitter, payee.payee);

                    self
                        .emit(
                            Event::PayeeshipTransferred(
                                PayeeshipTransferred {
                                    transmitter: payee.transmitter,
                                    previous: current_payee,
                                    current: payee.payee
                                }
                            )
                        );
                }
        }

        fn transfer_payeeship(
            ref self: ContractState, transmitter: ContractAddress, proposed: ContractAddress
        ) {
            assert(!proposed.is_zero(), 'cannot transfer to zero address');
            let caller = starknet::info::get_caller_address();
            let payee = self._payees.read(transmitter);
            assert(caller == payee, 'only current payee can update');
            assert(caller != proposed, 'cannot transfer to self');

            self._proposed_payees.write(transmitter, proposed);
            self
                .emit(
                    Event::PayeeshipTransferRequested(
                        PayeeshipTransferRequested {
                            transmitter: transmitter, current: payee, proposed: proposed
                        }
                    )
                );
        }

        fn accept_payeeship(ref self: ContractState, transmitter: ContractAddress) {
            let proposed = self._proposed_payees.read(transmitter);
            let caller = starknet::info::get_caller_address();
            assert(caller == proposed, 'only proposed payee can accept');
            let previous = self._payees.read(transmitter);

            self._payees.write(transmitter, proposed);
            self._proposed_payees.write(transmitter, Zeroable::zero());
            self
                .emit(
                    Event::PayeeshipTransferred(
                        PayeeshipTransferred {
                            transmitter: transmitter, previous: previous, current: caller
                        }
                    )
                );
        }
    }
}
