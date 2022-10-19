%lang starknet

from starkware.cairo.common.uint256 import Uint256

@event
func Transfer(from_: felt, to: felt, value: Uint256, data_len: felt, data: felt*) {
}

@contract_interface
namespace IERC677 {
    func transferAndCall(to: felt, value: Uint256, data_len: felt, data: felt*) -> (success: felt) {
    }
}
