#[starknet::contract]
mod MockMultisigTarget {
    use array::ArrayTrait;

    #[storage]
    struct Storage {
        value: felt252,
        toggle: bool
    }

    #[abi(per_item)]
    #[generate_trait]
    impl HelperImpl of HelperTrait {
        #[external(v0)]
        fn increment(ref self: ContractState, val1: felt252, val2: felt252) -> Array<felt252> {
            array![val1 + 1, val2 + 1]
        }

        #[external(v0)]
        fn set_value(ref self: ContractState, value: felt252) {
            self.value.write(value);
        }

        #[external(v0)]
        fn flip_toggle(ref self: ContractState) {
            self.toggle.write(!self.toggle.read());
        }
    }
}
