use starknet::ContractAddress;
#[starknet::interface]
trait MockValidReceiver<TContractState> {
    fn verify(self: @TContractState) -> ContractAddress;
}

#[starknet::contract]
mod ValidReceiver {
    use starknet::ContractAddress;
    use array::ArrayTrait;
    use chainlink::libraries::token::erc677::IERC677Receiver;

    #[storage]
    struct Storage {
        _sender: ContractAddress,
    }

    #[constructor]
    fn constructor(ref self: ContractState) {}


    #[abi(embed_v0)]
    impl ERC677Receiver of IERC677Receiver<ContractState> {
        fn on_token_transfer(
            ref self: ContractState, sender: ContractAddress, value: u256, data: Array<felt252>
        ) {
            self._sender.write(sender);
        }

        fn supports_interface(ref self: ContractState, interface_id: u32) -> bool {
            true
        }
    }

    #[abi(embed_v0)]
    impl ValidReceiver of super::MockValidReceiver<ContractState> {
        fn verify(self: @ContractState) -> ContractAddress {
            self._sender.read()
        }
    }
}
