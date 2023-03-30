use starknet::ContractAddress;

// todo augustus: whoever implements ownable must also expose external functions owner() and proposedOwner()
mod Ownable {
    use starknet::ContractAddress;
    use starknet::contract_address_const;
    use starknet::ContractAddressZeroable;
    use zeroable::Zeroable;

    //
    // Events
    //

    #[event]
    fn OwnershipTransferred(previousOwner: ContractAddress, newOwner: ContractAddress) {}

    #[event]
    fn OwnershipTransferRequested(from: starknet::ContractAddress, to: starknet::ContractAddress){}

    //
    // Constructor
    //

    fn initializer(owner: ContractAddress) {
        assert(!owner.is_zero(), 'Ownable: transfer to 0');
        _acceptOwnershipTransfer(owner);
    }

    // 
    // Modifiers
    // 

    fn assertOnlyOwner() {
        let owner = _owner::read();
        let caller = starknet::get_caller_address();
        // todo augustus: verify i can remove this (caller is the zero address should not be possible anymore with introduction of fees)
        assert(!caller.is_zero(), 'Ownable: caller is 0');

        assert(caller == owner, 'Ownable: caller is not owner');
    }

    //
    // Getters 
    //

    fn getOwner() -> ContractAddress {
        _owner::read()
    }

    fn getProposedOwner() -> ContractAddress {
        _proposedOwner::read()
    }

    // 
    // Setters
    // 

    // todo augustus: add the check for transferring to self?
    fn transferOwnership(newOwner: ContractAddress) {
        assertOnlyOwner();

        assert(!newOwner.is_zero(), 'Ownable: transfer to 0');
        _proposedOwner::write(newOwner);
        let previousOwner = _owner::read();
        OwnershipTransferRequested(previousOwner, newOwner);
    } 

    fn acceptOwnership() {
        let proposedOwner = _proposedOwner::read();
        let caller = starknet::get_caller_address();

        // todo augustus: verify this can be removed (  // caller cannot be zero address to avoid overwriting owner when proposed_owner is not set)
        assert(!caller.is_zero(), 'Ownable: caller is 0');
        assert(caller == proposedOwner, 'Ownable: not proposedOwner');

        _acceptOwnershipTransfer(proposedOwner);
    }

    // todo augustus: verify we don't need this (this isn't even defined in the solidity contracts for Simple[Read/Write]AccessController)
    fn renounceOwnership() {
        assertOnlyOwner();

        _acceptOwnershipTransfer(starknet::contract_address_const::<0>());
    }


    //
    // Internal
    //

    fn _acceptOwnershipTransfer(newOwner: starknet::ContractAddress) {
        let previousOwner = _owner::read();
        _owner::write(newOwner);
        _proposedOwner::write(starknet::contract_address_const::<0>());
        OwnershipTransferred(previousOwner, newOwner);
    }

    // STORAGE
    mod _owner {
        use starknet::SyscallResultTrait;

        fn read() -> starknet::ContractAddress {
            starknet::StorageAccess::<starknet::ContractAddress>::read(0_u32, address()).unwrap_syscall()
        }

        fn write(value: starknet::ContractAddress) {
            starknet::StorageAccess::<starknet::ContractAddress>::write(0_u32, address(), value).unwrap_syscall()
        }

        fn address() -> starknet::StorageBaseAddress {
            // "Ownable::owner" selector
            starknet::storage_base_address_const::<0x3db50198d2471ec1c5b126cf42805578fd6ddbfbfe01821f502e48da5e2e2f>()
        }
    }

    mod _proposedOwner {
        use starknet::SyscallResultTrait;

        fn read() -> starknet::ContractAddress {
            starknet::StorageAccess::<starknet::ContractAddress>::read(0_u32, address()).unwrap_syscall()
        }

        fn write(value: starknet::ContractAddress) {
            starknet::StorageAccess::<starknet::ContractAddress>::write(0_u32, address(), value).unwrap_syscall()
        }

        fn address() -> starknet::StorageBaseAddress {
            // "Ownable:proposedOwner" selector
            starknet::storage_base_address_const::<0x6b72f27cab291f26d21b31a72bade175e06809199d60b4f69c6d8e36803158>()
        }
    }

}
