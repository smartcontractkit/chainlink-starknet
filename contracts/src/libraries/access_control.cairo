use starknet::ContractAddress;
#[starknet::interface]
trait IAccessController<TContractState> {
    fn has_access(self: @TContractState, user: ContractAddress, data: Array<felt252>) -> bool;
    fn has_read_access(self: @TContractState, user: ContractAddress, data: Array<felt252>) -> bool;
    fn add_access(ref self: TContractState, user: ContractAddress);
    fn remove_access(ref self: TContractState, user: ContractAddress);
    fn enable_access_check(ref self: TContractState);
    fn disable_access_check(ref self: TContractState);
}

// Requires Ownable subcomponent.
#[starknet::component]
mod AccessControlComponent {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;
    use zeroable::Zeroable;

    use openzeppelin::access::ownable::OwnableComponent;

    use OwnableComponent::InternalImpl as OwnableInternalImpl;

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
        #[key]
        user: ContractAddress
    }

    #[derive(Drop, starknet::Event)]
    struct RemovedAccess {
        #[key]
        user: ContractAddress
    }

    #[derive(Drop, starknet::Event)]
    struct AccessControlEnabled {}

    #[derive(Drop, starknet::Event)]
    struct AccessControlDisabled {}

    #[embeddable_as(AccessControlImpl)]
    impl AccessControl<
        TContractState,
        +HasComponent<TContractState>,
        impl Ownable: OwnableComponent::HasComponent<TContractState>,
        +Drop<TContractState>,
    > of super::IAccessController<ComponentState<TContractState>> {
        fn has_access(
            self: @ComponentState<TContractState>, user: ContractAddress, data: Array<felt252>
        ) -> bool {
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

        fn has_read_access(
            self: @ComponentState<TContractState>, user: ContractAddress, data: Array<felt252>
        ) -> bool {
            let _has_access = self.has_access(user, data);
            if _has_access {
                return true;
            }

            // NOTICE: read access is granted to direct calls, to enable off-chain reads.
            if user.is_zero() {
                return true;
            }

            false
        }

        fn add_access(ref self: ComponentState<TContractState>, user: ContractAddress) {
            get_dep_component!(@self, Ownable).assert_only_owner();
            let has_access = self._access_list.read(user);
            if !has_access {
                self._access_list.write(user, true);
                self.emit(Event::AddedAccess(AddedAccess { user: user }));
            }
        }

        fn remove_access(ref self: ComponentState<TContractState>, user: ContractAddress) {
            get_dep_component!(@self, Ownable).assert_only_owner();
            let has_access = self._access_list.read(user);
            if has_access {
                self._access_list.write(user, false);
                self.emit(Event::RemovedAccess(RemovedAccess { user: user }));
            }
        }

        fn enable_access_check(ref self: ComponentState<TContractState>) {
            get_dep_component!(@self, Ownable).assert_only_owner();
            let check_enabled = self._check_enabled.read();
            if !check_enabled {
                self._check_enabled.write(true);
                self.emit(Event::AccessControlEnabled(AccessControlEnabled {}));
            }
        }

        fn disable_access_check(ref self: ComponentState<TContractState>) {
            get_dep_component!(@self, Ownable).assert_only_owner();
            let check_enabled = self._check_enabled.read();
            if check_enabled {
                self._check_enabled.write(false);
                self.emit(Event::AccessControlDisabled(AccessControlDisabled {}));
            }
        }
    }

    #[generate_trait]
    impl InternalImpl<
        TContractState,
        +HasComponent<TContractState>,
        impl Ownable: OwnableComponent::HasComponent<TContractState>,
        +Drop<TContractState>,
    > of InternalTrait<TContractState> {
        fn initializer(ref self: ComponentState<TContractState>) {
            self._check_enabled.write(true);
            self.emit(Event::AccessControlEnabled(AccessControlEnabled {}));
        }

        fn check_access(self: @ComponentState<TContractState>, user: ContractAddress) {
            let allowed = AccessControl::has_access(self, user, ArrayTrait::new());
            assert(allowed, 'user does not have access');
        }

        fn check_read_access(self: @ComponentState<TContractState>, user: ContractAddress) {
            let allowed = AccessControl::has_read_access(self, user, ArrayTrait::new());
            assert(allowed, 'user does not have read access');
        }
    }
}
