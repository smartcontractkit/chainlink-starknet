import { starknet } from 'hardhat'
import { StarknetContract } from 'hardhat/types'
import { constants, encode, number } from 'starknet'
import { BigNumberish } from 'starknet/utils/number'
$
const OCR2_ADDRESS = 42
const ACCOUNT_ADD = '0x046057141187f7ce7f6477c5f6df0f65c28a4a62dfb4504bac72d164f9f6af73'

function toFelt(int: number | BigNumberish): BigNumberish {
  let prime = number.toBN(encode.addHexPrefix(constants.FIELD_PRIME))
  return number.toBN(int).umod(prime)
}
describe('ContractTests', function () {
  this.timeout(600_000)
  let OCR2Contract: StarknetContract

  before(async () => {
    let minAnswer = -10
    let maxAnswer = 1000000000
    const tokenContractFactory = await starknet.getContractFactory('token')
    const token = await tokenContractFactory.deploy({
      name: starknet.shortStringToBigInt('LINK Token'),
      symbol: starknet.shortStringToBigInt('LINK'),
      decimals: 18,
      initial_supply: { high: 1n, low: 0n },
      recipient: BigInt(ACCOUNT_ADD as string),
      owner: BigInt(ACCOUNT_ADD as string),
    })
    console.log('token: ', token.address)

    const aggregatorContractFactory = await starknet.getContractFactory('aggregator')
    const aggregator = await aggregatorContractFactory.deploy({
      owner: BigInt(ACCOUNT_ADD as string),
      link: BigInt(token.address),
      min_answer: toFelt(minAnswer),
      max_answer: toFelt(maxAnswer),
      billing_access_controller: 0, // TODO: billing AC
      decimals: 8,
      description: starknet.shortStringToBigInt('FOO/BAR'),
    })
    console.log('aggregator: ', aggregator.address)

    let OCR2Factory = await starknet.getContractFactory('./OCR2_consumer.cairo')
    OCR2Contract = await OCR2Factory.deploy({ address: OCR2_ADDRESS })
    console.log('OCR2Contract: ', OCR2Contract.address)
  })

  it('Test readStoredRound and storeLatestRound', async () => {
    const { round: round } = await OCR2Contract.call('readStoredRound', {})
    console.log('latestRound= ', round)
  })
})
