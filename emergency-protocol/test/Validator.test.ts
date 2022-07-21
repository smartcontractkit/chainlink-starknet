import { ethers, starknet, network } from 'hardhat'
import { BigNumber, Contract, ContractFactory } from 'ethers'
import { StarknetContractFactory, StarknetContract, HttpNetworkConfig } from 'hardhat/types'
import { expect } from 'chai'
import { SignerWithAddress } from '@nomiclabs/hardhat-ethers/signers'
/// Pick ABIs from compilation
// @ts-ignore
import { abi as optimismSequencerStatusRecorderAbi } from '../../../artifacts/src/v0.8/dev/OptimismSequencerUptimeFeed.sol/OptimismSequencerUptimeFeed.json'
// @ts-ignore
import { abi as optimismL1CrossDomainMessengerAbi } from '@eth-optimism/contracts/artifacts/contracts/L1/messaging/L1CrossDomainMessenger.sol'
// @ts-ignore
import { abi as aggregatorAbi } from '../../../artifacts/src/v0.8/interfaces/AggregatorV2V3Interface.sol/AggregatorV2V3Interface.json'

/**
 * Receives a hex address, converts it to bigint, converts it back to hex.
 * This is done to strip leading zeros.
 * @param address a hex string representation of an address
 * @returns an adapted hex string representation of the address
 */
function adaptAddress(address: string) {
  return '0x' + BigInt(address).toString(16)
}

/**
 * Expects address equality after adapting them.
 * @param actual
 * @param expected
 */
export function expectAddressEquality(actual: string, expected: string) {
  expect(adaptAddress(actual)).to.equal(adaptAddress(expected))
}

describe('StarknetValidator', () => {
  /** Fake L2 target */
  const networkUrl: string = (network.config as HttpNetworkConfig).url
  let Validator: Contract
  let MockStarknetMessaging: ContractFactory
  let mockStarknetMessaging: Contract
  //   let mockOptimismL1CrossDomainMessenger: Contract
  let deployer: SignerWithAddress
  let eoaValidator: SignerWithAddress

  let L2contractFactory: StarknetContractFactory
  let l2contract: StarknetContract

  // const L1_STARKNET_CORE = "0xc662c410C0ECf747543f5bA90660f6ABeBD9C8c4"

  before(async () => {

    const account = await starknet.deployAccount('OpenZeppelin')

    L2contractFactory = await starknet.getContractFactory('Mock_Uptime_feed')
    l2contract = await L2contractFactory.deploy()
    console.log('L2 address: ', l2contract.address)

    const accounts = await ethers.getSigners()
    deployer = accounts[0]
    eoaValidator = accounts[1]

    const ValidatorFactory = await ethers.getContractFactory('MockValidator', deployer)
    MockStarknetMessaging = await ethers.getContractFactory('MockStarknetMessaging', deployer)

    mockStarknetMessaging = await MockStarknetMessaging.deploy()
    await mockStarknetMessaging.deployed()

    Validator = await ValidatorFactory.deploy(mockStarknetMessaging.address, l2contract.address)
    console.log('Validator address: ', Validator.address)

    await account.invoke(l2contract, 'set_l1_sender', {address: Validator.address})
  })

  describe('#validate', () => {
    it('reverts if called by account with no access', async () => {
      await expect(Validator.connect(eoaValidator).validate(0, 0, 1, 1)).to.be.revertedWith('No access')
    })

    it('should connect successfully', async () => {
      Validator.addAccess(eoaValidator.address)
      Validator.connect(eoaValidator).validate(0, 0, 1, 1)
    })

    it('should deploy the messaging contract', async () => {
      Validator.addAccess(eoaValidator.address)

      const { address: deployedTo, l1_provider: L1Provider } = await starknet.devnet.loadL1MessagingContract(networkUrl)

      expect(deployedTo).not.to.be.undefined
      expect(L1Provider).to.equal(networkUrl)

      const { address: loadedFrom } = await starknet.devnet.loadL1MessagingContract(
        networkUrl,
        mockStarknetMessaging.address,
      )

      expect(mockStarknetMessaging.address).to.equal(loadedFrom)

      await Validator.connect(eoaValidator).validate(0, 0, 1, 1)

      const flushL1Response = await starknet.devnet.flush()
      const flushL1Messages = flushL1Response.consumed_messages.from_l1
      expect(flushL1Messages).to.have.a.lengthOf(1)
      expect(flushL1Response.consumed_messages.from_l2).to.be.empty

      expectAddressEquality(flushL1Messages[0].args.from_address, Validator.address)
      expectAddressEquality(flushL1Messages[0].args.to_address, l2contract.address)
      expectAddressEquality(flushL1Messages[0].address, mockStarknetMessaging.address)
    })
  })
})
