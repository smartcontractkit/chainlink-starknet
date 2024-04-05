use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let consumer_address = 0x56e078ee90929f13f2ca83545c71b98136c99b22822ada66ad2aff9595439fc
        .try_into()
        .unwrap();

    let result = call(consumer_address, selector!("read_answer"), array![]);
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", result);
    }
}

