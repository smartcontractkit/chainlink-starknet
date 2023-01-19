%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin

from chainlink.cairo.access.SimpleWriteAccessController.library import SimpleWriteAccessController

@constructor
func constructor{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    owner_address: felt
) {
    SimpleWriteAccessController.initialize(owner_address);
    return ();
}

// implements IAccessController
@view
func has_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    user: felt, data_len: felt, data: felt*
) -> (bool: felt) {
    let (has_access) = SimpleWriteAccessController.has_access(user, data_len, data);
    return (has_access,);
}

// implements IAccessController
@view
func check_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(user: felt) {
    SimpleWriteAccessController.check_access(user);
    return ();
}
