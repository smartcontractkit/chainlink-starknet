use starknet::ContractAddress;

// TODO: consider replacing with OwnableTwoStep if https://github.com/OpenZeppelin/cairo-contracts/pull/809/ lands

#[starknet::interface]
trait IOwnable<TState> {
    fn owner(self: @TState) -> ContractAddress;
    fn proposed_owner(self: @TState) -> ContractAddress;
    fn transfer_ownership(ref self: TState, new_owner: ContractAddress);
    fn accept_ownership(ref self: TState);
    fn renounce_ownership(ref self: TState);
}

#[starknet::component]
mod OwnableComponent {
    use starknet::ContractAddress;
    use starknet::get_caller_address;

    #[storage]
    struct Storage {
        Ownable_owner: ContractAddress,
        Ownable_proposed_owner: ContractAddress
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        OwnershipTransferred: OwnershipTransferred,
        OwnershipTransferRequested: OwnershipTransferRequested
    }

    #[derive(Drop, starknet::Event)]
    struct OwnershipTransferred {
        #[key]
        previous_owner: ContractAddress,
        #[key]
        new_owner: ContractAddress,
    }

    #[derive(Drop, starknet::Event)]
    struct OwnershipTransferRequested {
        #[key]
        previous_owner: ContractAddress,
        #[key]
        new_owner: ContractAddress,
    }

    mod Errors {
        const NOT_OWNER: felt252 = 'Caller is not the owner';
        const NOT_PROPOSED_OWNER: felt252 = 'Caller is not proposed owner';
        const ZERO_ADDRESS_CALLER: felt252 = 'Caller is the zero address';
        const ZERO_ADDRESS_OWNER: felt252 = 'New owner is the zero address';
    }

    /// Adds support for two step ownership transfer.
    #[embeddable_as(OwnableImpl)]
    impl Ownable<
        TContractState, +HasComponent<TContractState>
    > of super::IOwnable<ComponentState<TContractState>> {
        /// Returns the address of the current owner.
        fn owner(self: @ComponentState<TContractState>) -> ContractAddress {
            self.Ownable_owner.read()
        }

        /// Returns the address of the pending owner.
        fn proposed_owner(self: @ComponentState<TContractState>) -> ContractAddress {
            self.Ownable_proposed_owner.read()
        }

        /// Finishes the two-step ownership transfer process by accepting the ownership.
        /// Can only be called by the pending owner.
        fn accept_ownership(ref self: ComponentState<TContractState>) {
            let caller = get_caller_address();
            let proposed_owner = self.Ownable_proposed_owner.read();
            assert(caller == proposed_owner, Errors::NOT_PROPOSED_OWNER);
            self._accept_ownership();
        }

        /// Starts the two-step ownership transfer process by setting the pending owner.
        fn transfer_ownership(
            ref self: ComponentState<TContractState>, new_owner: ContractAddress
        ) {
            assert(!new_owner.is_zero(), 'Ownable: transfer to 0');
            self.assert_only_owner();
            self._propose_owner(new_owner);
        }

        /// Leaves the contract without owner. It will not be possible to call `assert_only_owner`
        /// functions anymore. Can only be called by the current owner.
        fn renounce_ownership(ref self: ComponentState<TContractState>) {
            self.assert_only_owner();
            self._transfer_ownership(Zeroable::zero());
        }
    }

    #[generate_trait]
    impl InternalImpl<
        TContractState, +HasComponent<TContractState>
    > of InternalTrait<TContractState> {
        /// Sets the contract's initial owner.
        ///
        /// This function should be called at construction time.
        fn initializer(ref self: ComponentState<TContractState>, owner: ContractAddress) {
            self._transfer_ownership(owner);
        }

        /// Panics if called by any account other than the owner. Use this
        /// to restrict access to certain functions to the owner.
        fn assert_only_owner(self: @ComponentState<TContractState>) {
            let owner = self.Ownable_owner.read();
            let caller = get_caller_address();
            assert(!caller.is_zero(), Errors::ZERO_ADDRESS_CALLER);
            assert(caller == owner, Errors::NOT_OWNER);
        }

        /// Transfers ownership to the pending owner.
        ///
        /// Internal function without access restriction.
        fn _accept_ownership(ref self: ComponentState<TContractState>) {
            let proposed_owner = self.Ownable_proposed_owner.read();
            self.Ownable_proposed_owner.write(Zeroable::zero());
            self._transfer_ownership(proposed_owner);
        }

        /// Sets a new pending owner.
        ///
        /// Internal function without access restriction.
        fn _propose_owner(ref self: ComponentState<TContractState>, new_owner: ContractAddress) {
            let previous_owner = self.Ownable_owner.read();
            self.Ownable_proposed_owner.write(new_owner);
            self
                .emit(
                    OwnershipTransferRequested {
                        previous_owner: previous_owner, new_owner: new_owner
                    }
                );
        }

        /// Transfers ownership of the contract to a new address.
        ///
        /// Internal function without access restriction.
        ///
        /// Emits an `OwnershipTransferred` event.
        fn _transfer_ownership(
            ref self: ComponentState<TContractState>, new_owner: ContractAddress
        ) {
            let previous_owner: ContractAddress = self.Ownable_owner.read();
            self.Ownable_owner.write(new_owner);
            self
                .emit(
                    OwnershipTransferred { previous_owner: previous_owner, new_owner: new_owner }
                );
        }
    }
}
