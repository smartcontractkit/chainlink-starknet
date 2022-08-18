%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.uint256 import Uint256
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.bool import TRUE
from contracts.ERC677.ERC20.ERC20_base import ERC20_transfer

from contracts.ERC677.interfaces.IERC677Receiver import IERC677Receiver

@event
func Transfer(from_ : felt, to : felt, value : Uint256, data_len : felt, data : felt*):
end

# PRIVATE

func contractFallback{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    to : felt, selector : felt, data_len : felt, data : felt*
):
    IERC677Receiver.onTokenTransfer(to, selector, data_len, data)
    return ()
end

namespace ERC677:
    func transfer_and_call{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        to : felt, value : Uint256, selector : felt, data_len : felt, data : felt*
    ) -> (success : felt):
        alloc_locals

        let (caller) = get_caller_address()
        with_attr error_message("ERC677: address can not be null"):
            assert_not_zero(to)
        end
        ERC20_transfer(caller, to, value)
        Transfer.emit(caller, to, value, data_len, data)

        contractFallback(to, selector, data_len, data)
        return (TRUE)
    end
end
