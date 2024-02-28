use snforge_std::{declare, ContractClassTrait};

use aggregator_consumer::mocks::mock_aggregator::IMockAggregatorDispatcherTrait;
use aggregator_consumer::mocks::mock_aggregator::IMockAggregatorDispatcher;
use aggregator_consumer::ocr2::consumer::IAggregatorConsumerDispatcherTrait;
use aggregator_consumer::ocr2::consumer::IAggregatorConsumerDispatcher;

use starknet::ContractAddress;

fn deploy_mock_aggregator(decimals: u8) -> ContractAddress {
    let mut calldata = ArrayTrait::new();
    calldata.append(decimals.into());
    return declare('MockAggregator').deploy(@calldata).unwrap();
}

fn deploy_consumer(aggregator_address: ContractAddress) -> ContractAddress {
    let mut calldata = ArrayTrait::new();
    calldata.append(aggregator_address.into());
    return declare('AggregatorConsumer').deploy(@calldata).unwrap();
}

#[test]
fn test_read_decimals() {
    let decimals = 16;
    let mock_aggregator_address = deploy_mock_aggregator(decimals);
    let consumer_address = deploy_consumer(mock_aggregator_address);
    let consumer_dispatcher = IAggregatorConsumerDispatcher { contract_address: consumer_address };
    assert(decimals == consumer_dispatcher.read_decimals(), 'Invalid decimals');
}

#[test]
fn test_read_latest_round() {
    // Deploys the mock aggregator
    let mock_aggregator_address = deploy_mock_aggregator(16);
    let mock_aggregator_dispatcher = IMockAggregatorDispatcher {
        contract_address: mock_aggregator_address
    };

    // Deploys the consumer
    let consumer_address = deploy_consumer(mock_aggregator_address);
    let consumer_dispatcher = IAggregatorConsumerDispatcher { contract_address: consumer_address };

    // No round data has been initialized, so reading the latest round should return no data
    let empty_latest_round = consumer_dispatcher.read_latest_round();
    assert(empty_latest_round.round_id == 0, 'round_id != 0');
    assert(empty_latest_round.answer == 0, 'answer != 0');
    assert(empty_latest_round.block_num == 0, 'block_num != 0');
    assert(empty_latest_round.started_at == 0, 'started_at != 0');
    assert(empty_latest_round.updated_at == 0, 'updated_at != 0');

    // Now let's set the latest round data to some random values
    let answer = 1;
    let block_num = 12345;
    let observation_timestamp = 100000;
    let transmission_timestamp = 200000;
    mock_aggregator_dispatcher
        .set_latest_round_data(answer, block_num, observation_timestamp, transmission_timestamp);

    // The latest round should now have some data
    let latest_round = consumer_dispatcher.read_latest_round();
    assert(latest_round.round_id == 1, 'round_id != 1');
    assert(latest_round.answer == answer, 'bad answer');
    assert(latest_round.block_num == block_num, 'bad block_num');
    assert(latest_round.started_at == observation_timestamp, 'bad started_at');
    assert(latest_round.updated_at == transmission_timestamp, 'bad updated_at');
}
