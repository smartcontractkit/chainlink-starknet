use starknet::ContractAddress;

use chainlink::ocr2::mocks::mock_aggregator::IMockAggregatorDispatcherTrait;
use chainlink::ocr2::mocks::mock_aggregator::IMockAggregatorDispatcher;
use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;
use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;

use aggregator_consumer::ocr2::consumer::IAggregatorConsumerDispatcherTrait;
use aggregator_consumer::ocr2::consumer::IAggregatorConsumerDispatcher;

use snforge_std::{declare, ContractClassTrait, DeclareResultTrait};


fn deploy_mock_aggregator(decimals: u8) -> ContractAddress {
    let mut calldata = ArrayTrait::new();
    calldata.append(decimals.into());

    let contract = declare("MockAggregator").unwrap().contract_class();

    let (contract_address, _) = contract.deploy(@calldata).unwrap();

    contract_address
}

fn deploy_consumer(aggregator_address: ContractAddress) -> ContractAddress {
    let mut calldata = ArrayTrait::new();
    calldata.append(aggregator_address.into());

    let contract = declare("AggregatorConsumer").unwrap().contract_class();

    let (contract_address, _) = contract.deploy(@calldata).unwrap();

    contract_address
}

#[test]
fn test_read_decimals() {
    // Deploys the mock aggregator
    let decimals = 16;
    let mock_aggregator_address = deploy_mock_aggregator(decimals);
    let aggregator_dispatcher = IAggregatorDispatcher { contract_address: mock_aggregator_address };

    // Let's make sure the constructor arguments were passed in correctly
    assert(decimals == aggregator_dispatcher.decimals(), 'Invalid decimals');
}
#[test]
fn test_set_and_read_latest_round() {
    // Deploys the mock aggregator
    let mock_aggregator_address = deploy_mock_aggregator(16);
    let mock_aggregator_dispatcher = IMockAggregatorDispatcher {
        contract_address: mock_aggregator_address
    };
    let aggregator_dispatcher = IAggregatorDispatcher { contract_address: mock_aggregator_address };

    // No round data has been initialized, so reading the latest round should return no data
    let empty_latest_round = aggregator_dispatcher.latest_round_data();
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
    let latest_round = aggregator_dispatcher.latest_round_data();
    assert(latest_round.round_id == 1, 'round_id != 1');
    assert(latest_round.answer == answer, 'bad answer');
    assert(latest_round.block_num == block_num, 'bad block_num');
    assert(latest_round.started_at == observation_timestamp, 'bad started_at');
    assert(latest_round.updated_at == transmission_timestamp, 'bad updated_at');
}

#[test]
fn test_set_and_read_answer() {
    // Deploys the mock aggregator
    let mock_aggregator_address = deploy_mock_aggregator(16);
    let mock_aggregator_dispatcher = IMockAggregatorDispatcher {
        contract_address: mock_aggregator_address
    };

    // Deploys the consumer
    let consumer_address = deploy_consumer(mock_aggregator_address);
    let consumer_dispatcher = IAggregatorConsumerDispatcher { contract_address: consumer_address };

    // Let's make sure the AggregatorConsumer was initialized correctly
    assert(
        consumer_dispatcher.read_ocr_address() == mock_aggregator_address, 'Invalid OCR address'
    );
    assert(consumer_dispatcher.read_answer() == 0, 'Invalid initial answer');

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

    // The consumer should be able to query the aggregator for the new latest round data
    let latest_round = consumer_dispatcher.read_latest_round();
    assert(latest_round.round_id == 1, 'round_id != 1');
    assert(latest_round.answer == answer, 'bad answer');
    assert(latest_round.block_num == block_num, 'bad block_num');
    assert(latest_round.started_at == observation_timestamp, 'bad started_at');
    assert(latest_round.updated_at == transmission_timestamp, 'bad updated_at');

    // Now let's test that we can set the answer 
    consumer_dispatcher.set_answer(latest_round.answer);
    assert(answer == consumer_dispatcher.read_answer(), 'Invalid answer');
}

