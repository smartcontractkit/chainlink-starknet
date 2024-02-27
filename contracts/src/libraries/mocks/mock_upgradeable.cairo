use starknet::class_hash::ClassHash;

#[starknet::interface]
trait IMockUpgradeable<TContractState> {
    fn foo(self: @TContractState) -> bool;
    fn upgrade(ref self: TContractState, new_impl: ClassHash);
}

#[starknet::contract]
mod MockUpgradeable {
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::upgradeable::Upgradeable;

    #[storage]
    struct Storage {}

    #[constructor]
    fn constructor(ref self: ContractState) {}

    #[abi(embed_v0)]
    impl MockUpgradeableImpl of super::IMockUpgradeable<ContractState> {
        fn foo(self: @ContractState) -> bool {
            true
        }

        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            Upgradeable::upgrade(new_impl)
        }
    }
}
