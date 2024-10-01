trait IERC677Receiver {
    fn on_token_transfer(sender: ContractAddress, value: u256, data: Array<felt252>);
}

trait IERC677 {
    fn transfer_and_call(to: ContractAddress, value: u256, data: Array<felt252>) -> bool;
}

