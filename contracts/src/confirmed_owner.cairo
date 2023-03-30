#[contract]
mod ConfirmedOwner {
    use starknet::ContractAddressIntoFelt252;

    use poc::libraries::ownable;

    #[constructor]
    fn constructor(newOwner: starknet::ContractAddress) {
        Ownable::initialize(newOwner);
    }

    #[view]
    fn owner() -> starknet::ContractAddress {
        Ownable::getOwner();
    }

    #[view]
    fn proposedOwner() -> starknet::ContractAddress {
        Ownable::proposedOwner();
    }

    #[external]
    fn transferOwnership(proposedOwner: starknet::ContractAddress) {
        Ownable::transferOwnership(proposedOwner);
    }

    #[external]
    fn acceptOwnership() {
        Ownable::acceptOwnership();
    }
}