import {defaultProvider } from 'starknet'
import { loadORC2ConsumerContract } from "./index";

const CONSUMER_NAME = "OCR2_consumer"
const OCR2_FEED = 42

async function main() {
    const OCR2Artifact = loadORC2ConsumerContract(CONSUMER_NAME)

    const consumerDeploy = await defaultProvider.deployContract({
        contract: OCR2Artifact,
        constructorCalldata: [OCR2_FEED],
    });

    console.log("Tx hash: ", consumerDeploy.transaction_hash)
    console.log("address: ", consumerDeploy.address)
}

main();