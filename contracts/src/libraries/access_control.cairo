#[contract]
mod AccessControl {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;
    use zeroable::Zeroable;

    struct Storage {
        _check_enabled: bool,
        _access_list: LegacyMap<ContractAddress, bool>,
    }

    #[event]
    fn AddedAccess(user: ContractAddress) {}

    #[event]
    fn RemovedAccess(user: ContractAddress) {}

    #[event]
    fn AccessControlEnabled() {}

    #[event]
    fn AccessControlDisabled() {}

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

    fn has_read_access(user: ContractAddress, data: Array<felt252>) -> bool {
        let _has_access = has_access(user, data);
        if _has_access {
            return true;
        }

        // NOTICE: read access is granted to direct calls, to enable off-chain reads.
        if user.is_zero() {
            return true;
        }

        false
    }

    fn check_access(user: ContractAddress) {
        let allowed = has_access(user, ArrayTrait::new());
        assert(allowed, 'user does not have access');
    }

    fn check_read_access(user: ContractAddress) {
        let allowed = has_read_access(user, ArrayTrait::new());
        assert(allowed, 'user does not have read access');
    }

    //
    // Unprotected
    //

    fn initializer() {
        _check_enabled::write(true);
        AccessControlEnabled();
    }

    fn add_access(user: ContractAddress) {
        let has_access = _access_list::read(user);
        if !has_access {
            _access_list::write(user, true);
            AddedAccess(user);
        }
    }

    fn remove_access(user: ContractAddress) {
        let has_access = _access_list::read(user);
        if has_access {
            _access_list::write(user, false);
            RemovedAccess(user);
        }
    }

    fn enable_access_check() {
        let check_enabled = _check_enabled::read();
        if !check_enabled {
            _check_enabled::write(true);
            AccessControlEnabled();
        }
    }

    fn disable_access_check() {
        let check_enabled = _check_enabled::read();
        if check_enabled {
            _check_enabled::write(false);
            AccessControlDisabled();
        }
    }
}
