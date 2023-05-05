#[contract]
mod ValidReceiver {
    use starknet::ContractAddress;
    use array::ArrayTrait;


    struct Storage {
        _sender: ContractAddress, 
    }

    #[constructor]
    fn constructor() {}

    #[external]
    fn on_token_transfer(sender: ContractAddress, value: u256, data: Array<felt252>) {
        _sender::write(sender);
    }

    #[external]
    fn supports_interface(interface_id: u32) -> bool {
        true
    }

    #[view]
    fn verify() -> ContractAddress {
        _sender::read()
    }
}
