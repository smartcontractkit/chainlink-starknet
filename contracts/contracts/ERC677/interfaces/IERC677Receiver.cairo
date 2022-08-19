%lang starknet

from starkware.cairo.common.uint256 import Uint256

@contract_interface
namespace IERC677Receiver:
    func on_token_transfer(sender : felt, value : Uint256, data_len : felt, data : felt*):
    end
    func is_supports_interface(interface_id : felt) -> (success : felt):
    end
end
