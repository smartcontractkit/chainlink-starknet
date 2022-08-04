// SPDX-License-Identifier: MIT

pragma solidity >0.6.0 <=0.8.0;

/// @title IStarknetCore - Sends a message to an L2 contract and consumes a message that was sent from an L2 contract.
interface IStarknetCore {
    /**
     * @notice Sends a message to toAddress.
     * @param toAddress the contract address on L2.
     * @param selector of the function with l1_handler.
     * @param payload the data to send.
     * @return hash of the message.
     */
    function sendMessageToL2(
        uint256 toAddress,
        uint256 selector,
        uint256[] calldata payload
    ) external returns (bytes32);

    /**
     * @notice Consumes a message that was sent from an L2 contract.
     * @param fromAddress the contract address on L2.
     * @param payload the data to consume.
     * @return hash of the message.
     */
    function consumeMessageFromL2(uint256 fromAddress, uint256[] calldata payload) external returns (bytes32);
}
