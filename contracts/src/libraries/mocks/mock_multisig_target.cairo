use array::ArrayTrait;

#[starknet::interface]
trait IMockMultisigTarget<TContractState> {
    fn increment(ref self: TContractState, val1: felt252, val2: felt252) -> Array<felt252>;
    fn set_value(ref self: TContractState, value: felt252);
    fn flip_toggle(ref self: TContractState);
    fn read(self: @TContractState) -> (felt252, bool);
}

#[starknet::contract]
mod MockMultisigTarget {
    use array::ArrayTrait;
    use super::IMockMultisigTarget;

    #[storage]
    struct Storage {
        value: felt252,
        toggle: bool
    }

    #[abi(embed_v0)]
    impl MockMultisigTargetImpl of super::IMockMultisigTarget<ContractState> {
        fn increment(ref self: ContractState, val1: felt252, val2: felt252) -> Array<felt252> {
            array![val1 + 1, val2 + 1]
        }

        fn set_value(ref self: ContractState, value: felt252) {
            self.value.write(value);
        }

        fn flip_toggle(ref self: ContractState) {
            self.toggle.write(!self.toggle.read());
        }

        fn read(self: @ContractState) -> (felt252, bool) {
            (self.value.read(), self.toggle.read())
        }
    }
}
