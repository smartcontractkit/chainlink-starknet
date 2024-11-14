#[starknet::interface]
pub trait IStorage<TContractState> {
    fn store(ref self: TContractState, value: u256);
    fn retrieve(self: @TContractState) -> u256;
}

#[starknet::contract]
mod Storage {
    #[storage]
    struct Storage {
        value: u256
    }

    #[constructor]
    fn constructor(ref self: ContractState) {}

    #[abi(embed_v0)]
    impl StorageImpl of super::IStorage<ContractState> {
        fn store(ref self: ContractState, value: u256) {
            self.value.write(value)
        }

        fn retrieve(self: @ContractState) -> u256 {
            self.value.read()
        }
    }
}
