use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    let consumer_address = 0xa208894d6f5726cb9ba1d256bbd0ff9ff9eafa4fa187fd5f444fd2139f269a
        .try_into()
        .unwrap();

    let result = call(consumer_address, selector!("read_decimals"), array![]);
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", result);
    }
}

