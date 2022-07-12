%lang starknet

from SimpleWriteAccessController.library import simple_write_access_controller

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.starknet.common.syscalls import get_tx_info
from starkware.cairo.common.bool import TRUE, FALSE

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

    func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
        # TODO: cal; has_access here
    end
end
