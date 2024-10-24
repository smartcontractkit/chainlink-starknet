use starknet::ContractAddress;

#[starknet::interface]
trait IERC677<TContractState> {
    fn transfer_and_call(
        ref self: TContractState, to: ContractAddress, value: u256, data: Array<felt252>
    ) -> bool;
}

#[starknet::interface]
trait IERC677Receiver<TContractState> {
    fn on_token_transfer(
        ref self: TContractState, sender: ContractAddress, value: u256, data: Array<felt252>
    );
    // implements EIP-165, where function selectors are defined by Ethereum ABI using the ethereum function signatures
    fn supports_interface(ref self: TContractState, interface_id: u32) -> bool;
}

#[starknet::component]
mod ERC677Component {
    use starknet::ContractAddress;
    use openzeppelin::token::erc20::interface::IERC20;
    use array::ArrayTrait;
    use array::SpanTrait;
    use clone::Clone;
    use array::ArrayTCloneImpl;

    use super::IERC677ReceiverDispatcher;
    use super::IERC677ReceiverDispatcherTrait;

    // ethereum function selector of "onTokenTransfer(address,uint256,bytes)"
    const IERC677_RECEIVER_ID: u32 = 0xa4c0ed36_u32;

    #[storage]
    struct Storage {}

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        TransferAndCall: TransferAndCall,
    }

    #[derive(Drop, starknet::Event)]
    struct TransferAndCall {
        #[key]
        from: ContractAddress,
        #[key]
        to: ContractAddress,
        value: u256,
        data: Array<felt252>
    }

    #[embeddable_as(ERC677Impl)]
    impl ERC677<
        TContractState,
        +HasComponent<TContractState>,
        +IERC20<TContractState>,
        +Drop<TContractState>,
    > of super::IERC677<ComponentState<TContractState>> {
        fn transfer_and_call(
            ref self: ComponentState<TContractState>,
            to: ContractAddress,
            value: u256,
            data: Array<felt252>
        ) -> bool {
            let sender = starknet::info::get_caller_address();

            let mut contract = self.get_contract_mut();
            contract.transfer(to, value);
            self
                .emit(
                    Event::TransferAndCall(
                        TransferAndCall { from: sender, to: to, value: value, data: data.clone(), }
                    )
                );

            let receiver = IERC677ReceiverDispatcher { contract_address: to };

            let supports = receiver.supports_interface(IERC677_RECEIVER_ID);
            if supports {
                receiver.on_token_transfer(sender, value, data);
            }
            true
        }
    }
}
