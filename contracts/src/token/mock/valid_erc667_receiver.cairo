#[contract]
mod ValidReceiver {
    use starknet::ContractAddress;
    use array::ArrayTrait;


    struct Storage {
        _sender: ContractAddress,
        _value: u256,
    }

    #[constructor]
    fn constructor() {}

    #[external]
    fn on_token_transfer(sender: ContractAddress, value: u256, data: Array<felt252>) {
        _sender::write(sender);
        _value::write(value);
    }

    #[external]
    fn supports_interface(interface_id: u32) -> bool {
        true
    }
}
