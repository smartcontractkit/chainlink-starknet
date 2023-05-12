use chainlink::ocr2::aggregator::Round;

trait IAggregatorProxy {
    fn latest_round_data() -> Round;
    fn round_data(round_id: felt252) -> Round;
    fn description() -> felt252;
    fn decimals() -> u8;
    fn type_and_version() -> felt252;
}

// TODO: is it possible not to duplicate this trait when we require the abi attribute?
#[abi]
trait IAggregator {
    fn latest_round_data() -> Round;
    fn round_data(round_id: u128) -> Round;
    fn description() -> felt252;
    fn decimals() -> u8;
    fn type_and_version() -> felt252;
}

#[contract]
mod AggregatorProxy {
    use super::IAggregatorProxy;
    use super::IAggregatorDispatcher;
    use super::IAggregatorDispatcherTrait;

    use integer::u128s_from_felt252;
    use option::OptionTrait;
    use traits::Into;
    use traits::TryInto;
    use zeroable::Zeroable;

    use starknet::ContractAddress;
    use starknet::ContractAddressIntoFelt252;
    use starknet::Felt252TryIntoContractAddress;
    use integer::Felt252TryIntoU128;
    use starknet::StorageAccess;
    use starknet::StorageBaseAddress;
    use starknet::SyscallResult;
    use integer::U128IntoFelt252;
    use integer::U128sFromFelt252Result;
    use starknet::storage_read_syscall;
    use starknet::storage_write_syscall;
    use starknet::storage_address_from_base_and_offset;
    use starknet::class_hash::ClassHash;

    use chainlink::ocr2::aggregator::IAggregator;
    use chainlink::ocr2::aggregator::Round;
    use chainlink::libraries::ownable::Ownable;
<<<<<<< HEAD
    use chainlink::libraries::simple_read_access_controller::SimpleReadAccessController;
    use chainlink::libraries::simple_write_access_controller::SimpleWriteAccessController;
=======
    use chainlink::libraries::access_control::AccessControl;
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
    use chainlink::utils::split_felt;
    use chainlink::libraries::upgradeable::Upgradeable;

    const SHIFT: felt252 = 0x100000000000000000000000000000000_felt252;
    const MAX_ID: felt252 = 0xffffffffffffffffffffffffffffffff_felt252;

    #[derive(Copy, Drop, Serde)]
    struct Phase {
        id: u128,
        aggregator: ContractAddress
    }

    impl PhaseStorageAccess of StorageAccess<Phase> {
        fn read(address_domain: u32, base: StorageBaseAddress) -> SyscallResult::<Phase> {
            Result::Ok(
                Phase {
                    id: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 0_u8)
                    )?.try_into().unwrap(),
                    aggregator: storage_read_syscall(
                        address_domain, storage_address_from_base_and_offset(base, 1_u8)
                    )?.try_into().unwrap(),
                }
            )
        }

        fn write(
            address_domain: u32, base: StorageBaseAddress, value: Phase
        ) -> SyscallResult::<()> {
            storage_write_syscall(
                address_domain, storage_address_from_base_and_offset(base, 0_u8), value.id.into()
            )?;
            storage_write_syscall(
                address_domain,
                storage_address_from_base_and_offset(base, 1_u8),
                value.aggregator.into()
            )
        }
    }

    struct Storage {
        _current_phase: Phase,
        _proposed_aggregator: ContractAddress,
        _phases: LegacyMap<u128, ContractAddress>
    }

    #[event]
    fn AggregatorProposed(current: ContractAddress, proposed: ContractAddress) {}

    #[event]
    fn AggregatorConfirmed(previous: ContractAddress, latest: ContractAddress) {}

    impl AggregatorProxy of IAggregatorProxy {
        fn latest_round_data() -> Round {
<<<<<<< HEAD
            _require_access();
=======
            _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
            let phase = _current_phase::read();
            let aggregator = IAggregatorDispatcher { contract_address: phase.aggregator };
            let round = aggregator.latest_round_data();
            Round {
                round_id: (phase.id.into() * SHIFT) + round.round_id,
                answer: round.answer,
                block_num: round.block_num,
                started_at: round.started_at,
                updated_at: round.updated_at,
            }
        }
        fn round_data(round_id: felt252) -> Round {
<<<<<<< HEAD
            _require_access();
=======
            _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
            let (phase_id, round_id) = split_felt(round_id);
            let address = _phases::read(phase_id);
            assert(!address.is_zero(), 'aggregator address is 0');

            let aggregator = IAggregatorDispatcher { contract_address: address };
            let round = aggregator.round_data(round_id);

            Round {
                round_id: (phase_id.into() * SHIFT) + round.round_id,
                answer: round.answer,
                block_num: round.block_num,
                started_at: round.started_at,
                updated_at: round.updated_at,
            }
        }
        fn description() -> felt252 {
<<<<<<< HEAD
            _require_access();
=======
            _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
            let phase = _current_phase::read();
            let aggregator = IAggregatorDispatcher { contract_address: phase.aggregator };
            aggregator.description()
        }

        fn decimals() -> u8 {
<<<<<<< HEAD
            _require_access();
=======
            _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
            let phase = _current_phase::read();
            let aggregator = IAggregatorDispatcher { contract_address: phase.aggregator };
            aggregator.decimals()
        }

        fn type_and_version() -> felt252 {
            let phase = _current_phase::read();
            let aggregator = IAggregatorDispatcher { contract_address: phase.aggregator };
            aggregator.type_and_version()
        }
    }

    #[constructor]
    fn constructor(owner: ContractAddress, address: ContractAddress) {
<<<<<<< HEAD
        SimpleReadAccessController::initializer(owner);
=======
        Ownable::initializer(owner);
        AccessControl::initializer();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
        _set_aggregator(address);
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

    // -- Upgradeable -- 

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl)
    }

