import {
  InspectCommandConfig,
  IStarknetProvider,
  makeInspectionCommand,
} from '@chainlink/starknet-gauntlet'
import { shortString, validateAndParseAddress } from 'starknet'
import { CATEGORIES } from '../../../lib/categories'
import { ocr2ContractLoader } from '../../../lib/contracts'

type QueryResult = {
  typeAndVersion: string
  description: string
  owner: string
  decimals: number
  latestConfigDetails: {
    configCount: number
    blockNumber: number
    configDigest: string
  }
  transmitterInfo: {
    transmitter: string
    owedPayment: string
  }[]
  billing: {
    observationPaymentGjuels: string
    transmissionPaymentGjuels: string
    gasBase: string
    gasPerSignature: string
  }
  linkAvailableForPayment: {
    isNegative: boolean
    absoluteDifference: string
  }
}

const makeComparisionData = (provider: IStarknetProvider) => async (
  results: any[],
  input: null,
  contractAddress: string,
): Promise<{
  toCompare: null
  result: QueryResult
}> => {
  const typeAndVersion = shortString.decodeShortString(results[0])
  const description = shortString.decodeShortString(results[1])
  const owner = validateAndParseAddress(results[2])
  const decimals = parseInt(results[3])
  const latestConfigDetails = {
    configCount: parseInt(results[4][0]),
    blockNumber: parseInt(results[4][1]),
    configDigest: '0x' + results[4][2].toString(16),
  }
  const transmitters = results[5].map((address) => validateAndParseAddress(address))
  let transmitterInfo = []
  for (const transmitter of transmitters) {
    const owedPayment = await provider.provider.callContract({
      contractAddress,
      entrypoint: 'owed_payment',
      calldata: [transmitter],
    })
    transmitterInfo.push({
      transmitter,
      owedPayment: parseInt(owedPayment[0]).toString(),
    })
  }
  const billing = {
    observationPaymentGjuels: parseInt(results[6].observation_payment_gjuels).toString(),
    transmissionPaymentGjuels: parseInt(results[6].transmission_payment_gjuels).toString(),
    gasBase: parseInt(results[6].gas_base).toString(),
    gasPerSignature: parseInt(results[6].gas_per_signature).toString(),
  }
  const linkAvailableForPayment = {
    isNegative: results[7][0],
    absoluteDifference: parseInt(results[7][1]).toString(),
  }
  return {
    toCompare: null,
    result: {
      typeAndVersion,
      description,
      owner,
      decimals,
      latestConfigDetails,
      transmitterInfo,
      billing,
      linkAvailableForPayment,
    },
  }
}

const commandConfig: InspectCommandConfig<null, null, null, QueryResult> = {
  ux: {
    category: CATEGORIES.OCR2,
    function: 'inspect',
    examples: ['yarn gauntlet ocr2:inspect --network=<NETWORK> <CONTRACT_ADDRESS>'],
  },
  queries: [
    'type_and_version',
    'description',
    'owner',
    'decimals',
    'latest_config_details',
    'transmitters',
    'billing',
    'link_available_for_payment',
  ],
  makeComparisionData,
  loadContract: ocr2ContractLoader,
}

export default makeInspectionCommand(commandConfig)
