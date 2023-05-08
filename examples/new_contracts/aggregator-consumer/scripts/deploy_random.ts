import dotenv from 'dotenv'
import { createDeployerAccount, makeProvider } from './utils'
import { CompiledContract, json, ec, Account, Provider, constants } from 'starknet'
import fs from 'fs'

dotenv.config({ path: __dirname + '/../.env' })

const run = async () => {
    const provider = makeProvider()
    const account = createDeployerAccount(provider)

    await account.declareAndDeploy({
        compiledClassHash: '0x4309b2a1abce2c4050a0c58876ce80de06f7d6adecd55c5914625c426951902',
        contract: json.parse(
            fs.readFileSync(`${__dirname}/../../../contracts/target/release/chainlink_SequencerUptimeFeed.json`).toString('ascii'),
        ),
        // casm: json.parse(
        //     fs.readFileSync(`${__dirname}/../../../contracts/chainlink_SequencerUptimeFeed.casm.json`).toString('ascii'),
        // ),
        constructorCalldata: ['0', account.address]
    })
}

run()
