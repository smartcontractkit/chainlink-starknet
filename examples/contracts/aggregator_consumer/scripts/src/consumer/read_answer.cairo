use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let consumer_address = 0x5b12015734ce4bc3c72f9ae4d87ed80e2a28497b21e220a702d1e20e854b0cd
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

