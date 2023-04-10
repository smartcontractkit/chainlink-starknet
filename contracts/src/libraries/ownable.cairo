use starknet::ContractAddress;

// todo augustus: whoever implements ownable must also expose external functions owner() and proposed_owner()
#[contract]
mod Ownable {
    use starknet::ContractAddress;
    use starknet::contract_address_const;
    use starknet::ContractAddressZeroable;
    use zeroable::Zeroable;

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
    fn OwnershipTransferRequested(from: starknet::ContractAddress, to: starknet::ContractAddress){}

    //
    // Constructor
    //

    #[constructor]
    fn constructor(owner: ContractAddress) {
      initializer(owner);
    }

    fn initializer(owner: ContractAddress) {
        assert(!owner.is_zero(), 'Ownable: transfer to 0');
        _accept_ownership_transfer(owner);
    }

    //
    // Modifiers
    //
    fn assert_only_owner() {
        let owner = _owner::read();
        let caller = starknet::get_caller_address();
        // todo augustus: verify i can remove this (caller is the zero address should not be possible anymore with introduction of fees)
        assert(!caller.is_zero(), 'Ownable: caller is 0');
        assert(caller == owner, 'Ownable: caller is not owner');
    }

    //
    // Getters
    //
    #[view]
    fn owner() -> ContractAddress {
        _owner::read()
    }

    #[view]
    fn proposed_owner() -> ContractAddress {
        _proposed_owner::read()
    }

    //
    // Setters
    //

    // todo augustus: add the check for transferring to self?
    #[external]
    fn transfer_ownership(new_owner: ContractAddress) {
        assert(!new_owner.is_zero(), 'Ownable: transfer to 0');
        assert_only_owner();

        _proposed_owner::write(new_owner);
        let previous_owner = _owner::read();
        OwnershipTransferRequested(previous_owner, new_owner);
    }

    #[external]
    fn accept_ownership() {
        let proposed_owner = _proposed_owner::read();
        let caller = starknet::get_caller_address();

        // todo augustus: verify this can be removed (  // caller cannot be zero address to avoid overwriting owner when proposed_owner is not set)
        assert(!caller.is_zero(), 'Ownable: caller is 0');
        assert(caller == proposed_owner, 'Ownable: not proposed_owner');

        _accept_ownership_transfer(proposed_owner);
    }

    // todo augustus: verify we don't need this (this isn't even defined in the solidity contracts for Simple[Read/Write]AccessController)
    #[external]
    fn renounce_ownership() {
        assert_only_owner();
        _accept_ownership_transfer(starknet::contract_address_const::<0>());
    }


    //
    // Internal
    //
    fn _accept_ownership_transfer(new_owner: starknet::ContractAddress) {
        let previous_owner = _owner::read();
        _owner::write(new_owner);
        _proposed_owner::write(starknet::contract_address_const::<0>());
        OwnershipTransferred(previous_owner, new_owner);
    }
}
