use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let consumer_address = 0xa208894d6f5726cb9ba1d256bbd0ff9ff9eafa4fa187fd5f444fd2139f269a
        .try_into()
        .unwrap();

    let result = call(consumer_address, selector!("read_latest_round"), array![]);
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", result);
    }
}

