#[contract]
mod SimpleWriteAccessController {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::access_controller::AccessController;
    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::upgradeable::Upgradeable;

    struct Storage {
        _check_enabled: bool,
        _access_list: LegacyMap<ContractAddress, bool>,
    }

    #[event]
    fn AddedAccess(user: ContractAddress) {}

    #[event]
    fn RemovedAccess(user: ContractAddress) {}

    #[event]
    fn CheckAccessEnabled() {}

    #[event]
    fn CheckAccessDisabled() {}

    impl SimpleWriteAccessController of AccessController {
        fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
            let has_access = _access_list::read(user);
            if has_access {
                return true;
            }

            let check_enabled = _check_enabled::read();
            if !check_enabled {
                return true;
            }

            false
        }

        fn check_access(user: ContractAddress) {
            let allowed = SimpleWriteAccessController::has_access(user, ArrayTrait::new());
            assert(allowed, 'address does not have access');
        }
    }

    #[constructor]
    fn constructor(owner_address: ContractAddress) {
        initializer(owner_address);
    }

    #[view]
    fn type_and_version() -> felt252 {
        'WriteAccessController 1.0.0'
    }

    #[external]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        SimpleWriteAccessController::has_access(user, data)
    }

    #[external]
    fn check_access(user: ContractAddress) {
        SimpleWriteAccessController::check_access(user);
    }

    #[external]
    fn add_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        let has_access = _access_list::read(user);
        if !has_access {
            _access_list::write(user, true);
            AddedAccess(user);
        }
    }

    #[external]
    fn remove_access(user: ContractAddress) {
        Ownable::assert_only_owner();
        let has_access = _access_list::read(user);
        if has_access {
            _access_list::write(user, false);
            RemovedAccess(user);
        }
    }

    #[external]
    fn enable_access_check() {
        Ownable::assert_only_owner();
        let check_enabled = _check_enabled::read();
        if !check_enabled {
            _check_enabled::write(true);
            CheckAccessEnabled();
        }
    }

    #[external]
    fn disable_access_check() {
        Ownable::assert_only_owner();
        let check_enabled = _check_enabled::read();
        if check_enabled {
            _check_enabled::write(false);
            CheckAccessDisabled();
        }
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
        Ownable::initializer(owner_address);
        _check_enabled::write(true);
    }
}
