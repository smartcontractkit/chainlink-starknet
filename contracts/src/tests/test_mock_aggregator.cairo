use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use chainlink::ocr2::mocks::mock_aggregator::MockAggregator;
use starknet::contract_address_const;
use chainlink::ocr2::aggregator::Round;

fn STATE() -> MockAggregator::ContractState {
    MockAggregator::contract_state_for_testing()
}

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    // Set account as default caller
    set_caller_address(account);
    account
}

#[test]
fn test_deploy() {
    setup();

    let mut state = STATE();

    MockAggregator::constructor(ref state, 18_u8);

    assert(MockAggregator::Aggregator::decimals(@state) == 18_u8, 'decimals');

    let latest_round = MockAggregator::Aggregator::latest_round_data(@state);

    let _ = Round {
        round_id: 0, answer: 0_u128, block_num: 0_u64, started_at: 0_u64, updated_at: 0_u64
    };

    assert(
        latest_round == Round {
            round_id: 0, answer: 0_u128, block_num: 0_u64, started_at: 0_u64, updated_at: 0_u64
        },
        'rounds'
    );
}

#[test]
fn test_set_latest_round() {
    setup();

    let mut state = STATE();

    MockAggregator::constructor(ref state, 18_u8);

    MockAggregator::MockImpl::set_latest_round_data(ref state, 777_u128, 777_u64, 777_u64, 777_u64);

    let expected_round = Round {
        round_id: 1, answer: 777_u128, block_num: 777_u64, started_at: 777_u64, updated_at: 777_u64
    };

    assert(
        MockAggregator::Aggregator::latest_round_data(@state) == expected_round, 'round not equal'
    );

    assert(
        MockAggregator::Aggregator::latest_answer(@state) == expected_round.answer,
        'latest answer not equal'
    );
}

