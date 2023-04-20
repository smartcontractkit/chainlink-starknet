use starknet::ContractAddress;

#[abi]
trait IERC677Receiver {
    fn on_token_transfer(sender: ContractAddress, value: u256, data: Array<felt252>);
    // implements EIP-165, where function selectors are defined by Ethereum ABI using the ethereum function signatures
    fn supports_interface(interface_id: felt252) -> bool;
}

#[contract]
mod ERC677 {
    use starknet::ContractAddress;
    use chainlink::libraries::token::erc20::ERC20;
    use array::ArrayTrait;
    use array::SpanTrait;
    use clone::Clone;
    use array::ArrayTCloneImpl;

    use super::IERC677ReceiverDispatcher;
    use super::IERC677ReceiverDispatcherTrait;

    const IERC677_RECEIVER_ID: felt252 = 0xa53f2491;

    #[event]
    fn Transfer(from: ContractAddress, to: ContractAddress, value: u256, data: Array<felt252>) {}

    #[external]
    fn transfer_and_call(to: ContractAddress, value: u256, data: Array<felt252>) -> bool {
        let sender = starknet::info::get_caller_address();

        ERC20::transfer(to, value);
        Transfer(sender, to, value, data.clone());

        let receiver = IERC677ReceiverDispatcher { contract_address: to };

        let supports = receiver.supports_interface(IERC677_RECEIVER_ID);
        if supports {
            receiver.on_token_transfer(sender, value, data);
        }
        true
    }
}
