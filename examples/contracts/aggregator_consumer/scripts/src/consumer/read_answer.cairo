use sncast_std::{call, CallResult};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let consumer_address = 0x622f81d884ed45b3874e038b38f7e1c6dcbe789dc96da5161c17d47d9c53570
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

