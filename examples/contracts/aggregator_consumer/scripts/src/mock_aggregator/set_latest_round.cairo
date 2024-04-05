use sncast_std::{invoke, InvokeResult, get_nonce};

use starknet::ContractAddress;

fn main() {
    // If you are using testnet, this address may need to be changed
    // If you are using the local starknet-devnet-rs container, this can be left alone
    let mock_aggregator_address = 0x3c6f82da5dbfa89ec9dbe414f33d23d1720d15568e4a880afcc9b0c3d98d127
        .try_into()
        .unwrap();

    // Feel free to modify these 
    let answer = 1;
    let block_num = 12345;
    let observation_timestamp = 1711716556;
    let transmission_timestamp = 1711716514;

    let result = invoke(
        mock_aggregator_address,
        selector!("set_latest_round_data"),
        array![answer, block_num, observation_timestamp, transmission_timestamp],
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

