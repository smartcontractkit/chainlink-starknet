import { starknet } from 'hardhat'
import { expect } from 'chai'
import { StarknetContract, ArgentAccount } from "hardhat/types/runtime";
import BN from 'bn.js';
import { toBN } from 'starknet/utils/number';
import { loadStarkgateContract } from "../utils"

const NAME = starknet.shortStringToBigInt("LINK")
const SYMBOL = starknet.shortStringToBigInt("LINKTOKEN")
const DECIMALS = 18

function uint256ToBigInt(uint256: {low : bigint, high: bigint}) {
    return new BN(((BigInt(uint256.high) << BigInt(128)) + BigInt(uint256.low)).toString());
  }

describe('ContractTests', function () {
    this.timeout(600_000);
    let accountMinter: ArgentAccount;
    let accountUser1: ArgentAccount;
    let accountUser2: ArgentAccount;
    let ERC20Contract: StarknetContract;

    before(async () => {
        accountMinter = (await starknet.deployAccount("Argent")) as ArgentAccount;
        console.log("accountMinter: ", accountMinter.starknetContract.address)
        // const starkGateContractToTest = loadStarkgateContract('ERC20')
        // console.log("starkgateContract", starkGateContractToTest.abi)
        let ERC20Factory = await starknet.getContractFactory("ERC20.cairo")
        ERC20Contract = await ERC20Factory.deploy({name: NAME, symbol: SYMBOL, decimals: DECIMALS, minter_address: accountMinter.starknetContract.address })
        console.log("ERC20Contract: ", ERC20Contract.address)

        accountUser1 = (await starknet.deployAccount("Argent")) as ArgentAccount;
        console.log("accountUser1: ", accountUser1.starknetContract.address)

        accountUser2 = (await starknet.deployAccount("Argent")) as ArgentAccount;
        console.log("accountUser2: ", accountUser2.starknetContract.address)
    });

    it('Test Permissioned Mint', async () => {

        await accountMinter.invoke(ERC20Contract, 'permissionedMint', {recipient: accountUser1.starknetContract.address, amount: { low: 15n, high: 0n } })
        await new Promise((resolve) => setTimeout(resolve,30000))
        {
            const { balance: balance1 } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance = uint256ToBigInt(balance1)
            expect(sum_balance).to.deep.equal(toBN(15));
        }

        await accountMinter.invoke(ERC20Contract, 'permissionedMint', {recipient: accountUser2.starknetContract.address, amount: { low: 12n, high: 0n } })
        
        {
            const { totalSupply: totalSupply1 } = await ERC20Contract.call('totalSupply', {})
            let supply = uint256ToBigInt(totalSupply1)
            expect(supply).to.deep.equal(toBN(27));

            const { balance: balance1 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance = uint256ToBigInt(balance1)
            expect(sum_balance).to.deep.equal(toBN(12));
        }

    });
    it('Test Permissioned Burn', async () => {
        await new Promise((resolve) => setTimeout(resolve,30000))
        await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {account: accountUser1.starknetContract.address, amount: { low: 3n, high: 0n } })
        {
            const { totalSupply: totalSupply1 } = await ERC20Contract.call('totalSupply', {})
            let supply = uint256ToBigInt(totalSupply1)
            expect(supply).to.deep.equal(toBN(24));

            const { balance: balance1 } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance = uint256ToBigInt(balance1)
            expect(sum_balance).to.deep.equal(toBN(12));
        }

        await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {account: accountUser2.starknetContract.address, amount: { low: 10n, high: 0n } })
        {
            const { totalSupply: totalSupply1 } = await ERC20Contract.call('totalSupply', {})
            let supply = uint256ToBigInt(totalSupply1)
            expect(supply).to.deep.equal(toBN(14));

            const { balance: balance1 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance = uint256ToBigInt(balance1)
            expect(sum_balance).to.deep.equal(toBN(2));
        }
        {
            try {
                await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {account: accountUser2.starknetContract.address, amount: { low: 3n, high: 10n } })
            } catch (error) {
                console.log("ERROR NOT ENOUGHT BALANCE TO BURN")
            }
            try {
                await accountMinter.invoke(ERC20Contract, 'permissionedBurn', {account: accountUser1.starknetContract.address, amount: { low: 3n, high: 13n } })
            } catch (error) {
                console.log("ERROR NOT ENOUGHT BALANCE TO BURN")
            }
            try {
                await accountUser1.invoke(ERC20Contract, 'permissionedBurn', {account: accountUser2.starknetContract.address, amount: { low: 3n, high: 10n } })
            } catch (error) {
                console.log("WRONG MINTER")
            }
            try {
                await accountUser2.invoke(ERC20Contract, 'permissionedBurn', {account: accountUser1.starknetContract.address, amount: { low: 3n, high: 13n } })
            } catch (error) {
                console.log("WRONG MINTER")
            } 
        }
    });
    it('Test transfer', async () => {

        await accountMinter.invoke(ERC20Contract, 'permissionedMint', {recipient: accountUser2.starknetContract.address, amount: { low: 3n, high: 0n } })
        {
            await accountUser1.invoke(ERC20Contract, 'transfer', {recipient: accountUser2.starknetContract.address, amount: { low: 3n, high: 0n } })
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance = uint256ToBigInt(balance)
            expect(sum_balance).to.deep.equal(toBN(9));

            const { balance: balance1 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance1)
            expect(sum_balance1).to.deep.equal(toBN(8));
        }
        await accountUser2.invoke(ERC20Contract, 'transfer', {recipient: accountUser1.starknetContract.address, amount: { low: 4n, high: 0n } })
        {
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance)
            expect(sum_balance1).to.deep.equal(toBN(4));

            const { balance: balance2 } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance2 = uint256ToBigInt(balance2)
            expect(sum_balance2).to.deep.equal(toBN(13));
        }   
        {
            try {
                await accountUser2.invoke(ERC20Contract, 'transfer', {recipient: accountUser1.starknetContract.address, amount: { low: 12n, high: 0n } })
            } catch (error) {
                console.log("ERROR INSUFFICIENT BALANCE")
            }
            try {
                await accountUser1.invoke(ERC20Contract, 'transfer', {recipient: accountUser2.starknetContract.address, amount: { low: 17n, high: 0n } })
            } catch (error) {
                console.log("ERROR INSUFFICIENT BALANCE")
            }
            
        }

    });
    it('Test transferFrom', async () => {
        await new Promise((resolve) => setTimeout(resolve,30000))
        await  accountUser2.invoke(ERC20Contract, 'increaseAllowance', {spender: accountMinter.starknetContract.address, added_value: { low: 7n, high: 0n } })
        await  accountUser1.invoke(ERC20Contract, 'increaseAllowance', {spender: accountMinter.starknetContract.address, added_value: { low: 7n, high: 0n } })
        await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser1.starknetContract.address, recipient: accountUser2.starknetContract.address, amount: { low: 3n, high: 0n } })
        {
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance)
            expect(sum_balance1).to.deep.equal(toBN(10));

            const { balance: balance2 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance2 =  uint256ToBigInt(balance2)
            expect(sum_balance2).to.deep.equal(toBN(7));
        }

        await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser2.starknetContract.address, recipient: accountUser1.starknetContract.address, amount: { low: 4n, high: 0n } })
        {
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance)
            expect(sum_balance1).to.deep.equal(toBN(14));

            const { balance: balance2 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance2 =  uint256ToBigInt(balance2)
            expect(sum_balance2).to.deep.equal(toBN(3));
        }

        {
            try {
                await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser2.starknetContract.address, recipient: accountUser1.starknetContract.address, amount: { low: 8n, high: 10n } })
            } catch (error) {
                console.log("ERROR INSUFFICIENT BALANCE")
            }
            try {
                await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser1.starknetContract.address, recipient: accountUser2.starknetContract.address, amount: { low: 8n, high: 20n } })
            } catch (error) {
                console.log("ERROR INSUFFICIENT BALANCE")
            }
            
        }

    });

    it('Test increase and decrease Alllowance', async () => {
        await new Promise((resolve) => setTimeout(resolve,60000))
        await  accountUser2.invoke(ERC20Contract, 'increaseAllowance', {spender: accountMinter.starknetContract.address, added_value: { low: 7n, high: 0n } })
        await  accountUser1.invoke(ERC20Contract, 'increaseAllowance', {spender: accountMinter.starknetContract.address, added_value: { low: 7n, high: 0n } })

        await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser1.starknetContract.address, recipient: accountUser2.starknetContract.address, amount: { low: 3n, high: 0n } })
        {
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance)
            expect(sum_balance1).to.deep.equal(toBN(11));

            const { balance: balance2 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance2 =  uint256ToBigInt(balance2)
            expect(sum_balance2).to.deep.equal(toBN(6));
        }
        await  accountUser2.invoke(ERC20Contract, 'increaseAllowance', {spender: accountMinter.starknetContract.address, added_value: { low: 15n, high: 0n } })
        await  accountUser1.invoke(ERC20Contract, 'increaseAllowance', {spender: accountMinter.starknetContract.address, added_value: { low: 15n, high: 0n } })

        await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser1.starknetContract.address, recipient: accountUser2.starknetContract.address, amount: { low: 11n, high: 0n } })
        {
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance)
            expect(sum_balance1).to.deep.equal(toBN(0));

            const { balance: balance2 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance2 =  uint256ToBigInt(balance2)
            expect(sum_balance2).to.deep.equal(toBN(17));
        }
        await  accountUser2.invoke(ERC20Contract, 'decreaseAllowance', {spender: accountMinter.starknetContract.address, subtracted_value: { low: 10n, high: 0n } })
        
        await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser2.starknetContract.address, recipient: accountUser1.starknetContract.address, amount: { low: 1n, high: 0n } })
        {
            const { balance: balance } = await ERC20Contract.call('balanceOf', { account: accountUser1.starknetContract.address })
            let sum_balance1 = uint256ToBigInt(balance)
            expect(sum_balance1).to.deep.equal(toBN(1));

            const { balance: balance2 } = await ERC20Contract.call('balanceOf', { account: accountUser2.starknetContract.address })
            let sum_balance2 =  uint256ToBigInt(balance2)
            expect(sum_balance2).to.deep.equal(toBN(16));
        }
        {
            try {
                await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser2.starknetContract.address, recipient: accountUser1.starknetContract.address, amount: { low: 8n, high: 10n } })
            } catch (error) {
                console.log("ERROR INSUFFICIENT BALANCE")
            }
            try {
                await accountMinter.invoke(ERC20Contract, 'transferFrom', {sender: accountUser1.starknetContract.address, recipient: accountUser2.starknetContract.address, amount: { low: 4n, high: 0n } })
            } catch (error) {
                console.log("ERROR INSUFFICIENT BALANCE")
            }
            
        }
    });

});
