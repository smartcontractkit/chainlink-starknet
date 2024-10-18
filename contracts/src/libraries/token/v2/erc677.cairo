use starknet::ContractAddress;

const IERC677_ID: felt252 = 0x3c4538abc63e0cdf912cef3d2e1389d0b2c3f24ee0c06b21736229f52ece6c8;

#[starknet::interface]
trait IERC677<TContractState> {
    fn transfer_and_call(
        ref self: TContractState, to: ContractAddress, value: u256, data: Array<felt252>
    ) -> bool;
}

#[starknet::component]
mod ERC677Component {
    use starknet::ContractAddress;
    use openzeppelin::token::erc20::interface::IERC20;
    use openzeppelin::introspection::interface::{ISRC5, ISRC5Dispatcher, ISRC5DispatcherTrait};
    use array::ArrayTrait;
    use array::SpanTrait;
    use clone::Clone;
    use array::ArrayTCloneImpl;
    use chainlink::libraries::token::v2::erc677_receiver::{
        IERC677ReceiverDispatcher, IERC677ReceiverDispatcherTrait, IERC677_RECEIVER_ID
    };

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

            let receiver = ISRC5Dispatcher { contract_address: to };

            let supports = receiver.supports_interface(IERC677_RECEIVER_ID);

            if supports {
                IERC677ReceiverDispatcher { contract_address: to }
                    .on_token_transfer(sender, value, data);
            }
            true
        }
    }
}
