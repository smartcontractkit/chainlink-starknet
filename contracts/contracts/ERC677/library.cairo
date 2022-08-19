%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.uint256 import Uint256
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.bool import TRUE
from contracts.starkware.starknet.std_contracts.ERC20.ERC20_base import ERC20_transfer
from contracts.ERC677.interfaces.IERC677Receiver import IERC677Receiver
from contracts.ERC677.interfaces.IERC677 import Transfer

const IERC677_RECEIVER_ID = 0x4f3dcd

# PRIVATE

func contract_fallback{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    to : felt, sender : felt, value : Uint256, data_len : felt, data : felt*
):
    IERC677Receiver.on_token_transfer(to, sender, value, data_len, data)
    return ()
end

func is_contract{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(to : felt) -> (
    bool : felt
):
    let (bool) = IERC677Receiver.supportsInterface(to, IERC677_RECEIVER_ID)
    return (bool)
end

namespace ERC677:
    func transfer_and_call{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        to : felt, value : Uint256, data_len : felt, data : felt*
    ) -> (success : felt):
        alloc_locals

        let (sender) = get_caller_address()
        with_attr error_message("ERC677: address can not be null"):
            assert_not_zero(to)
        end
        ERC20_transfer(sender, to, value)
        Transfer.emit(sender, to, value, data_len, data)

        let (bool) = is_contract(to)
        if bool == 1:
            contract_fallback(to, sender, value, data_len, data)
            return (TRUE)
        end
        return (TRUE)
    end
end
