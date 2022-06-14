import { expect } from "chai";
import { starknet } from "hardhat";
import {
    StarknetContract,
    StarknetContractFactory,
    Account,
} from "hardhat/types/runtime";
import { number, stark } from "starknet";
import { getSelectorFromName } from "starknet/dist/utils/hash";

describe("Multisig integration tests", function () {
    this.timeout(300_000);

    let multisigFactory: StarknetContractFactory;
    let multisig: StarknetContract;
    let account: Account;

    let txIndex = -1; // faster to track this internally than to request from contract

    before(async () => {
        multisigFactory = await starknet.getContractFactory("Multisig");
        account = await starknet.deployAccount("OpenZeppelin");
    })

    it('Deploy contract', async () => {
        multisig = await multisigFactory.deploy({
            owners: [account.starknetContract.address],
            confirmations_required: 1
        });

        expect(multisig).to.be.ok;
    })

    it('quick multisig test', async () => {
        txIndex++;

        const selector = getSelectorFromName("get_owners");
        const payload = {
            to: multisig.address,
            function_selector: selector,
            calldata: [],
        }

        const res = await account.invoke(multisig, "submit_transaction", payload);

        console.log(res);
        const txReciept = await starknet.getTransactionReceipt(res);
        console.log(txReciept);
        console.log(txReciept.l2_to_l1_messages, txReciept.l2_to_l1_messages.length);
        console.log(await starknet.getTransaction(res));

        await account.invoke(multisig, 'confirm_transaction', {
            tx_index: txIndex
        });

        console.log(await account.invoke(multisig, 'execute_transaction', {
            tx_index: txIndex
        }));
    })
})