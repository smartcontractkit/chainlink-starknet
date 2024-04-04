use sncast_std::{invoke, InvokeResult, get_nonce};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let consumer_address = 0x622f81d884ed45b3874e038b38f7e1c6dcbe789dc96da5161c17d47d9c53570
        .try_into()
        .unwrap();

    let result = invoke(
        consumer_address,
        selector!("set_answer"),
        array![],
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

