// SPDX-License-Identifier: MIT
pragma solidity >0.6.0 <=0.8.0;

import '@chainlink/contracts/src/v0.8/interfaces/AggregatorValidatorInterface.sol';
import '@chainlink/contracts/src/v0.8/interfaces/TypeAndVersionInterface.sol';
import '@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol';
import '@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol';
import '@chainlink/contracts/src/v0.8/SimpleWriteAccessController.sol';

import '@chainlink/contracts/src/v0.8/dev/vendor/openzeppelin-solidity/v4.3.1/contracts/utils/Address.sol';

import './interfaces/IStarknetCore.sol';
import './utils/utils.sol';

/**
 * @title Validator - makes cross chain call to update the Sequencer Uptime Feed on L2
 */
contract Validator is TypeAndVersionInterface, AggregatorValidatorInterface, SimpleWriteAccessController {
    int256 private constant ANSWER_SEQ_OFFLINE = 1;
    /* Selector hardcoded because StarkNet generates selectors differently than the standard ethereum way
    different hash function on stark curve is used */
    uint256 constant STARK_SELECTOR_UPDATE_STATUS =
        1585322027166395525705364165097050997465692350398750944680096081848180365267;

    IStarknetCore public immutable STARKNET_CORE;
    uint256 public L2_UPTIME_FEED_ADDR;

    /**
     * @param starknetCore the address of the StarknetCore contract address
     * @param l2UptimeFeedAddr the address of the Sequencer Uptime Feed on L2
     */
    constructor(address starknetCore, uint256 l2UptimeFeedAddr) {
        require(starknetCore != address(0), 'Invalid Starknet Core address');
        require(l2UptimeFeedAddr != 0, 'Invalid StarknetSequencerUptimeFeed contract address');
        STARKNET_CORE = IStarknetCore(starknetCore);
        L2_UPTIME_FEED_ADDR = l2UptimeFeedAddr;
    }

    /**
     * @notice versions:
     *
     * - Validator 0.1.0: initial release
     * - Validator 1.0.0: change target of L2 sequencer status update
     *   - now calls `updateStatus` on an L2 SequencerUptimeFeed contract instead of
     *     directly calling the Flags contract
     *
     * @inheritdoc TypeAndVersionInterface
     */
    function typeAndVersion() external pure virtual override returns (string memory) {
        return 'StarknetValidator 0.1.0';
    }

    /**
     * @notice validate method sends an xDomain L2 tx to update Uptime Feed contract on L2.
     * @dev A message is sent using the L1CrossDomainMessenger. This method is accessed controlled.
     * @param previousAnswer previous aggregator answer
     * @param currentAnswer new aggregator answer - value of 1 considers the sequencer offline.
     */
    function validate(
        uint256, /* previousRoundId */
        int256 previousAnswer,
        uint256, /* currentRoundId */
        int256 currentAnswer
    ) external override checkAccess returns (bool) {
        bool status = currentAnswer == ANSWER_SEQ_OFFLINE;
        uint64 timestamp = uint64(block.timestamp);
        uint256[] memory payload = new uint256[](2);

        // File payload with `status` and `timestamp`
        payload[0] = toUInt256(status);
        payload[1] = timestamp;
        // Make the starknet cross chain call
        STARKNET_CORE.sendMessageToL2(L2_UPTIME_FEED_ADDR, STARK_SELECTOR_UPDATE_STATUS, payload);
        return true;
    }
}
