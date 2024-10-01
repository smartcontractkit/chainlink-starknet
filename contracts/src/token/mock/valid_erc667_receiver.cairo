use starknet::ContractAddress;
#[starknet::interface]
trait MockValidReceiver<TContractState> {
    fn verify(self: @TContractState) -> ContractAddress;
}

#[starknet::contract]
mod ValidReceiver {
    use starknet::ContractAddress;
    use array::ArrayTrait;
    use openzeppelin::introspection::src5::SRC5Component;
    use chainlink::libraries::token::v2::erc677_receiver::{
        ERC677ReceiverComponent, IERC677Receiver
    };

    component!(path: SRC5Component, storage: src5, event: SRC5Event);
    component!(path: ERC677ReceiverComponent, storage: erc677_receiver, event: ERC677ReceiverEvent);

    // SRC5
    #[abi(embed_v0)]
    impl SRC5Impl = SRC5Component::SRC5Impl<ContractState>;

    // ERC677Receiver
    impl SRC5InternalImpl = ERC677ReceiverComponent::InternalImpl<ContractState>;

    #[storage]
    struct Storage {
        _sender: ContractAddress,
        #[substorage(v0)]
        src5: SRC5Component::Storage,
        #[substorage(v0)]
        erc677_receiver: ERC677ReceiverComponent::Storage
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        SRC5Event: SRC5Component::Event,
        #[flat]
        ERC677ReceiverEvent: ERC677ReceiverComponent::Event
    }

    #[constructor]
    fn constructor(ref self: ContractState) {
        self.erc677_receiver.initializer();
    }

    #[abi(embed_v0)]
    impl ERC677ReceiverImpl of IERC677Receiver<ContractState> {
        fn on_token_transfer(
            ref self: ContractState, sender: ContractAddress, value: u256, data: Array<felt252>
        ) {
            self._sender.write(sender);
        }
    }

    #[abi(embed_v0)]
    impl ValidReceiver of super::MockValidReceiver<ContractState> {
        fn verify(self: @ContractState) -> ContractAddress {
            self._sender.read()
        }
    }
}
