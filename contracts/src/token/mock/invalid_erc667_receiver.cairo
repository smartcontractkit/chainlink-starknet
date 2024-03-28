#[starknet::contract]
mod InvalidReceiver {
    #[storage]
    struct Storage {
        _supports: bool
    }

    #[constructor]
    fn constructor(ref self: ContractState) {}

    #[abi(per_item)]
    #[generate_trait]
    impl HelperImpl of HelperTrait {
        // toggle whether or not receiver says it supports the interface id
        #[external(v0)]
        fn set_supports(ref self: ContractState, support: bool) {
            self._supports.write(support);
        }

        #[external(v0)]
        fn supports_interface(self: @ContractState, interface_id: u32) -> bool {
            self._supports.read()
        }
    }
}
