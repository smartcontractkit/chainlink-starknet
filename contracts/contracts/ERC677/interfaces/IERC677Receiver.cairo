%lang starknet

from starkware.cairo.common.uint256 import Uint256

@contract_interface
namespace IERC677Receiver:
    func onTokenTransfer(sender : felt, value : Uint256, data_len : felt, data : felt*):
    end
    func supportsInterface(interface_id : felt) -> (success : felt):
    end
end
