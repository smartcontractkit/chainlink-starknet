// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

/**
 * @notice toUInt256 convert a bool to uint256.
 */
function toUInt256(bool x) pure returns (uint256 r) {
    assembly {
        r := x
    }
}

/**
 * @notice addressToUint convert an address to uint256.
 */
function addressToUint(address value) pure returns (uint256 convertedValue) {
    return uint256(uint160(address(value)));
}
