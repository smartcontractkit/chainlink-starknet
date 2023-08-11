use starknet::ContractAddress;
use starknet::testing::set_caller_address;
use chainlink::ocr2::mocks::mock_aggregator::MockAggregator;
use starknet::contract_address_const;
use chainlink::ocr2::aggregator::Round;

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    // Set account as default caller
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
fn test_deploy() {
    setup();

    MockAggregator::constructor(18_u8);

    assert(MockAggregator::decimals() == 18_u8, 'decimals');

    let latest_round = MockAggregator::latest_round_data();

    let zeroed_round = Round {
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
#[available_gas(2000000)]
fn test_set_latest_round() {
    setup();

    MockAggregator::constructor(18_u8);

    MockAggregator::set_latest_round_data(777_u128, 777_u64, 777_u64, 777_u64);

    let expected_round = Round {
        round_id: 1, answer: 777_u128, block_num: 777_u64, started_at: 777_u64, updated_at: 777_u64
    };

    assert(MockAggregator::latest_round_data() == expected_round, 'round not equal');
}

