#[contract]
mod AccessController {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::access_control::AccessControl;
    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::upgradeable::Upgradeable;

    #[constructor]
    fn constructor(owner_address: ContractAddress) {
        Ownable::initializer(owner_address);
        AccessControl::initializer();
    }

    #[view]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        AccessControl::has_access(user, data)
    }

    #[external]
    fn add_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        AccessControl::add_access(user);
    }

    #[external]
    fn remove_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        AccessControl::remove_access(user);
    }

    #[external]
    fn enable_access_check() {
        Ownable::assert_only_owner();
        AccessControl::enable_access_check();
    }

    #[external]
    fn disable_access_check() {
        Ownable::assert_only_owner();
        AccessControl::disable_access_check();
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
        'WriteAccessController 1.0.0'
    }

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl);
    }
}
