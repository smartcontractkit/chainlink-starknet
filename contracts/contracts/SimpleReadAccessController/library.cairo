%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from SimpleWriteAccessController.library import simple_write_access_controller

namespace simple_read_access_controller:
    func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        owner_address : felt
    ):
        simple_write_access_controller.constructor(owner_address)
        return ()
    end

    func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        user : felt, data_len : felt, data : felt*
    ) -> (bool : felt):
        let (bool) = simple_write_access_controller.has_access(user, data_len, data)
        return (bool)
    end

    func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
        # TODO: cal; has_access here
    end
end
