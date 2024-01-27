use starknet::ContractAddress;

#[starknet::contract]
mod AccessController {
    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::access_control::{AccessControl, IAccessController};
    use chainlink::libraries::ownable::{Ownable, IOwnable};
    use chainlink::libraries::upgradeable::{Upgradeable, IUpgradeable};

    #[storage]
    struct Storage {}

    #[constructor]
    fn constructor(ref self: ContractState, owner_address: ContractAddress) {
        let mut ownable = Ownable::unsafe_new_contract_state();
        Ownable::constructor(ref ownable, owner_address);
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
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::add_access(ref state, user);
        }

        fn remove_access(ref self: ContractState, user: ContractAddress) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::remove_access(ref state, user);
        }

        fn enable_access_check(ref self: ContractState) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::enable_access_check(ref state);
        }

        fn disable_access_check(ref self: ContractState) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            let mut state = AccessControl::unsafe_new_contract_state();
            AccessControl::disable_access_check(ref state);
        }
    }

    ///
    /// Ownable
    ///

    #[external(v0)]
    impl OwnableImpl of IOwnable<ContractState> {
        fn owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::owner(@state)
        }

        fn proposed_owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::proposed_owner(@state)
        }

        fn transfer_ownership(ref self: ContractState, new_owner: ContractAddress) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::transfer_ownership(ref state, new_owner)
        }

        fn accept_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::accept_ownership(ref state)
        }

        fn renounce_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::renounce_ownership(ref state)
        }
    }

    ///
    /// Upgradeable
    ///

    #[view]
    fn type_and_version(self: @ContractState) -> felt252 {
        'AccessController 1.0.0'
    }

    #[external(v0)]
    impl UpgradeableImpl of IUpgradeable<ContractState> {
        fn upgrade(ref self: ContractState, new_impl: ClassHash) {
            let ownable = Ownable::unsafe_new_contract_state();
            Ownable::assert_only_owner(@ownable);
            Upgradeable::upgrade(new_impl);
        }
    }
}
