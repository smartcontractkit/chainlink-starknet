#[contract]
mod SimpleWriteAccessController {
    use starknet::ContractAddress;
    use chainlink::libraries::access_controller::AccessController;

    struct Storage {
        _enabled: bool,
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

            let check_enabled = _enabled::read();
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

    #[external]
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool {
        SimpleWriteAccessController::has_access(user, data)
    }

    #[external]
    fn check_access(user: ContractAddress) {
        SimpleWriteAccessController::check_access(user)
    }

    #[external]
    fn add_access(user: ContractAddress) {
        // TODO: Ownable.assert_only_owner();
        let has_access = _access_list::read(user);
        if !has_access {
            _access_list::write(user, true);
            AddedAccess(user);
        }
    }

    #[external]
    fn remove_access(user: ContractAddress) {
        // TODO: Ownable.assert_only_owner();
        let has_access = _access_list::read(user);
        if has_access {
            _access_list::write(user, false);
            RemovedAccess(user);
        }
    }

    #[external]
    fn enable_access_check() {
        // TODO: Ownable.assert_only_owner();
        let check_enabled = _enabled::read();
        if !check_enabled {
            _enabled::write(true);
            CheckAccessEnabled();
        }
    }

    #[external]
    fn disable_access_check() {
        // TODO: Ownable.assert_only_owner();
        let check_enabled = _enabled::read();
        if check_enabled {
            _enabled::write(false);
            CheckAccessDisabled();
        }
    }
}
