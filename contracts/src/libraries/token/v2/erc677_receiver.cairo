use starknet::ContractAddress;

const IERC677_RECEIVER_ID: felt252 =
    0x224f0246bc4ebdcc391196e93f522342f12393c1a456db2a57043638940254;

#[starknet::interface]
trait IERC677Receiver<TContractState> {
    fn on_token_transfer(
        ref self: TContractState, sender: ContractAddress, value: u256, data: Array<felt252>
    );
}

#[starknet::component]
mod ERC677ReceiverComponent {
    use openzeppelin::introspection::src5::SRC5Component::InternalTrait as SRC5InternalTrait;
    use openzeppelin::introspection::src5::SRC5Component;
    use starknet::ContractAddress;
    use super::{IERC677Receiver, IERC677_RECEIVER_ID};

    #[storage]
    struct Storage {}

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {}

    #[generate_trait]
    impl InternalImpl<
        TContractState,
        +HasComponent<TContractState>,
        // ensure that the contract implements the IERC677Receiver interface
        +IERC677Receiver<TContractState>,
        impl SRC5: SRC5Component::HasComponent<TContractState>,
        +Drop<TContractState>
    > of InternalTrait<TContractState> {
        /// Initializes the contract by registering the IERC677Receiver interface ID.
        /// This should be used inside the contract's constructor.
        fn initializer(ref self: ComponentState<TContractState>) {
            let mut src5_component = get_dep_component_mut!(ref self, SRC5);
            src5_component.register_interface(IERC677_RECEIVER_ID);
        }
    }
}
