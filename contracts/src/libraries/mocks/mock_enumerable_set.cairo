#[starknet::interface]
trait IMockEnumerableSet<TContractState> {
    fn add(ref self: TContractState, set_id: u256, value: u256) -> bool;
    fn remove(ref self: TContractState, set_id: u256, target_value: u256) -> bool;
    fn contains(self: @TContractState, set_id: u256, value: u256) -> bool;
    fn length(self: @TContractState, set_id: u256) -> u256;
    fn at(self: @TContractState, set_id: u256, index: u256) -> u256;
    fn values(self: @TContractState, set_id: u256) -> Array<u256>;
}

#[starknet::contract]
mod MockEnumerableSet {
    use chainlink::libraries::enumerable_set::EnumerableSetComponent;

    component!(path: EnumerableSetComponent, storage: set, event: EnumerableSetEvent);

    // EnumerableSet
    impl EnumerableSetInternalImpl = EnumerableSetComponent::InternalImpl<ContractState>;

    #[storage]
    struct Storage {
        #[substorage(v0)]
        set: EnumerableSetComponent::Storage,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        EnumerableSetEvent: EnumerableSetComponent::Event,
    }

    #[abi(embed_v0)]
    impl MockEnumerableSetImpl of super::IMockEnumerableSet<ContractState> {
        fn add(ref self: ContractState, set_id: u256, value: u256) -> bool {
            self.set.add(set_id, value)
        }
        fn remove(ref self: ContractState, set_id: u256, target_value: u256) -> bool {
            self.set.remove(set_id, target_value)
        }
        fn contains(self: @ContractState, set_id: u256, value: u256) -> bool {
            self.set.contains(set_id, value)
        }
        fn length(self: @ContractState, set_id: u256) -> u256 {
            self.set.length(set_id)
        }
        fn at(self: @ContractState, set_id: u256, index: u256) -> u256 {
            self.set.at(set_id, index)
        }
        fn values(self: @ContractState, set_id: u256) -> Array<u256> {
            self.set.values(set_id)
        }
    }
}
