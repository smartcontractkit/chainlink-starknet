#[starknet::contract]
mod MockMultisigTarget {
    use array::ArrayTrait;

    #[storage]
    struct Storage {}

    #[abi(per_item)]
    #[generate_trait]
    impl HelperImpl of HelperTrait {
        #[external(v0)]
        fn increment(ref self: ContractState, val1: felt252, val2: felt252) -> Array<felt252> {
            array![val1 + 1, val2 + 1]
        }
    }
}
