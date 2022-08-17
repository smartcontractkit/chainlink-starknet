%lang starknet

from starkware.cairo.common.uint256 import Uint256

@contract_interface
namespace IERC721Receiver:
    func transferAndCall(to : felt, value : Uint256, selector : felt, data_len : felt, data : felt*) -> (
        success : felt
    ):
    end
end
