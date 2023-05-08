use starknet::contract_address_const;
use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use starknet::syscalls::deploy_syscall;
use starknet::class_hash::Felt252TryIntoClassHash;
use starknet::class_hash::class_hash_const;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::ocr2::mocks::mock_aggregator::MockAggregator;
use chainlink::ocr2::aggregator_proxy::AggregatorProxy;
use chainlink::ocr2::aggregator::Round;
use chainlink::utils::split_felt;

#[abi]
trait IMockAggregator {
    fn set_latest_round_data(
        answer: u128, block_num: u64, observation_timestamp: u64, transmission_timestamp: u64
    );
    fn latest_round_data() -> Round;
    fn decimals() -> u8;
}

fn setup() -> (
    ContractAddress,
    ContractAddress,
    IMockAggregatorDispatcher,
    ContractAddress,
    IMockAggregatorDispatcher
) {
    // Set account as default caller
    let account: ContractAddress = contract_address_const::<1>();
    set_caller_address(account);

    // Deploy mock aggregator 1
    let mut calldata = ArrayTrait::new();
    calldata.append(8); // decimals = 8
    let (mockAggregatorAddr1, _) = deploy_syscall(
        MockAggregator::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();
    let mockAggregator1 = IMockAggregatorDispatcher { contract_address: mockAggregatorAddr1 };

    // Deploy mock aggregator 2
    let (mockAggregatorAddr2, _) = deploy_syscall(
        MockAggregator::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();
    let mockAggregator2 = IMockAggregatorDispatcher { contract_address: mockAggregatorAddr2 };

    // Return account, mock aggregator address and mock aggregator contract
    (account, mockAggregatorAddr1, mockAggregator1, mockAggregatorAddr2, mockAggregator2)
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_non_owner() {
    let (_, _, _, _, _) = setup();
    AggregatorProxy::upgrade(class_hash_const::<123>());
}

fn test_query_latest_round_data() {
    let (owner, mockAggregatorAddr, mockAggregator, _, _) = setup();
    // init aggregator proxy with mock aggregator
    AggregatorProxy::constructor(owner, mockAggregatorAddr);
    AggregatorProxy::add_access(owner);
    // insert round into mock aggregator
    mockAggregator.set_latest_round_data(10, 1, 9, 8);
    // query latest round
    let round = AggregatorProxy::latest_round_data();
    let (phase_id, round_id) = split_felt(round.round_id);
    assert(phase_id == 1, 'phase_id should be 1');
    assert(round_id == 1, 'round_id should be 1');
    assert(round.answer == 10, 'answer should be 10');
    assert(round.block_num == 1, 'block_num should be 1');
    assert(round.started_at == 9, 'started_at should be 9');
    assert(round.updated_at == 8, 'updated_at should be 8');
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('address does not have access', ))]
fn test_query_latest_round_data_without_access() {
    let (owner, mockAggregatorAddr, mockAggregator, _, _) = setup();
    // init aggregator proxy with mock aggregator
    AggregatorProxy::constructor(owner, mockAggregatorAddr);
    AggregatorProxy::add_access(owner);
    // insert round into mock aggregator
    mockAggregator.set_latest_round_data(10, 1, 9, 8);
    // set caller to non-owner address with no read access
    set_caller_address(contract_address_const::<2>());
    // query latest round
    AggregatorProxy::latest_round_data();
}

#[test]
#[available_gas(2000000)]
fn test_propose_new_aggregator() {
    let (owner, mockAggregatorAddr1, mockAggregator1, mockAggregatorAddr2, mockAggregator2) =
        setup();
    // init aggregator proxy with mock aggregator 1
    AggregatorProxy::constructor(owner, mockAggregatorAddr1);
    AggregatorProxy::add_access(owner);
    // insert rounds into mock aggregators
    mockAggregator1.set_latest_round_data(10, 1, 9, 8);
    mockAggregator2.set_latest_round_data(12, 2, 10, 11);

    // propose new mock aggregator to AggregatorProxy
    AggregatorProxy::propose_aggregator(mockAggregatorAddr2);

    // latest_round_data should return old aggregator round data
    let round = AggregatorProxy::latest_round_data();
    assert(round.answer == 10, 'answer should be 10');

    // proposed_round_data should return new aggregator round data
    let proposed_round = AggregatorProxy::proposed_latest_round_data();
    assert(proposed_round.answer == 12, 'answer should be 12');

    // aggregator should still be set to the old aggregator
    let aggregator = AggregatorProxy::aggregator();
    assert(aggregator == mockAggregatorAddr1, 'aggregator should be old addr');
}

#[test]
#[available_gas(2000000)]
fn test_confirm_new_aggregator() {
    let (owner, mockAggregatorAddr1, mockAggregator1, mockAggregatorAddr2, mockAggregator2) =
        setup();
    // init aggregator proxy with mock aggregator 1
    AggregatorProxy::constructor(owner, mockAggregatorAddr1);
    AggregatorProxy::add_access(owner);
    // insert rounds into mock aggregators
    mockAggregator1.set_latest_round_data(10, 1, 9, 8);
    mockAggregator2.set_latest_round_data(12, 2, 10, 11);

    // propose new mock aggregator to AggregatorProxy
    AggregatorProxy::propose_aggregator(mockAggregatorAddr2);

    // confirm new mock aggregator
    AggregatorProxy::confirm_aggregator(mockAggregatorAddr2);

    // aggregator should be set to the new aggregator
    let aggregator = AggregatorProxy::aggregator();
    assert(aggregator == mockAggregatorAddr2, 'aggregator should be new addr');

    // phase ID should be 2
    let phase_id = AggregatorProxy::phase_id();
    assert(phase_id == 2, 'phase_id should be 2');

    // latest_round_data should return new aggregator round data
    let round = AggregatorProxy::latest_round_data();
    assert(round.answer == 12, 'answer should be 12');
}
