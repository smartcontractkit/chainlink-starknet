use sncast_std::{invoke, InvokeResult, call, CallResult, get_nonce};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let consumer_address = 0x56e078ee90929f13f2ca83545c71b98136c99b22822ada66ad2aff9595439fc
        .try_into()
        .unwrap();

    // Reads the aggregator address from the AggregatorConsumer
    let read_ocr_address = call(consumer_address, selector!("read_ocr_address"), array![]);
    if read_ocr_address.is_err() {
        println!("{:?}", read_ocr_address.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", read_ocr_address);
    }

    // Queries the aggregator for the latest round data
    let mut read_ocr_address_data = read_ocr_address.unwrap().data.span();
    let aggregator_address = Serde::<
        starknet::ContractAddress
    >::deserialize(ref read_ocr_address_data)
        .unwrap();
    let latest_round = call(aggregator_address, selector!("latest_round_data"), array![]);
    if latest_round.is_err() {
        println!("{:?}", latest_round.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", latest_round);
    }

    // Uses the latest round data to set a new answer on the AggregatorConsumer
    let mut latest_round_data = latest_round.unwrap().data.span();
    let round = Serde::<chainlink::ocr2::aggregator::Round>::deserialize(ref latest_round_data)
        .unwrap();
    let result = invoke(
        consumer_address,
        selector!("set_answer"),
        array![round.answer.into()],
        Option::None,
        Option::Some(get_nonce('pending'))
    );
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('invoke failed');
    } else {
        println!("{:?}", result);
    }
}

