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

use chainlink::ocr2::mocks::mock_aggregator::{
    MockAggregator, IMockAggregator, IMockAggregatorDispatcher, IMockAggregatorDispatcherTrait
};
use chainlink::ocr2::aggregator_proxy::AggregatorProxy;
use chainlink::ocr2::aggregator_proxy::AggregatorProxy::{
    AggregatorProxyImpl, AggregatorProxyInternal, UpgradeableImpl
};
use chainlink::libraries::access_control::AccessControlComponent::AccessControlImpl;
use chainlink::ocr2::aggregator::Round;
use chainlink::utils::split_felt;
use chainlink::tests::test_ownable::should_implement_ownable;
use chainlink::tests::test_access_controller::should_implement_access_control;

fn STATE() -> AggregatorProxy::ContractState {
    AggregatorProxy::contract_state_for_testing()
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
    )
        .unwrap();
    let mockAggregator1 = IMockAggregatorDispatcher { contract_address: mockAggregatorAddr1 };

    // Deploy mock aggregator 2
    // note: deployment address is deterministic based on deploy_syscall parameters
    // so we need to change the decimals parameter to avoid an address conflict with mock aggregator 1
    let mut calldata2 = ArrayTrait::new();
    calldata2.append(10); // decimals = 10
    let (mockAggregatorAddr2, _) = deploy_syscall(
        MockAggregator::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata2.span(), false
    )
        .unwrap();
    let mockAggregator2 = IMockAggregatorDispatcher { contract_address: mockAggregatorAddr2 };

    // Return account, mock aggregator address and mock aggregator contract
    (account, mockAggregatorAddr1, mockAggregator1, mockAggregatorAddr2, mockAggregator2)
}

