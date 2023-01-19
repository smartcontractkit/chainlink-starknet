%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin

from chainlink.cairo.access.SimpleReadAccessController.library import SimpleReadAccessController
from chainlink.cairo.access.SimpleWriteAccessController.library import (
    owner,
    proposed_owner,
    transfer_ownership,
    accept_ownership,
    add_access,
    remove_access,
    enable_access_check,
    disable_access_check,
)

@constructor
func constructor{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    owner_address: felt
) {
    SimpleReadAccessController.initialize(owner_address);
    return ();
}

// implements IAccessController
@view
func has_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    user: felt, data_len: felt, data: felt*
) -> (bool: felt) {
    let (has_access) = SimpleReadAccessController.has_access(user, data_len, data);
    return (has_access,);
}

// implements IAccessController
@view
func check_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(user: felt) {
    SimpleReadAccessController.check_access(user);
    return ();
}
