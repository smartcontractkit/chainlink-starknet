use starknet::ContractAddress;

#[starknet::interface]
trait IOwnable<TContractState> {
    fn owner(self: @TContractState) -> ContractAddress;
    fn proposed_owner(self: @TContractState) -> ContractAddress;
    fn transfer_ownership(ref self: TContractState, new_owner: ContractAddress);
    fn accept_ownership(ref self: TContractState);
    fn renounce_ownership(ref self: TContractState);
}

#[starknet::contract]
mod Ownable {
    use starknet::ContractAddress;
    use starknet::contract_address_const;
    use zeroable::Zeroable;

    #[storage]
    struct Storage {
        _owner: ContractAddress,
        _proposed_owner: ContractAddress,
    }

    //
    // Events
    //

    #[event]
    fn OwnershipTransferred(previous_owner: ContractAddress, newOwner: ContractAddress) {}

    #[event]
    fn OwnershipTransferRequested(from: starknet::ContractAddress, to: starknet::ContractAddress) {}

   //
    // Constructor
    //

    #[constructor]
    fn constructor(ref self: ContractState, owner: ContractAddress) {
        assert(!owner.is_zero(), 'Ownable: transfer to 0');
        self._accept_ownership_transfer(owner);
    }

    //
    // Modifiers
    //
    fn assert_only_owner(self: @ContractState) {
        let owner = self._owner.read();
        let caller = starknet::get_caller_address();
        assert(caller == owner, 'Ownable: caller is not owner');
    }

    #[external(v0)]
    impl OwnableImpl of super::IOwnable<ContractState> {
        //
        // Getters
        //
        fn owner(self: @ContractState) -> ContractAddress {
            self._owner.read()
        }

        fn proposed_owner(self: @ContractState) -> ContractAddress {
            self._proposed_owner.read()
        }

        //
        // Setters
        //

        fn transfer_ownership(ref self: ContractState, new_owner: ContractAddress) {
            assert(!new_owner.is_zero(), 'Ownable: transfer to 0');
            assert_only_owner(@self);

            self._proposed_owner.write(new_owner);
            let previous_owner = self._owner.read();
            OwnershipTransferRequested(previous_owner, new_owner);
        }

        fn accept_ownership(ref self: ContractState) {
            let proposed_owner = self._proposed_owner.read();
            let caller = starknet::get_caller_address();

            assert(caller == proposed_owner, 'Ownable: not proposed_owner');

            self._accept_ownership_transfer(proposed_owner);
        }

        fn renounce_ownership(ref self: ContractState) {
            assert_only_owner(@self);
            self._accept_ownership_transfer(starknet::contract_address_const::<0>());
        }
    }


    //
    // Internal
    //

    #[generate_trait]
    impl InternalImpl of InternalTrait {
        fn _accept_ownership_transfer(ref self: ContractState, new_owner: starknet::ContractAddress) {
            let previous_owner = self._owner.read();
            self._owner.write(new_owner);
            self._proposed_owner.write(starknet::contract_address_const::<0>());
            OwnershipTransferred(previous_owner, new_owner);
        }
    }
}
