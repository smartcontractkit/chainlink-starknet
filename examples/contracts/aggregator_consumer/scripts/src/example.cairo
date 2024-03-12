use sncast_std::{call, CallResult};

// The example below uses a contract deployed to the Goerli testnet
fn main() {
    let contract_address = 0x7ad10abd2cc24c2e066a2fee1e435cd5fa60a37f9268bfbaf2e98ce5ca3c436;
    let call_result = call(contract_address.try_into().unwrap(), 'get_greeting', array![]);
    assert(*call_result.data[0] == 'Hello, Starknet!', *call_result.data[0]);
    println!("{:?}", call_result);
}
