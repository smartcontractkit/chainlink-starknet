import { defaultProvider, ec } from 'starknet'
import { loadContract } from './index'
import fs from 'fs'

const CONSUMER_NAME = 'Aggregator_consumer'
const MOCK_NAME = 'MockAggregator'
const DECIMALS = 18

async function main() {
  const MockArtifact = loadContract(MOCK_NAME)
  const AggregatorArtifact = loadContract(CONSUMER_NAME)

  const mockDeploy = await defaultProvider.deployContract({
    contract: MockArtifact,
    constructorCalldata: [DECIMALS],
  })

  const consumerDeploy = await defaultProvider.deployContract({
    contract: AggregatorArtifact,
    constructorCalldata: [mockDeploy.address as string],
  })

  fs.appendFile(__dirname + '/.env', '\nCONSUMER=' + consumerDeploy.address, function (err) {
    if (err) throw err
  })
  fs.appendFile(__dirname + '/.env', '\nMOCK=' + mockDeploy.address, function (err) {
    if (err) throw err
  })
}

main()
