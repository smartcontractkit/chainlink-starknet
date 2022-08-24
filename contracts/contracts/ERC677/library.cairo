%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.uint256 import Uint256
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.bool import TRUE
from openzeppelin.token.erc20.library import ERC20
from contracts.ERC677.interfaces.IERC677Receiver import IERC677Receiver
from contracts.ERC677.interfaces.IERC677 import Transfer

const IERC677_RECEIVER_ID = 0x4f3dcd

namespace ERC677:
    func transfer_and_call{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        to : felt, value : Uint256, data_len : felt, data : felt*
    ) -> (success : felt):
        alloc_locals

        let (sender) = get_caller_address()
        with_attr error_message("ERC677: address can not be null"):
            assert_not_zero(to)
        end
        ERC20.transfer(to, value)
        Transfer.emit(sender, to, value, data_len, data)

        let (bool) = IERC677Receiver.supportsInterface(to, IERC677_RECEIVER_ID)
        if bool == TRUE:
            IERC677Receiver.onTokenTransfer(to, sender, value, data_len, data)
            return (TRUE)
        end
        return (TRUE)
    end
end
