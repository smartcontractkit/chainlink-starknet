import fs from 'fs'
import { Account, Provider, Contract, json, ec, constants } from 'starknet'


// The Cairo contract that is compiled and ready to declare and deploy
const consumerContractName = 'ProxyConsumer'

/**
 * Network: Starknet Goerli testnet
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
const accountPrivateKey = process.env.DEPLOYER_PRIVATE_KEY as string
const starkKeyPub = ec.starkCurve.getStarkKey(accountPrivateKey)

export async function deployContract() {
  const provider = new Provider({
    sequencer: {
      // Starknet network: Either goerli-alpha or mainnet-alpha
      network: constants.NetworkName.SN_GOERLI,
    },
  })

  const account = new Account(provider, accountAddress, accountPrivateKey)

  const consumerContract = json.parse(
    fs
      .readFileSync(
        `${__dirname}/../target/release/proxy_consumer_${consumerContractName}.sierra.json`,
      )
      .toString('ascii'),
  )

  const declareDeployConsumer = await account.declareAndDeploy({
    contract: consumerContract,
    casm: json.parse(
      fs
        .readFileSync(
          `${__dirname}/../target/release/proxy_consumer_${consumerContractName}.casm.json`,
        )
        .toString('ascii'),
    ),
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
