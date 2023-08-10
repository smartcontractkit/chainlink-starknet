use starknet::ContractAddress;
#[starknet::interface]
trait IAccessController<TContractState> {
    fn has_access(self: @TContractState, user: ContractAddress, data: Array<felt252>) -> bool;
    fn add_access(ref self: TContractState, user: ContractAddress);
    fn remove_access(ref self: TContractState, user: ContractAddress);
    fn enable_access_check(ref self: TContractState);
    fn disable_access_check(ref self: TContractState);
}

#[starknet::contract]
mod AccessControl {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;
    use zeroable::Zeroable;

    #[storage]
    struct Storage {
        _check_enabled: bool,
        _access_list: LegacyMap<ContractAddress, bool>,
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        AddedAccess: AddedAccess,
        RemovedAccess: RemovedAccess,
        AccessControlEnabled: AccessControlEnabled,
        AccessControlDisabled: AccessControlDisabled,
    }

    #[derive(Drop, starknet::Event)]
    struct AddedAccess {
        user: ContractAddress
    }

    #[derive(Drop, starknet::Event)]
    struct RemovedAccess {
        user: ContractAddress
    }

    #[derive(Drop, starknet::Event)]
    struct AccessControlEnabled {}

    #[derive(Drop, starknet::Event)]
    struct AccessControlDisabled {}

    fn has_access(self: @ContractState, user: ContractAddress, data: Array<felt252>) -> bool {
        let has_access = self._access_list.read(user);
        if has_access {
            return true;
        }

        let check_enabled = self._check_enabled.read();
        if !check_enabled {
            return true;
        }

        false
    }

    fn has_read_access(self: @ContractState, user: ContractAddress, data: Array<felt252>) -> bool {
        let _has_access = has_access(self, user, data);
        if _has_access {
            return true;
        }

        // NOTICE: read access is granted to direct calls, to enable off-chain reads.
        if user.is_zero() {
            return true;
        }

        false
    }

    fn check_access(self: @ContractState, user: ContractAddress) {
        let allowed = has_access(self, user, ArrayTrait::new());
        assert(allowed, 'user does not have access');
    }

    fn check_read_access(self: @ContractState, user: ContractAddress) {
        let allowed = has_read_access(self, user, ArrayTrait::new());
        assert(allowed, 'user does not have read access');
    }

    //
    // Unprotected
    //

    #[constructor]
    fn constructor(ref self: ContractState) {
        self._check_enabled.write(true);
        self.emit(Event::AccessControlEnabled(AccessControlEnabled {}));
    }

    fn add_access(ref self: ContractState, user: ContractAddress) {
        let has_access = self._access_list.read(user);
        if !has_access {
            self._access_list.write(user, true);
            self.emit(Event::AddedAccess(AddedAccess { user: user }));
        }
    }

    fn remove_access(ref self: ContractState, user: ContractAddress) {
        let has_access = self._access_list.read(user);
        if has_access {
            self._access_list.write(user, false);
            self.emit(Event::RemovedAccess(RemovedAccess { user: user }));
        }
    }

    fn enable_access_check(ref self: ContractState) {
        let check_enabled = self._check_enabled.read();
        if !check_enabled {
            self._check_enabled.write(true);
            self.emit(Event::AccessControlEnabled(AccessControlEnabled {}));
        }
    }

    fn disable_access_check(ref self: ContractState) {
        let check_enabled = self._check_enabled.read();
        if check_enabled {
            self._check_enabled.write(false);
            self.emit(Event::AccessControlDisabled(AccessControlDisabled {}));
        }
    }
}
