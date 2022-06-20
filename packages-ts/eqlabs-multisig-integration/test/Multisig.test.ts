import { expect } from "chai";
import { starknet } from "hardhat";
import {
    StarknetContract,
    StarknetContractFactory,
    Account,
} from "hardhat/types/runtime";
import { getSelectorFromName } from "starknet/dist/utils/hash";
import { number } from "starknet";

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

    it('should submit & confirm transaction', async () => {
        txIndex++;
        const selector = getSelectorFromName("get_owners");
        const payload = {
            to: multisig.address,
            function_selector: selector,
            calldata: [],
        }

        const res = await account.invoke(multisig, "submit_transaction", payload);
        const txReciept = await starknet.getTransactionReceipt(res);

        expect(txReciept.events.length).to.equal(1);
        expect(txReciept.events[0].data.length).to.equal(3);
        expect(txReciept.events[0].data[1]).to.equal(number.toHex(number.toBN(txIndex, "hex")));

        await account.invoke(multisig, 'confirm_transaction', {
            tx_index: txIndex
        });

        await account.invoke(multisig, 'execute_transaction', {
            tx_index: txIndex
        });
    })
})