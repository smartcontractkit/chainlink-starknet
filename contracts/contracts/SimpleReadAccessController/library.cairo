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
        let (tx_info) = get_tx_info()
        let (has_access) = simple_write_access_controller.has_access(user, data_len, data)

        if has_access == TRUE:
            return (TRUE)
        end

        if tx_info.account_contract_address == user:
            return (TRUE)
        end

        return (FALSE)
    end
end
