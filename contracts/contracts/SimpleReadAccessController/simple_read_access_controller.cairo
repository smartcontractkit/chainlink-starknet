%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info
from starkware.cairo.common.bool import TRUE, FALSE

from ocr2.interfaces.IAccessController import IAccessController
from SimpleReadAccessController.library import simple_read_access_controller
from SimpleWriteAccessController.library import simple_write_access_controller

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner_address : felt
):
    simple_read_access_controller.constructor(owner_address)
    return ()
end

@view
func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    user : felt, data_len : felt, data : felt*
) -> (bool : felt):
    let (has_access) = simple_read_access_controller.has_access(user, data_len, data)
    return (has_access)
end

@view
func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    address : felt
):
    simple_write_access_controller.check_access(address)
    return ()
end
