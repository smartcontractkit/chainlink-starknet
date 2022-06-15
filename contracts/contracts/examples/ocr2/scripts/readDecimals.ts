import { Account, Contract, defaultProvider, ec } from 'starknet'
import { loadContract } from './index'
import dotenv from 'dotenv'

const CONSUMER_NAME = 'OCR2_consumer'
let account: Account
let consumer: Contract

dotenv.config({ path: __dirname + '/.env' })
async function main() {
  const keyPair = ec.getKeyPair(process.env.PRIVATE_KEY as string)
  account = new Account(defaultProvider, process.env.ACCOUNT_ADDRESS as string, keyPair)
  const OCR2Artifact = loadContract(CONSUMER_NAME)
  consumer = new Contract(OCR2Artifact.abi, process.env.CONSUMER as string)

  const transaction = await account.execute(
    {
      contractAddress: consumer.address,
      entrypoint: 'storeDecimals',
      calldata: [],
    },
    [consumer.abi],
    { maxFee: 30000000000000 },
  )

  console.log('Waiting for Tx to be Accepted on Starknet...')
  await defaultProvider.waitForTransaction(transaction.transaction_hash)

  const decimals = await account.callContract({
    contractAddress: consumer.address,
    entrypoint: 'readDecimals',
    calldata: [],
  })

  console.log('decimals= ', parseInt(decimals.result[0], 16))
}

// async function callFunction() {
// const starkKeyPair = ec.genKeyPair();

// console.log("Waiting for Tx to be Accepted on Starknet - ACCOUNT Deployment...");

// const account : Account = new Account(defaultProvider, Accountaddress, starkKeyPair)

// const OCR2Artifact = loadContract(CONSUMER_NAME)

// const OCR2Deploy = await defaultProvider.deployContract({
//     contract: OCR2Artifact,
//     constructorCalldata: [process.env.MOCK as string],
// });

//     console.log("Waiting for Tx to be Accepted on Starknet - OCR2 Deployment...");
//     await defaultProvider.waitForTransaction(OCR2Deploy.transaction_hash);
//     console.log("OCR2Deploy= ", OCR2Deploy.address)

//     const OCR2Address = OCR2Deploy.address;
//     const OCR2 = new Contract(OCR2Artifact.abi, OCR2Address as string, account);

//     await OCR2.storeDecimals()
//     const decimals = await OCR2.readDecimals()

//     console.log("decimals= ", decimals)

//     const transaction = await account.execute({
//         contractAddress: mock.address,
//         entrypoint: "set_latest_round_data",
//         calldata: [number.toFelt(transmission.answer), number.toFelt(transmission.block_num), number.toFelt(transmission.observation_timestamp), number.toFelt(transmission.transmission_timestamp)],
//     },[mock.abi],{maxFee : 30000000000000})
//     console.log("Waiting for Tx to be Accepted on Starknet - OCR2 Deployment...");
//     console.log(transaction.code)
//     await defaultProvider.waitForTransaction(transaction.transaction_hash)

//     const response = await account.callContract({
//         contractAddress: mock.address,
//         entrypoint: "latest_round_data",
//         calldata: [],
//     })
// }

main()
// .then(
//     () => process.exit(),
//     err => {
//         console.error(err);
//         process.exit(-1);
//     },
// );
