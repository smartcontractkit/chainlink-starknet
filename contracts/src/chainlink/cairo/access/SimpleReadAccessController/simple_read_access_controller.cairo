%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin

from chainlink.cairo.access.SimpleReadAccessController.library import simple_read_access_controller

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner_address : felt
):
    simple_read_access_controller.initialize(owner_address)
    return ()
end

# implements IAccessController
@view
func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    user : felt, data_len : felt, data : felt*
) -> (bool : felt):
    let (has_access) = simple_read_access_controller.has_access(user, data_len, data)
    return (has_access)
end

# implements IAccessController
@view
func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    simple_read_access_controller.check_access(user)
    return ()
end
