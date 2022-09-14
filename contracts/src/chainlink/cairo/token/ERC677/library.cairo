%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math import assert_not_zero
from starkware.cairo.common.uint256 import Uint256
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.bool import TRUE
from openzeppelin.token.erc20.library import ERC20
from chainlink.cairo.token.ERC677.IERC677Receiver import IERC677Receiver
from chainlink.cairo.token.ERC677.IERC677 import Transfer

const IERC677_RECEIVER_ID = 0x4f3dcd

# https://github.com/ethereum/EIPs/issues/677
namespace ERC677:
    # All non-account contracts that want to receive the `transferAndCall` need to implement ERC165 + ERC677, or revert.
    # All account contracts need to implement ERC165 (to be detected as an account, or revert), and optionally ERC677
    #   if they want to be triggered, if not they'll still receive funds but not the trigger.
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

        # TODO: should we always just return TRUE, for all cases?
        let (bool) = IERC677Receiver.supportsInterface(to, IERC677_RECEIVER_ID)
        if bool == TRUE:
            IERC677Receiver.onTokenTransfer(to, sender, value, data_len, data)
            return (TRUE)
        end
        return (TRUE)
    end
end
