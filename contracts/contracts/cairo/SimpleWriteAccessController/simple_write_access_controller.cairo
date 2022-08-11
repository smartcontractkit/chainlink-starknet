%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.cairo.common.bool import TRUE, FALSE

from cairo.SimpleWriteAccessController.library import (
    s_access_list,
    s_check_enabled,
    AddedAccess,
    RemovedAccess,
    CheckAccessEnabled,
    CheckAccessDisabled,
    simple_write_access_controller,
)
from cairo.ocr2.interfaces.IAccessController import IAccessController

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner_address : felt
):
    simple_write_access_controller.constructor(owner_address)
    return ()
end

# implements IAccessController
@view
func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    user : felt, data_len : felt, data : felt*
) -> (bool : felt):
    let (has_access) = simple_write_access_controller.has_access(user, data_len, data)
    return (has_access)
end

# implements IAccessController
@view
func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    simple_write_access_controller.check_access(user)
    return ()
end
