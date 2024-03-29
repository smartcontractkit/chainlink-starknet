use sncast_std::{invoke, InvokeResult, get_nonce};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let mock_aggregator_address = 0x376b1abf788737bded2011a0f76ce61cabdeaec22e97b8a4e231b149dd49fc0
        .try_into()
        .unwrap();

    // Feel free to modify these 
    let answer = 1;
    let block_num = 12345;
    let observation_timestamp = 100000;
    let transmission_timestamp = 200000;

    let max_fee = 99999999999999999;
    let result = invoke(
        mock_aggregator_address,
        selector!("set_latest_round_data"),
        array![answer, block_num, observation_timestamp, transmission_timestamp],
        Option::Some(max_fee),
        Option::Some(get_nonce('pending'))
    );

    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('invoke failed');
    } else {
        println!("{:?}", result);
    }
}

