#[starknet::interface]
trait IMockNonUpgradeable<TContractState> {
    fn bar(self: @TContractState) -> bool;
}

#[starknet::contract]
mod MockNonUpgradeable {
    #[storage]
    struct Storage {}

    #[constructor]
    fn constructor(self: @ContractState) {}

    impl MockNonUpgradeableImpl of super::IMockNonUpgradeable<ContractState> {
        fn bar(self: @ContractState) -> bool {
            true
        }
    }
}