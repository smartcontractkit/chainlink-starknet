// SPDX-License-Identifier: Apache-2.0.
pragma solidity >0.6.0 <=0.8.0;

import '../../../vendor/starkgate-contracts-solidity-v0.8/contracts/starkware/starknet/solidity/StarknetMessaging.sol';

/**
 * @title MockStarknetMessaging make cross chain call.
 */
contract MockStarknetMessaging is StarknetMessaging {
    /**
     * @notice Mocks a message from L2 to L1.
     * @param from_address the contract address on L2.
     * @param to_address the contract address on L1.
     * @param payload the data to send.
     */
    function mockSendMessageFromL2(
        uint256 from_address,
        uint256 to_address,
        uint256[] calldata payload
    ) external {
        bytes32 msgHash = keccak256(abi.encodePacked(from_address, to_address, payload.length, payload));
        l2ToL1Messages()[msgHash] += 1;
    }

    /**
     * @notice Mocks consumption of a message from L1 to L2.
     * @param from_address the contract address on L1.
     * @param to_address the contract address on L2.
     * @param selector of the function with l1_handler.
     * @param payload the data to consume.
     * @param nonce the message nonce.
     */
    function mockConsumeMessageToL2(
        uint256 from_address,
        uint256 to_address,
        uint256 selector,
        uint256[] calldata payload,
        uint256 nonce
    ) external {
        bytes32 msgHash = keccak256(
            abi.encodePacked(from_address, to_address, nonce, selector, payload.length, payload)
        );

        require(l1ToL2Messages()[msgHash] > 0, 'INVALID_MESSAGE_TO_CONSUME');
        l1ToL2Messages()[msgHash] -= 1;
    }
}
