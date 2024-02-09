import { makeProvider } from '@chainlink/starknet-gauntlet'
import deployCommand from '../../src/commands/account/deploy'
import { registerExecuteCommand, TIMEOUT, LOCAL_URL } from '@chainlink/starknet-gauntlet/test/utils'
import { accountContractLoader } from '../../src/lib/contracts'
import { Contract } from 'starknet'

describe('OZ Account Contract', () => {
  let publicKey: string
  let contractAddress: string

  it(
    'Deployment',
    async () => {
      const command = await registerExecuteCommand(deployCommand).create({}, [])

      const report = await command.execute()
      expect(report.responses[0].tx.status).toEqual('ACCEPTED')

      contractAddress = report.responses[0].contract
      publicKey = report.data.publicKey

      const { contract: oz } = accountContractLoader()
      const ozContract = new Contract(oz.abi, contractAddress, makeProvider(LOCAL_URL).provider)
      const onChainPubKey = await ozContract.getPublicKey()
      expect(onChainPubKey).toEqual(BigInt(publicKey))
    },
    TIMEOUT,
  )
})
