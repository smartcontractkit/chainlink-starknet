use starknet::ContractAddress;

#[starknet::interface]
trait IERC677Receiver<TContractState> {
    fn on_token_transfer(ref self: TContractState, sender: ContractAddress, value: u256, data: Array<felt252>);
    // implements EIP-165, where function selectors are defined by Ethereum ABI using the ethereum function signatures
    fn supports_interface(ref self: TContractState, interface_id: u32) -> bool;
}

mod ERC677 {
    use starknet::ContractAddress;
    use chainlink::libraries::token::erc20::ERC20;
    use array::ArrayTrait;
    use array::SpanTrait;
    use clone::Clone;
    use array::ArrayTCloneImpl;

    use super::IERC677ReceiverDispatcher;
    use super::IERC677ReceiverDispatcherTrait;

    // ethereum function selector of "onTokenTransfer(address,uint256,bytes)"
    const IERC677_RECEIVER_ID: u32 = 0xa4c0ed36_u32;

    #[event]
    fn Transfer(from: ContractAddress, to: ContractAddress, value: u256, data: Array<felt252>) {}

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
