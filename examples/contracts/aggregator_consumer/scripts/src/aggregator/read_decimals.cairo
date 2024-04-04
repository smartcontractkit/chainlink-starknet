use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let aggregator_address = 0x3c6f82da5dbfa89ec9dbe414f33d23d1720d15568e4a880afcc9b0c3d98d127
        .try_into()
        .unwrap();

    let result = call(aggregator_address, selector!("read_decimals"), array![]);
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", result);
    }
}

