#[contract]
mod SimpleReadAccessController {
    use zeroable::Zeroable;

    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::access_controller::AccessController;
    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::upgradeable::Upgradeable;

    #[constructor]
    fn constructor(owner_address: ContractAddress) {
        Ownable::initializer(owner_address);
        AccessController::initializer();
    }

    #[view]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        AccessController::has_access(user, data)
    }

    ///
    /// Ownable
    ///

    #[view]
    fn owner() -> ContractAddress {
        Ownable::owner()
    }

    #[view]
    fn proposed_owner() -> ContractAddress {
        Ownable::proposed_owner()
    }

    #[external]
    fn transfer_ownership(new_owner: ContractAddress) {
        Ownable::transfer_ownership(new_owner);
    }

    #[external]
    fn accept_ownership() {
        Ownable::accept_ownership();
    }

    #[external]
    fn renounce_ownership() {
        Ownable::renounce_ownership();
    }

    ///
    /// Upgradeable
    ///

    #[view]
    fn type_and_version() -> felt252 {
        'ReadAccessController 1.0.0'
    }

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl);
    }
}
