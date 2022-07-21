// SPDX-License-Identifier: Apache-2.0.
pragma solidity ^0.8.0;

import './interfaces/IStarknetMessaging.sol';
import './library/NamedStorage.sol';

/**
* @title StarknetMessaging implements sending messages to L2 by adding them to a pipe and consuming messages from L2 by
  removing them from a different pipe. A deriving contract can handle the former pipe and add items
  to the latter pipe while interacting with L2..
*/
contract StarknetMessaging is IStarknetMessaging {
    /// Random slot storage elements and accessors.
    string constant L1L2_MESSAGE_MAP_TAG = 'STARKNET_1.0_MSGING_L1TOL2_MAPPPING_V2';
    string constant L2L1_MESSAGE_MAP_TAG = 'STARKNET_1.0_MSGING_L2TOL1_MAPPPING';

    string constant L1L2_MESSAGE_NONCE_TAG = 'STARKNET_1.0_MSGING_L1TOL2_NONCE';

    function addressToUint(address value) internal pure returns (uint256 convertedValue) {
        convertedValue = uint256(uint160(value));
    }

    function l1ToL2Messages(bytes32 msgHash) external view returns (uint256) {
        return l1ToL2Messages()[msgHash];
    }

    function l2ToL1Messages(bytes32 msgHash) external view returns (uint256) {
        return l2ToL1Messages()[msgHash];
    }

    function l1ToL2Messages() internal pure returns (mapping(bytes32 => uint256) storage) {
        return NamedStorage.bytes32ToUint256Mapping(L1L2_MESSAGE_MAP_TAG);
    }

    function l2ToL1Messages() internal pure returns (mapping(bytes32 => uint256) storage) {
        return NamedStorage.bytes32ToUint256Mapping(L2L1_MESSAGE_MAP_TAG);
    }

    function l1ToL2MessageNonce() public view returns (uint256) {
        return NamedStorage.getUintValue(L1L2_MESSAGE_NONCE_TAG);
    }

    /**
     * @notice sends of a message from L1 to L2.
     * @param to_address the contract address on L2.
     * @param selector of the function with l1_handler.
     * @param payload the data to send.
     * @return hash of the message.
     */
    function sendMessageToL2(
        uint256 to_address,
        uint256 selector,
        uint256[] calldata payload
    ) external override returns (bytes32) {
        uint256 nonce = l1ToL2MessageNonce();
        NamedStorage.setUintValue(L1L2_MESSAGE_NONCE_TAG, nonce + 1);
        emit LogMessageToL2(msg.sender, to_address, selector, payload, nonce);
        bytes32 msgHash = keccak256(
            abi.encodePacked(addressToUint(msg.sender), to_address, nonce, selector, payload.length, payload)
        );
        l1ToL2Messages()[msgHash] += 1;

        return msgHash;
    }

    /**
     * @notice Consumes a message that was sent from an L2 contract.
     * @param from_address the contract address on L2.
     * @param payload the data to consume.
     * @return hash of the message.
     */
    function consumeMessageFromL2(uint256 from_address, uint256[] calldata payload)
        external
        override
        returns (bytes32)
    {
        bytes32 msgHash = keccak256(abi.encodePacked(from_address, addressToUint(msg.sender), payload.length, payload));

        require(l2ToL1Messages()[msgHash] > 0, 'INVALID_MESSAGE_TO_CONSUME');
        emit ConsumedMessageToL1(from_address, msg.sender, payload);
        l2ToL1Messages()[msgHash] -= 1;
        return msgHash;
    }
}
