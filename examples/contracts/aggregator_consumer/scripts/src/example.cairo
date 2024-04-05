use sncast_std::{call, CallResult};

fn main() {
    let address = 0x775ee7f2f0b3e15953f4688f7a2ce5a0d1e7c8e18e5f929d461c037f14b690e
        .try_into()
        .unwrap();
    let result = call(address, selector!("description"), array![]);
    if result.is_err() {
        println!("{:?}", result.unwrap_err());
        panic_with_felt252('call failed');
    } else {
        println!("{:?}", result);
    }
}

