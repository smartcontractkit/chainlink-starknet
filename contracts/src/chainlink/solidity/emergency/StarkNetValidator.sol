// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@chainlink/contracts/src/v0.8/interfaces/AggregatorValidatorInterface.sol";
import "@chainlink/contracts/src/v0.8/interfaces/TypeAndVersionInterface.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AccessControllerInterface.sol";
import "@chainlink/contracts/src/v0.8/interfaces/AggregatorV3Interface.sol";
import "@chainlink/contracts/src/v0.8/SimpleWriteAccessController.sol";
import "@chainlink/contracts/src/v0.8/dev/vendor/openzeppelin-solidity/v4.3.1/contracts/utils/Address.sol";

import "../../../../vendor/starkware-libs/starkgate-contracts-solidity-v0.8/src/starkware/starknet/solidity/IStarknetMessaging.sol";

/// @title StarkNetValidator - makes cross chain calls to update the Sequencer Uptime Feed on L2
contract StarkNetValidator is TypeAndVersionInterface, AggregatorValidatorInterface, SimpleWriteAccessController {
  // Config for L1 -> L2 message cost approximation
  struct GasConfig {
    uint256 gasEstimate;
    address gasPriceL1Feed;
  }

  int256 private constant ANSWER_SEQ_OFFLINE = 1;

  uint256 public immutable SELECTOR_STARK_UPDATE_STATUS = _selectorStarkNet("update_status");
  uint256 public immutable L2_UPTIME_FEED_ADDR;

  IStarknetMessaging public immutable STARKNET_CROSS_DOMAIN_MESSENGER;

  AggregatorV3Interface private s_source;
  AccessControllerInterface private s_configAC;
  GasConfig private s_gasConfig;

  /// @notice This event is emitted when the gas config is set.
  event GasConfigSet(uint256 gasEstimate, address indexed gasPriceL1Feed);
  /// @notice emitted when a new gas access-control contract is set
  event ConfigACSet(address indexed previous, address indexed current);
  /// @notice emitted when a new source aggregator contract is set
  event SourceAggregatorSet(address indexed previous, address indexed current);

  /// @notice StarkNet messaging contract address - the address is 0.
  error InvalidStarkNetMessagingAddress();
  /// @notice StarkNet uptime feed address - the address is 0.
  error InvalidL2FeedAddress();
  /// @notice Error thrown when the source aggregator address is 0
  error InvalidSourceAggregatorAddress();
  /// @notice Error thrown when the l1 gas price feed address is 0
  error InvalidGasPriceL1FeedAddress();
  /// @notice Error thrown when caller is not the owner and does not have access
  error AccessForbidden();

  /**
   * @param starkNetMessaging the address of the StarkNet Messaging contract
   * @param configAC the address of the AccessController contract managing config access
   * @param gasPriceL1Feed address of the L1 gas price feed (used to approximate bridge L1 -> L2 message cost)
   * @param source the source aggregator that we'll read data from (on retries)
   * @param l2Feed the address of the target L2 contract (Sequencer Uptime Feed)
   */
  constructor(
    address starkNetMessaging,
    address configAC,
    address gasPriceL1Feed,
    address source,
    uint256 l2Feed,
    uint256 gasEstimate
  ) {
    if (starkNetMessaging == address(0)) {
      revert InvalidStarkNetMessagingAddress();
    }

    if (l2Feed == 0) {
      revert InvalidL2FeedAddress();
    }

    if (source == address(0)) {
      revert InvalidSourceAggregatorAddress();
    }

    STARKNET_CROSS_DOMAIN_MESSENGER = IStarknetMessaging(starkNetMessaging);
    L2_UPTIME_FEED_ADDR = l2Feed;

    _setSourceAggregator(source);
    _setConfigAC(configAC);
    _setGasConfig(gasEstimate, gasPriceL1Feed);
  }

  /// @notice converts a bool to uint256.
  function toUInt256(bool x) internal pure returns (uint256 r) {
    assembly {
      r := x
    }
  }

  /**
   * @notice versions:
   *
   * - StarkNetValidator 0.1.0: initial release
   * @inheritdoc TypeAndVersionInterface
   */
  function typeAndVersion() external pure virtual override returns (string memory) {
    return "StarkNetValidator 0.1.0";
  }

  /// @notice Returns the gas configuration for sending cross chain messages.
  function getGasConfig() external view returns (GasConfig memory) {
    return s_gasConfig;
  }

  /// @return config AccessControllerInterface contract address
  function getConfigAC() external view virtual returns (address) {
    return address(s_configAC);
  }

  /**
   * @notice validate method sends an xDomain L2 tx to update Uptime Feed contract on L2.
   * @dev A message is sent using the L1CrossDomainMessenger. This method is accessed controlled.
   * @param currentAnswer new aggregator answer - value of 1 considers the sequencer offline.
   */
  function validate(
    uint256, /* previousRoundId */
    int256, /* previousAnswer */
    uint256, /* currentRoundId */
    int256 currentAnswer
  ) external override checkAccess returns (bool) {
    return _sendUpdateMessageToL2(currentAnswer);
  }

  /**
   * @notice retries to send the latest answer as update message to L2
   * @dev only with access, useful in cases where a previous x-domain message was handeled unsuccessfully.
   */
  function retry() external checkAccess returns (bool) {
    (, int256 latestAnswer, , , ) = AggregatorV3Interface(s_source).latestRoundData();
    return _sendUpdateMessageToL2(latestAnswer);
  }

  /// @notice sends the 'update_status(answer, timestamp)' message via the L1 -> L2 bridge to the L2 feed contract
  function _sendUpdateMessageToL2(int256 answer) internal returns (bool) {
    // Bridge fees are paid on L1
    uint256 fee = _approximateFee();

    // Fill payload with `status` and `timestamp`
    uint256[] memory payload = new uint256[](2);
    bool status = answer == ANSWER_SEQ_OFFLINE;
    payload[0] = toUInt256(status);
    payload[1] = block.timestamp;

    // Make the StarkNet x-domain call.
    // NOTICE: we ignore the output of this call (msgHash, nonce).
    // We also don't raise any events as the 'LogMessageToL2' event will be emitted from the messaging contract.
    STARKNET_CROSS_DOMAIN_MESSENGER.sendMessageToL2{value: fee}(
      L2_UPTIME_FEED_ADDR,
      SELECTOR_STARK_UPDATE_STATUS,
      payload
    );
    return true;
  }

  /// @notice L1 oracle is asked for a fast L1 gas price, and the price multiplied by the configured gas estimate
  function _approximateFee() internal view returns (uint256) {
    (, int256 fastGasPriceInWei, , , ) = AggregatorV3Interface(s_gasConfig.gasPriceL1Feed).latestRoundData();
    return (uint256(fastGasPriceInWei) * s_gasConfig.gasEstimate);
  }

  /**
   * @notice The selector is the starknet_keccak hash of the function name
   * @dev StarkNet keccak is defined as the first 250 bits of the Keccak256 hash.
   *   This is just Keccak256 augmented in order to fit into a field element.
   * @param fn string function name
   */
  function _selectorStarkNet(string memory fn) internal pure returns (uint256) {
    bytes32 digest = keccak256(abi.encodePacked(fn));
    return uint256(digest) % 2**250; // get last 250 bits
  }

  function setGasConfig(uint256 gasEstimate, address gasPriceL1Feed) external onlyOwnerOrConfigAccess {
    _setGasConfig(gasEstimate, gasPriceL1Feed);
    emit GasConfigSet(gasEstimate, gasPriceL1Feed);
  }

  /// @notice internal method that stores the gas configuration
  function _setGasConfig(uint256 gasEstimate, address gasPriceL1Feed) internal {
    if (gasPriceL1Feed == address(0)) {
      revert InvalidGasPriceL1FeedAddress();
    }
    s_gasConfig = GasConfig(gasEstimate, gasPriceL1Feed);
    emit GasConfigSet(gasEstimate, gasPriceL1Feed);
  }

  /**
   * @notice sets config AccessControllerInterface contract
   * @dev only owner can call this
   * @param accessController new AccessControllerInterface contract address
   */
  function setConfigAC(address accessController) external onlyOwner {
    _setConfigAC(accessController);
  }

  /// @notice Internal method that stores the configuration access controller
  function _setConfigAC(address accessController) internal {
    address previousAccessController = address(s_configAC);
    if (accessController != previousAccessController) {
      s_configAC = AccessControllerInterface(accessController);
      emit ConfigACSet(previousAccessController, accessController);
    }
  }

  /**
   * @notice sets the source aggregator AggregatorInterface contract
   * @dev only owner can call this
   * @param source the source aggregator that we'll read data from (on retries)
   */
  function setSourceAggregator(address source) external onlyOwner {
    _setSourceAggregator(source);
  }

  /// @notice Internal method that sets the source aggregator AggregatorInterface contract
  function _setSourceAggregator(address source) internal {
    address previousSource = address(s_source);
    if (source != previousSource) {
      s_source = AggregatorV3Interface(source);
    }
  }

  /// @dev reverts if the caller does not have access to change the configuration
  modifier onlyOwnerOrConfigAccess() {
    if (msg.sender != owner() && (address(s_configAC) != address(0) && !s_configAC.hasAccess(msg.sender, msg.data))) {
      revert AccessForbidden();
    }
    _;
  }

  /**
   * @notice makes this contract payable
   * @dev funds are used to pay the bridge for x-domain messages to L2
   */
  receive() external payable {}

  /**
   * @notice withdraws all funds available in this contract to the msg.sender
   * @dev only owner can call this
   */
  function withdrawFunds() external onlyOwner {
    address payable recipient = payable(msg.sender);
    uint256 amount = address(this).balance;
    Address.sendValue(recipient, amount);
  }

  /**
   * @notice withdraws all funds available in this contract to the address specified
   * @dev only owner can call this
   * @param recipient address where to send the funds
   */
  function withdrawFundsTo(address payable recipient) external onlyOwner {
    uint256 amount = address(this).balance;
    Address.sendValue(recipient, amount);
  }
}
