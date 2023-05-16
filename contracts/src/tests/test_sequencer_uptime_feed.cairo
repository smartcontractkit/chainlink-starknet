use starknet::ContractAddress;
use starknet::EthAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::class_hash::Felt252TryIntoClassHash;
use starknet::syscalls::deploy_syscall;
use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::emergency::sequencer_uptime_feed::SequencerUptimeFeed;
use chainlink::ocr2::aggregator_proxy::AggregatorProxy;
use chainlink::tests::test_ownable::should_implement_ownable;
use chainlink::tests::test_access_controller::should_implement_access_control;

// NOTE: move to separate interface file once we can directly use the trait
#[abi]
trait ISequencerUptimeFeed {
    fn l1_sender() -> EthAddress;
    fn set_l1_sender(address: EthAddress);
}

fn setup() -> (ContractAddress, ContractAddress, ISequencerUptimeFeedDispatcher) {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);

    // Deploy seqeuencer uptime feed
    let mut calldata = ArrayTrait::new();
    calldata.append(0); // initial status
    calldata.append(account.into()); // owner
    let (sequencerFeedAddr, _) = deploy_syscall(
        SequencerUptimeFeed::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    ).unwrap();
    let sequencerUptimeFeed = ISequencerUptimeFeedDispatcher {
        contract_address: sequencerFeedAddr
    };

    (account, sequencerFeedAddr, sequencerUptimeFeed)
}

#[test]
#[available_gas(2000000)]
fn test_ownable() {
    let (account, sequencerFeedAddr, _) = setup();
    should_implement_ownable(sequencerFeedAddr, account);
}

#[test]
#[available_gas(2000000)]
fn test_access_control() {
    let (account, sequencerFeedAddr, _) = setup();
    should_implement_access_control(sequencerFeedAddr, account);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', 'ENTRYPOINT_FAILED'))]
fn test_set_l1_sender_not_owner() {
    let (_, _, sequencerUptimeFeed) = setup();
    sequencerUptimeFeed.set_l1_sender(contract_address_const::<789>().into().try_into().unwrap());
}

#[test]
#[available_gas(2000000)]
fn test_set_l1_sender() {
    let (owner, _, sequencerUptimeFeed) = setup();
    set_contract_address(owner);
    sequencerUptimeFeed.set_l1_sender(contract_address_const::<789>().into().try_into().unwrap());
    assert(
        sequencerUptimeFeed.l1_sender() == contract_address_const::<789>().into().try_into().unwrap(),
        'l1_sender should be set to 789'
    );
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('user does not have read access', ))]
fn test_latest_round_data_no_access() {
    let (owner, sequencerFeedAddr, _) = setup();
    AggregatorProxy::constructor(owner, sequencerFeedAddr);
    AggregatorProxy::latest_round_data();
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('user does not have read access', ))]
fn test_aggregator_proxy_response() {
    let (owner, sequencerFeedAddr, _) = setup();
    AggregatorProxy::constructor(owner, sequencerFeedAddr);

    // latest round data
    let latest_round_data = AggregatorProxy::latest_round_data();
    assert(latest_round_data.answer == 0, 'latest_round_data should be 0');

    // description
    let description = AggregatorProxy::description();
    assert(description == 'L2 Sequencer Uptime Status Feed', 'description does not match');

    // decimals
    let decimals = AggregatorProxy::decimals();
    assert(decimals == 0, 'decimals should be 0');
}
