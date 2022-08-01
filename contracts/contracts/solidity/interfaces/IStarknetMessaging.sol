// SPDX-License-Identifier: Apache-2.0.
pragma solidity ^0.8.0;

/// @title IStarknetMessaging - Sends a message from L1 to L2 and from L2 to L1 and consumes it.
interface IStarknetMessaging {
    /**
     * @notice emitted when a message is sent from L2 to L1.
     * @param from_address the L2 address.
     * @param to_address the L1 address.
     * @param payload the data received.
     */
    event LogMessageToL1(uint256 indexed from_address, address indexed to_address, uint256[] payload);

    /**
     * @notice An event that is raised when a message is sent from L1 to L2.
     * @param from_address the L1 address.
     * @param to_address the L2 address.
     * @param selector of the function with l1_handler.
     * @param payload the data to send.
     * @param nonce the message nonce.
     */
    event LogMessageToL2(
        address indexed from_address,
        uint256 indexed to_address,
        uint256 indexed selector,
        uint256[] payload,
        uint256 nonce
    );

    /**
     * @notice An event that is raised when a message from L2 to L1 is consumed.
     * @param from_address the L2 address.
     * @param to_address the L1 address.
     * @param payload the data received.
     */
    event ConsumedMessageToL1(uint256 indexed from_address, address indexed to_address, uint256[] payload);

    /**
     * @notice An event that is raised when a message from L1 to L2 is consumed.
     * @param from_address the L1 address.
     * @param to_address the L2 address.
     * @param selector of the function with l1_handler.
     * @param payload the data to send.
     * @param nonce the message nonce.
     */
    event ConsumedMessageToL2(
        address indexed from_address,
        uint256 indexed to_address,
        uint256 indexed selector,
        uint256[] payload,
        uint256 nonce
    );

    /**
     * @notice Sends a message to an L2 contract.
     * @param to_address the L2 address.
     * @param selector of the function with l1_handler.
     * @param payload the data to send.
     * @return the hash of the message.
     */
    function sendMessageToL2(
        uint256 to_address,
        uint256 selector,
        uint256[] calldata payload
    ) external returns (bytes32);

    /**
     * @notice Consumes a message that was sent from an L2 contract.
     * @param fromAddress the L2 address.
     * @param payload the data to send.
     * @return the hash of the message.
     */
    function consumeMessageFromL2(uint256 fromAddress, uint256[] calldata payload) external returns (bytes32);
}
