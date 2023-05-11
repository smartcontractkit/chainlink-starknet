#[contract]
mod SimpleReadAccessController {
    use zeroable::Zeroable;

    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::access_controller::AccessController;
    use chainlink::libraries::simple_write_access_controller::SimpleWriteAccessController;
    use chainlink::libraries::upgradeable::Upgradeable;

    #[constructor]
    fn constructor(owner_address: ContractAddress) {
        SimpleWriteAccessController::initializer(owner_address)
    }

    impl SimpleReadAccessController of AccessController {
        fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
            let has_access = SimpleWriteAccessController::has_access(user, data);
            if has_access {
                return true;
            }

            // NOTICE: access is granted to direct calls, to enable off-chain reads.
            if user.is_zero() {
                return true;
            }

            false
        }

        fn check_access(user: ContractAddress) {
            let allowed = SimpleReadAccessController::has_access(user, ArrayTrait::new());
            assert(allowed, 'address does not have access');
        }
    }

    #[view]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        SimpleReadAccessController::has_access(user, data)
    }

    #[external]
    fn check_access(user: ContractAddress) {
        SimpleReadAccessController::check_access(user)
    }

    ///
    /// Upgradeable
    ///

    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl)
    }

    ///
    /// Internals
    ///

    fn initializer(owner_address: ContractAddress) {
        SimpleWriteAccessController::initializer(owner_address);
    }
}
