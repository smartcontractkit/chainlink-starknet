use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let aggregator_address = 0x376b1abf788737bded2011a0f76ce61cabdeaec22e97b8a4e231b149dd49fc0
        .try_into()
        .unwrap();

    let result = call(aggregator_address, selector!("latest_round_data"), array![]);
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", result);
    }
}