<<<<<<< HEAD
    // -- SimpleReadAccessController --

    #[external]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        SimpleReadAccessController::has_access(user, data)
    }

    #[external]
    fn check_access(user: ContractAddress) {
        SimpleReadAccessController::check_access(user)
    }

    // -- SimpleWriteAccessController --

    #[external]
    fn add_access(user: ContractAddress) {
        SimpleWriteAccessController::add_access(user)
=======
    // -- Access Control --

    #[view]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        AccessControl::has_access(user, data)
    }

    #[external]
    fn add_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        AccessControl::add_access(user)
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
    }

    #[external]
    fn remove_access(user: ContractAddress) {
<<<<<<< HEAD
        SimpleWriteAccessController::remove_access(user)
=======
        Ownable::assert_only_owner();
        AccessControl::remove_access(user)
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
    }

    #[external]
    fn enable_access_check() {
<<<<<<< HEAD
        SimpleWriteAccessController::enable_access_check()
=======
        Ownable::assert_only_owner();
        AccessControl::enable_access_check()
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
    }

    #[external]
    fn disable_access_check() {
<<<<<<< HEAD
        SimpleWriteAccessController::disable_access_check()
=======
        Ownable::assert_only_owner();
        AccessControl::disable_access_check()
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
    }

    //

    #[external]
    fn propose_aggregator(address: ContractAddress) {
        Ownable::assert_only_owner();
        assert(!address.is_zero(), 'proposed address is 0');
        _proposed_aggregator::write(address);

        let phase = _current_phase::read();
        AggregatorProposed(phase.aggregator, address);
    }

    #[external]
    fn confirm_aggregator(address: ContractAddress) {
        Ownable::assert_only_owner();
        assert(!address.is_zero(), 'confirm address is 0');
        let phase = _current_phase::read();
        let previous = phase.aggregator;

        let proposed_aggregator = _proposed_aggregator::read();
        assert(address == proposed_aggregator, 'does not match proposed address');
        _proposed_aggregator::write(starknet::contract_address_const::<0>());
        _set_aggregator(proposed_aggregator);

        AggregatorConfirmed(previous, address);
    }

    #[view]
    fn proposed_latest_round_data() -> Round {
<<<<<<< HEAD
        _require_access();
=======
        _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
        let address = _proposed_aggregator::read();
        let aggregator = IAggregatorDispatcher { contract_address: address };
        aggregator.latest_round_data()
    }

    #[view]
    fn proposed_round_data(round_id: felt252) -> Round {
<<<<<<< HEAD
        _require_access();
=======
        _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
        let address = _proposed_aggregator::read();
        let round_id128: u128 = round_id.try_into().unwrap();
        let aggregator = IAggregatorDispatcher { contract_address: address };
        aggregator.round_data(round_id128)
    }

    #[view]
    fn aggregator() -> ContractAddress {
<<<<<<< HEAD
        _require_access();
=======
        _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
        let phase = _current_phase::read();
        phase.aggregator
    }

    #[view]
    fn phase_id() -> u128 {
<<<<<<< HEAD
        _require_access();
=======
        _require_read_access();
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
        let phase = _current_phase::read();
        phase.id
    }

    #[view]
    fn latest_round_data() -> Round {
        AggregatorProxy::latest_round_data()
    }

    #[view]
    fn round_data(round_id: felt252) -> Round {
        AggregatorProxy::round_data(round_id)
    }

    #[view]
    fn description() -> felt252 {
        AggregatorProxy::description()
    }

    #[view]
    fn decimals() -> u8 {
        AggregatorProxy::decimals()
    }

    #[view]
    fn type_and_version() -> felt252 {
        AggregatorProxy::type_and_version()
    }

    /// Internals

    fn _set_aggregator(address: ContractAddress) {
        let phase = _current_phase::read();
        let new_phase_id = phase.id + 1_u128;
        _current_phase::write(Phase { id: new_phase_id, aggregator: address });
        _phases::write(new_phase_id, address);
    }

<<<<<<< HEAD
    fn _require_access() {
        let caller = starknet::info::get_caller_address();
        SimpleReadAccessController::check_access(caller);
=======
    fn _require_read_access() {
        let caller = starknet::info::get_caller_address();
        AccessControl::check_read_access(caller);
>>>>>>> 46ee16921b53945562670843862fd7d759b4a629
    }
}
