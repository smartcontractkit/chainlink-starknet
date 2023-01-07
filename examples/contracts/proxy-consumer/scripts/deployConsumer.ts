import fs from 'fs'
import { Account, Provider, Contract, json, ec } from 'starknet'

// StarkNet network: Either goerli-alpha or mainnet-alpha
const network = 'goerli-alpha'

// The Cairo contract that is compiled and ready to declare and deploy
const consumerContractName = 'Proxy_consumer'
const consumerClassHash = '0x75f25b359402fa046e2b9c17d00138772b51c647c0352eb16954e9e39df4ca6'

/**
 * Network: StarkNet Goerli testnet
 * Aggregator: LINK/USD
 * Address: 0x2579940ca3c41e7119283ceb82cd851c906cbb1510908a913d434861fdcb245
 * Find more feed address at:
 * https://docs.chain.link/data-feeds/price-feeds/addresses?network=starknet
 */
const dataFeedAddress = '0x2579940ca3c41e7119283ceb82cd851c906cbb1510908a913d434861fdcb245'

/** Environment variables for a deployed and funded account to use for deploying contracts
 * Find your OpenZeppelin account address and private key at:
 * ~/.starknet_accounts/starknet_open_zeppelin_accounts.json
 */
const accountAddress = process.env.DEPLOYER_ACCOUNT_ADDRESS as string
const accountKeyPair = ec.getKeyPair(process.env.DEPLOYER_PRIVATE_KEY as string)

export async function deployContract() {
  const provider = new Provider({
    sequencer: {
      network: network,
    },
  })

  const account = new Account(provider, accountAddress, accountKeyPair)
  
  const consumerContract = json.parse(
    fs
      .readFileSync(`${__dirname}/../starknet-artifacts/contracts/${consumerContractName}.cairo/${consumerContractName}.json`)
      .toString('ascii'),
  )

  const declareDeployConsumer = await account.declareDeploy({
    contract: consumerContract,
    classHash: consumerClassHash,
    constructorCalldata: [dataFeedAddress as string],
  })

  const consumerDeploy = new Contract(
    consumerContract.abi,
    declareDeployConsumer.deploy.contract_address,
    provider,
  )

  console.log('Contract address: ' + consumerDeploy.address)
  console.log('Transaction hash: ' + declareDeployConsumer.deploy.transaction_hash)

}

deployContract()
