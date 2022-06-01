import { starknet } from 'hardhat'
import { expect } from 'chai'
import { StarknetContract, ArgentAccount } from "hardhat/types/runtime";


const NAME = starknet.shortStringToBigInt("LINK")
const SYMBOL = starknet.shortStringToBigInt("LINKTOKEN")
const DECIMALS = 18
const L1_BRIDGE_ADDRESS = 42n
const L1_ACCOUNT = 1

describe('ContractTestsBridge', function () {
    this.timeout(600_000);
    let accountUser1: ArgentAccount;
    let tokenBridgeContract: StarknetContract;
    let ERC20Contract: StarknetContract;

    before(async () => { 

        accountUser1 = (await starknet.deployAccount("Argent")) as ArgentAccount;
        console.log("accountUser1: ", accountUser1.starknetContract.address)

        
        let tokenBridgeFactory = await starknet.getContractFactory('token_bridge.cairo')
        tokenBridgeContract = await tokenBridgeFactory.deploy({governor_address: accountUser1.starknetContract.address })
        console.log("tokenBridgeContract: ", tokenBridgeContract.address)
        
        let ERC20Factory = await starknet.getContractFactory('ERC20.cairo')
        ERC20Contract = await ERC20Factory.deploy({name: NAME, symbol: SYMBOL, decimals: DECIMALS, minter_address: tokenBridgeContract.address })
        console.log("ERC20Contract: ", ERC20Contract.address)

    });

    it('Test Set and Get function', async () => {
        await new Promise((resolve) => setTimeout(resolve,30000))
        {
            await accountUser1.invoke(tokenBridgeContract, 'set_l1_bridge', {l1_bridge_address: L1_BRIDGE_ADDRESS})
            const { res: l1_address } = await tokenBridgeContract.call('get_l1_bridge', {})
            expect(l1_address.toString()).to.deep.equal(L1_BRIDGE_ADDRESS.toString());
        }
        {
            await accountUser1.invoke(tokenBridgeContract, 'set_l2_token', {l2_token_address: ERC20Contract.address})
            const { res: l2_address } = await tokenBridgeContract.call('get_l2_token', {})
            expect("0x0" + l2_address.toString(16)).to.deep.equal(ERC20Contract.address.toString());
        }
    });

});