#[test]
fn test_ownable() {
    let (account, mockAggregatorAddr, _, _, _) = setup();
    // Deploy aggregator proxy
    let calldata = array![account.into(), // owner = account
     mockAggregatorAddr.into(),];
    let (aggregatorProxyAddr, _) = deploy_syscall(
        AggregatorProxy::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();

    should_implement_ownable(aggregatorProxyAddr, account);
}

#[test]
fn test_access_control() {
    let (account, mockAggregatorAddr, _, _, _) = setup();
    // Deploy aggregator proxy
    let calldata = array![account.into(), // owner = account
     mockAggregatorAddr.into(),];
    let (aggregatorProxyAddr, _) = deploy_syscall(
        AggregatorProxy::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();

    should_implement_access_control(aggregatorProxyAddr, account);
}

#[test]
#[should_panic(expected: ('Caller is not the owner',))]
fn test_upgrade_non_owner() {
    let (_, _, _, _, _) = setup();
    let mut state = STATE();
    UpgradeableImpl::upgrade(ref state, class_hash_const::<123>());
}

fn test_query_latest_round_data() {
    let (owner, mockAggregatorAddr, mockAggregator, _, _) = setup();
    let mut state = STATE();
    // init aggregator proxy with mock aggregator
    AggregatorProxy::constructor(ref state, owner, mockAggregatorAddr);
    state.add_access(owner);
    // insert round into mock aggregator
    mockAggregator.set_latest_round_data(10, 1, 9, 8);
    // query latest round
    let round = AggregatorProxyImpl::latest_round_data(@state);
    let (phase_id, round_id) = split_felt(round.round_id);
    assert(phase_id == 1, 'phase_id should be 1');
    assert(round_id == 1, 'round_id should be 1');
    assert(round.answer == 10, 'answer should be 10');
    assert(round.block_num == 1, 'block_num should be 1');
    assert(round.started_at == 9, 'started_at should be 9');
    assert(round.updated_at == 8, 'updated_at should be 8');

    // latest_answer matches up with latest_round_data 
    let latest_answer = AggregatorProxyImpl::latest_answer(@state);
    assert(latest_answer == 10, '(latest) answer should be 10');
}

#[test]
#[should_panic(expected: ('user does not have read access',))]
fn test_query_latest_round_data_without_access() {
    let (owner, mockAggregatorAddr, mockAggregator, _, _) = setup();
    let mut state = STATE();
    // init aggregator proxy with mock aggregator
    AggregatorProxy::constructor(ref state, owner, mockAggregatorAddr);
    state.add_access(owner);
    // insert round into mock aggregator
    mockAggregator.set_latest_round_data(10, 1, 9, 8);
    // set caller to non-owner address with no read access
    set_caller_address(contract_address_const::<2>());
    // query latest round
    AggregatorProxyImpl::latest_round_data(@state);
}

#[test]
#[should_panic(expected: ('user does not have read access',))]
fn test_query_latest_answer_without_access() {
    let (owner, mockAggregatorAddr, mockAggregator, _, _) = setup();
    let mut state = STATE();
    // init aggregator proxy with mock aggregator
    AggregatorProxy::constructor(ref state, owner, mockAggregatorAddr);
    state.add_access(owner);
    // insert round into mock aggregator
    mockAggregator.set_latest_round_data(10, 1, 9, 8);
    // set caller to non-owner address with no read access
    set_caller_address(contract_address_const::<2>());
    // query latest round
    AggregatorProxyImpl::latest_answer(@state);
}

#[test]
fn test_propose_new_aggregator() {
    let (owner, mockAggregatorAddr1, mockAggregator1, mockAggregatorAddr2, mockAggregator2) =
        setup();
    let mut state = STATE();
    // init aggregator proxy with mock aggregator 1
    AggregatorProxy::constructor(ref state, owner, mockAggregatorAddr1);
    state.add_access(owner);
    // insert rounds into mock aggregators
    mockAggregator1.set_latest_round_data(10, 1, 9, 8);
    mockAggregator2.set_latest_round_data(12, 2, 10, 11);

    // propose new mock aggregator to AggregatorProxy
    AggregatorProxyInternal::propose_aggregator(ref state, mockAggregatorAddr2);

    // latest_round_data should return old aggregator round data
    let round = AggregatorProxyImpl::latest_round_data(@state);
    assert(round.answer == 10, 'answer should be 10');

    // mockAggregator2.set_latest_round_data(12, 2, 10, 11);

    // proposed_round_data should return new aggregator round data
    let proposed_round = AggregatorProxyInternal::proposed_latest_round_data(@state);
    assert(proposed_round.answer == 12, 'answer should be 12');

    // aggregator should still be set to the old aggregator
    let aggregator = AggregatorProxyInternal::aggregator(@state);
    assert(aggregator == mockAggregatorAddr1, 'aggregator should be old addr');
}

#[test]
fn test_confirm_new_aggregator() {
    let (owner, mockAggregatorAddr1, mockAggregator1, mockAggregatorAddr2, mockAggregator2) =
        setup();
    let mut state = STATE();
    // init aggregator proxy with mock aggregator 1
    AggregatorProxy::constructor(ref state, owner, mockAggregatorAddr1);
    state.add_access(owner);
    // insert rounds into mock aggregators
    mockAggregator1.set_latest_round_data(10, 1, 9, 8);
    mockAggregator2.set_latest_round_data(12, 2, 10, 11);

    // propose new mock aggregator to AggregatorProxy
    AggregatorProxyInternal::propose_aggregator(ref state, mockAggregatorAddr2);

    // confirm new mock aggregator
    AggregatorProxyInternal::confirm_aggregator(ref state, mockAggregatorAddr2);

    // aggregator should be set to the new aggregator
    let aggregator = AggregatorProxyInternal::aggregator(@state);
    assert(aggregator == mockAggregatorAddr2, 'aggregator should be new addr');

    // phase ID should be 2
    let phase_id = AggregatorProxyInternal::phase_id(@state);
    assert(phase_id == 2, 'phase_id should be 2');

    // latest_round_data should return new aggregator round data
    let round = AggregatorProxyImpl::latest_round_data(@state);
    assert(round.answer == 12, 'answer should be 12');
}
