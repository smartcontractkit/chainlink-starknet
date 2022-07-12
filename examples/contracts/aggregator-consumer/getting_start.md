1. Start by deploy a account with `npx ts-node ./scripts/deploy_accounts.ts`

2. It'll create an `.env`. Open it en copy the ACCOUNT_ADDRESS and ACCOUNT_ADDRESS_2 value.

3. Open that link https://faucet.goerli.starknet.io/ and past the address to get some faucet.

4. Deploy the contracts by running `npx ts-node ./scripts/deploy_contracts.ts.`

5. You can now read decimals and latest round data by runnin `yarn readDecimals` and `yarn readLatestRound`.

6. You can update value and see continuously the data by openning another terminal and run `npx ts-node ./scripts/updateLatestRound.ts` in one terminal and `npx ts-node ./scripts/readContinuously` into another one.
