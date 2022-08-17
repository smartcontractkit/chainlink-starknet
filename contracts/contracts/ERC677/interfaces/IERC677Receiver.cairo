%lang starknet

from starkware.cairo.common.uint256 import Uint256

@contract_interface
namespace IERC677Receiver:
    func onTokenTransfer(selector : felt, data_len : felt, data : felt*):
    end
end
