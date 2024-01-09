use starknet::ContractAddress;

#[starknet::contract]
mod AccessController {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::access_control::{AccessControl, IAccessController};
    use chainlink::libraries::ownable::{OwnableComponent, IOwnable};
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    component!(path: OwnableComponent, storage: ownable, event: OwnableEvent);

    #[abi(embed_v0)]
    impl OwnableImpl = OwnableComponent::OwnableImpl<ContractState>;
    impl InternalImpl = OwnableComponent::InternalImpl<ContractState>;

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        #[flat]
        OwnableEvent: OwnableComponent::Event,
    }

    #[storage]
    struct Storage {
        #[substorage(v0)]
        ownable: OwnableComponent::Storage,
    }

    #[constructor]
    fn constructor(ref self: ContractState, owner_address: ContractAddress) {
        self.ownable.initializer(owner_address);
        let mut access_control = AccessControl::unsafe_new_contract_state();
        AccessControl::constructor(ref access_control);
    }

    #[external(v0)]
    impl AccessControllerImpl of IAccessController<ContractState> {
        fn has_access(self: @ContractState, user: ContractAddress, data: Array<felt252>) -> bool {
            let state = AccessControl::unsafe_new_contract_state();
            AccessControl::has_access(@state, user, data)
        }

        fn add_access(ref self: ContractState, user: ContractAddress) {
            self.ownable.assert_only_owner();
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::add_access(ref state, user);
        }

        fn remove_access(ref self: ContractState, user: ContractAddress) {
            self.ownable.assert_only_owner();
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::remove_access(ref state, user);
        }

        fn enable_access_check(ref self: ContractState) {
            self.ownable.assert_only_owner();
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::enable_access_check(ref state);
        }

        fn disable_access_check(ref self: ContractState) {
            self.ownable.assert_only_owner();
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::disable_access_check(ref state);
        }
    }

    ///
    /// Upgradeable
    ///

    // #[view]
    fn type_and_version(self: @ContractState) -> felt252 {
        'AccessController 1.0.0'
    }

    #[external(v0)]
    impl UpgradeableImpl of IUpgradeable<ContractState> {
        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            self.ownable.assert_only_owner();
            Upgradeable::upgrade(new_impl);
        }
    }
}
