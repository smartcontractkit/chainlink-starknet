// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./interfaces/IStarknetCore.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorValidatorInterface.sol";
import "@chainlink/contracts/src/v0.8/interfaces/TypeAndVersionInterface.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";
import "@chainlink/contracts/src/v0.8/SimpleWriteAccessController.sol";

import "@chainlink/contracts/src/v0.8/dev/vendor/openzeppelin-solidity/v4.3.1/contracts/utils/Address.sol";

/**
 * @title Validator - makes cross chain call to update the Sequencer Uptime Feed on L2
 */
contract Validator is TypeAndVersionInterface, AggregatorValidatorInterface, SimpleWriteAccessController {
  uint256 constant BRIDGE_MODE_DEPOSIT = 0;
  uint256 constant BRIDGE_MODE_WITHDRAW = 1;
  int256 private constant ANSWER_SEQ_OFFLINE = 1;
  uint256 constant UPDATE_STATUS_SELECTOR = 1456392953608713042542366145306621198614634764826083394033556249611221792745;

  IStarknetCore public immutable STARKNET_CORE;
  uint256 public L2_UPTIME_FEED_ADDR;

  /**
   * @param starknetCore the address of the StarknetCore contract address
   */
  constructor(
    address starknetCore,
    uint256 l2UptimeFeedAddr
  ) {
    require(starknetCore != address(0), "Invalid Starknet Core address");
    require(l2UptimeFeedAddr != 0, "Invalid StarknetSequencerUptimeFeed contract address");
    STARKNET_CORE = IStarknetCore(starknetCore);
    L2_UPTIME_FEED_ADDR = l2UptimeFeedAddr;
  }


  function toUInt256(bool x) internal pure returns (uint r) {
    assembly { r := x }
  }

  function addressToUint(address value)
        internal
        pure
        returns (uint256 convertedValue)
    {
      return uint256(uint160(address(value)));
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
    return "StarknetValidator 1.0.0";
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
    STARKNET_CORE.sendMessageToL2(
            L2_UPTIME_FEED_ADDR,
            UPDATE_STATUS_SELECTOR,
            payload
        );
    return true;
  }
}