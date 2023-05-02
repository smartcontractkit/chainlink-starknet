0. `.env` contains predeploy and funded account details, such as DEPLOYER_ACCOUNT_ADDRESS and DEPLOYER_PRIVATE_KEY

1. Set up the network `export STARKNET_NETWORK=alpha-goerli`

2. Choose a wallet provider `export STARKNET_WALLET=starkware.starknet.wallets.open_zeppelin.OpenZeppelinAccount`

3. If you don't have an account create it with the command `starknet new_account`

4. Open that link https://faucet.goerli.starknet.io/ and past the address to get some faucet. If you can not get faucet you can transfer some eth from one of you L1 accounts to starknet https://goerli.starkgate.starknet.io/.

5. When you have been able to get some faucet you can deploy your account `starknet deploy_account` (for more information take a look at this documentation https://starknet.io/docs/hello_starknet/account_setup.html).

6. Create a `.env` in `exemples/contracts/aggregator-consumer`and copy past your account address under `DEPLOYER_ACCOUNT_ADDRESS` and you private key under `DEPLOYER_PRIVATE_KEY`. You can find the private key in your home directory under `.starknet_accounts` folder in `starknet_open_zeppelin_accounts.json`.

7. Start by deploying accounts with `npx ts-node ./scripts/deploy_accounts.ts`

8. It'll write in previously created `.env`.

9. Open that link https://faucet.goerli.starknet.io/ and past the address of the new accounts to get some faucet.

10. Deploy the contracts by running `npx ts-node ./scripts/deploy_contracts.ts.`

11. You can now read decimals and latest round data by running `yarn readDecimals` and `yarn readLatestRound`.

12. You can update value and see continuously the data by openning another terminal and run `npx ts-node ./scripts/updateLatestRound.ts` in one terminal and `npx ts-node ./scripts/readContinuously` into another one.
