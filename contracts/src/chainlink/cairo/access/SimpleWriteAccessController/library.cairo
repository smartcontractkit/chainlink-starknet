%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.cairo_builtins import HashBuiltin

from chainlink.cairo.access.ownable import Ownable

@event
func AddedAccess(user: felt) {
}

@event
func RemovedAccess(user: felt) {
}

@event
func CheckAccessEnabled() {
}

@event
func CheckAccessDisabled() {
}

@storage_var
func SimpleWriteAccessController_check_enabled() -> (checkEnabled: felt) {
}

@storage_var
func SimpleWriteAccessController_access_list(address: felt) -> (bool: felt) {
}

// --- Ownership ---

@view
func owner{pedersen_ptr: HashBuiltin*, syscall_ptr: felt*, range_check_ptr}() -> (owner: felt) {
    let (owner) = Ownable.get_owner();
    return (owner=owner);
}

@view
func proposed_owner{pedersen_ptr: HashBuiltin*, syscall_ptr: felt*, range_check_ptr}() -> (
    proposed_owner: felt
) {
    let (proposed_owner) = Ownable.get_proposed_owner();
    return (proposed_owner=proposed_owner);
}

@external
func transfer_ownership{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    new_owner: felt
) -> () {
    return Ownable.transfer_ownership(new_owner);
}

@external
func accept_ownership{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() {
    return Ownable.accept_ownership();
}

// --- AC ---

// Adds an address to the access list
@external
func add_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(user: felt) {
    Ownable.assert_only_owner();
    let (has_access) = SimpleWriteAccessController_access_list.read(user);
    if (has_access == FALSE) {
        SimpleWriteAccessController_access_list.write(user, TRUE);
        AddedAccess.emit(user);
        return ();
    }

    return ();
}

// Removes an address from the access list
@external
func remove_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(user: felt) {
    Ownable.assert_only_owner();
    let (has_access) = SimpleWriteAccessController_access_list.read(user);
    if (has_access == TRUE) {
        SimpleWriteAccessController_access_list.write(user, FALSE);
        RemovedAccess.emit(user);
        return ();
    }

    return ();
}

// Makes the access check enforced
@external
func enable_access_check{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() {
    Ownable.assert_only_owner();
    let (check_enabled) = SimpleWriteAccessController_check_enabled.read();
    if (check_enabled == FALSE) {
        SimpleWriteAccessController_check_enabled.write(TRUE);
        CheckAccessEnabled.emit();
        return ();
    }

    return ();
}

// makes the access check unenforced
@external
func disable_access_check{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() {
    Ownable.assert_only_owner();
    let (check_enabled) = SimpleWriteAccessController_check_enabled.read();
    if (check_enabled == TRUE) {
        SimpleWriteAccessController_check_enabled.write(FALSE);
        CheckAccessDisabled.emit();
        return ();
    }

    return ();
}

namespace SimpleWriteAccessController {
    func initialize{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
        owner_address: felt
    ) {
        Ownable.initializer(owner_address);
        SimpleWriteAccessController_check_enabled.write(TRUE);

        return ();
    }

    func has_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
        user: felt, data_len: felt, data: felt*
    ) -> (bool: felt) {
        let (has_access) = SimpleWriteAccessController_access_list.read(user);
        if (has_access == TRUE) {
            return (TRUE,);
        }

        let (check_enabled) = SimpleWriteAccessController_check_enabled.read();
        if (check_enabled == FALSE) {
            return (TRUE,);
        }

        return (FALSE,);
    }

    func check_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(user: felt) {
        alloc_locals;

        let empty_data_len = 0;
        let (empty_data) = alloc();

        let (bool) = SimpleWriteAccessController.has_access(user, empty_data_len, empty_data);
        with_attr error_message("SimpleWriteAccessController: address does not have access") {
            assert bool = TRUE;
        }

        return ();
    }
}
