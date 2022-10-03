# Starknet Emergency Protocol

## Background

Today Chainlink price feeds are used by many DeFi protocols to secure billions of dollars. Whilst feeds report fresh prices the majority of the time, L2 feeds will report stale price data whenever the L2 chain stops producing new blocks. This can happen whenever the L2 Sequencer fails to process any new transactions. Whenever this happens, an arbitrage opportunity is created for malicious actors to take advantage of the price difference between the price inside and outside the L2 chain.

The Starknet Emergency Protocol provides a way for Chainlink price feed consumers to guard against the scenario described above. The protocol tracks the last known health of the Sequencer and reports it's health on chain along with the timestamp of when it either comes back online or goes offline. This allows consuming contracts to implement a grace period in their contracts to revert transactions whenever the Sequencer is down.

**WARNING**
This implementation only deals with **centralized** Sequencers. As Starknet has plans to decentralize the Sequencer in the future,
this protocol will either need to be redesigned or even deprecated. The reason for this is that the protocol relies on polling the `pending` block from the Sequencer to determine if new transactions are being added. A decentralized Starknet sequencer will no longer allow this hence breaking the protocol. For more information on how the pending block is used, take a look at the [Layer2 Sequencer Health External Adapter](#layer2-sequencer-health-external-adapter) section.

[Official Docs](https://docs.chain.link/docs/l2-sequencer-flag/)

## Architecture

The diagram above illustrates the general path of how the Sequencer’s status is relayed from L1 to L2.

[![](https://mermaid.ink/img/pako:eNqNk99PwjAQx_-VprxCwo8YtSQmOtgTCQaiL46H2l1HQ9fOrosSwv9u59aBzBH3svXu8737Xi87YKZjwARzqT_ZlhqLFqtIIfcwSfN8Bhy5sFBSqB3iQkrSm0wmQRhOUSfGtNSG9DjnLSi31OwU2LrU03B0G46nXdCVQkyrvEjB1IXC-9Hd8GbaBbULLYMVIeTkeTB4QI9JYiCh1qFnqYpf155cxtsrJWv4KEAxMGfxWuAzP6W9_CWzIoUQIK6ovHhPDM22aDFumCoT1N7d2xrKysZ-nAr47fjZ6K992_bfvdscqEtDXoWWnAclfDFWl24xQnO7BQNFWmWu3aqz90qliP85QcNexzqtQdIEAy8PdJppBcq-NaFN9_pPtI9sWpv3m2ttzDXHfeyOKRWx--cOZTjC7rZSiDBxn7GrE-FIHR1XZG5UmMfCzYuJNQX0MS2sXu8V8-eKmQnqBkwx4VTmcPwGV_dK6g)](https://mermaid.live/edit#pako:eNqNk99PwjAQx_-VprxCwo8YtSQmOtgTCQaiL46H2l1HQ9fOrosSwv9u59aBzBH3svXu8737Xi87YKZjwARzqT_ZlhqLFqtIIfcwSfN8Bhy5sFBSqB3iQkrSm0wmQRhOUSfGtNSG9DjnLSi31OwU2LrU03B0G46nXdCVQkyrvEjB1IXC-9Hd8GbaBbULLYMVIeTkeTB4QI9JYiCh1qFnqYpf155cxtsrJWv4KEAxMGfxWuAzP6W9_CWzIoUQIK6ovHhPDM22aDFumCoT1N7d2xrKysZ-nAr47fjZ6K992_bfvdscqEtDXoWWnAclfDFWl24xQnO7BQNFWmWu3aqz90qliP85QcNexzqtQdIEAy8PdJppBcq-NaFN9_pPtI9sWpv3m2ttzDXHfeyOKRWx--cOZTjC7rZSiDBxn7GrE-FIHR1XZG5UmMfCzYuJNQX0MS2sXu8V8-eKmQnqBkwx4VTmcPwGV_dK6g)

**L1**

1. The EA is run by a network of Node operators to post the latest sequencer status to the `Aggregator` contract and relayed to the `ValidatorProxy` contract. The `Aggregator` contract then calls the `validate` function in the `ValidatorProxy` contract, which proxies the call to the `StarknetValidator` contract.

2. The `StarknetValidator` then calls the `sendMessageToL2` function on the `Starknet` contract. This message will contain instructions to call the `updateStatus(bool status, uint64 timestamp)` function in the `StarknetSequencerUptimeFeed` contract deployed on L2

3. The core `Starknet` contract then emits a new `LogMessageToL2` event to to signal that a new message needs to be sent from L1 to L2.

```jsx
event LogMessageToL2(
    address indexed fromAddress,
    uint256 indexed toAddress,
    uint256 indexed selector,
    uint256[] payload,
    uint256 nonce
);
```

4. The `Sequencer` will then pickup the `LogMessageToL2` event emitted above and forward the message to the target contract on L2.

**L2**

1. The Sequencer posts the message to the `starknet_sequencer_uptime_feed` contract and calls the `update_status` function to update the Sequencer status.

2. Consumers can then read from the `aggregator_proxy` contract, which fetches the latest round data from the `starknet_sequencer_uptime_feed` contract.

## Sequencer Downtime

### L1 → L2 Transactions

In the event that the Sequencer is down, messages will not be transmitted from L1 to L2 and **no L2 transactions are executed**. Instead messages will be enqueued in the Sequencer and only processed in the order they arrived later once the Sequencer comes back up. This means that as long as the message from the `StarknetValidator` on L1 is already enqueued in the Starknet Sequencer, the flag on the `starknet_sequencer_uptime_feed` on L2 will be guaranteed to be flipped prior to any subsequent transactions. This happens as the transaction flipping the flag on the uptime feed will get executed before transactions that were enqueued after it. This is further explained in the diagrams below.

**During Sequencer downtime**

- New `LogMessageToL2` events emitted are not picked up whilst the Sequencer is down.
- When the Sequencer is down, all L2 transactions sent from L1 are stuck in the pending queue, which lives in Starknet’s centralized Sequencer.
- **Tx1** contains Chainlink’s transaction to set the status of the Sequencer as being down on L2.
- **Tx2** is a transaction made by a consumer that is dependent on

[![](https://mermaid.ink/img/pako:eNo1jrEOwjAMRH8l8twFxsywMQBlzOImbhPRJMWNBajqvxNU1dO709PJC9jsCDT0Y35bj1zU5W6SqjdLNzBOXrX0EkqWWF0puZAGdRMS2qzH57DDcYPqQAOROGJwdXn51waKp0gGdEWH_DRg0lo9mRwWOrtQMoMuLNQASsntN9k9b84pYP0ngu5xnGn9ATwDPgo)](https://mermaid.live/edit#pako:eNo1jrEOwjAMRH8l8twFxsywMQBlzOImbhPRJMWNBajqvxNU1dO709PJC9jsCDT0Y35bj1zU5W6SqjdLNzBOXrX0EkqWWF0puZAGdRMS2qzH57DDcYPqQAOROGJwdXn51waKp0gGdEWH_DRg0lo9mRwWOrtQMoMuLNQASsntN9k9b84pYP0ngu5xnGn9ATwDPgo)

**After Sequencer comes back online**

- `LogMessageToL2` events are picked up and added to the pending queue.
- Transactions in the pending queue are processed chronologically so **Tx1** is processed before **Tx2.**
- As **Tx1** happens before **Tx2, Tx2** will read the status of the Sequencer as being down

### Bridge Fees

As of version v0.10.0, Starknet has begun introducing fees to send messages from L1 to L2. These fees are used to pay for the transaction
on L2. This has been done so that the Sequencer has an incentive to process L2 messags. Whilst today `v0.10.0` still processes messages sent
with `msg.value = 0`, Starknet has plans to charge a non 0 fee in the upcoming release. As the Emergency Protocol needs to send messages cross chain,
the protocol needs a way to estimate gas fees. Currently, the `StarkwareValidator` contract on L1 does the following to estimate the amount of required
gas.

1. Estimate gas fees by running the command below. The command is from Starkware's standard CLI.

```
starknet estimate_message_fee \
  --feeder_gateway_url=https://alpha4.starknet.io/feeder_gateway/ \
  --abi=contracts/starknet-artifacts/src/chainlink/cairo/emergency/SequencerUptimeFeed/sequencer_uptime_feed.cairo/sequencer_uptime_feed.json \
  --from_address ${FROM} \
  --address ${TO} \
  --function update_status \
  --inputs ${STATUS} ${TIMESTAMP}
```

In order to reliably ensure that cross chain messages are sent with sufficient gas, the estimate is multiplied by a buffer. This buffer
is arbitrarily chosen depending on what the Ops team determines to be a sufficient buffer.

2. Read the current L1 gas price from Chainlink's L1 gas price feed

3. Estimate gas as the product of the steps above.

```
gasFee = gasUnitsIncludingBuffer * l1GasPrice
```

### Layer2 Sequencer Health External Adapter

[Code](https://github.com/smartcontractkit/external-adapters-js/tree/develop/packages/sources/layer2-sequencer-health)

The emergency protocol requires an off chain component to tracks the health of the centralized Starkware sequencer. Today, this is made up by a DON (Decentralized Oracle Network) that triggers using OCR (Offchain Reporting). A new OCR round is initiated every 30s whereby each node in the DON checks the health of the Sequencer using the Layer2 Sequencer Health External Adapter. If the nodes in the DON determine that the Sequencer’s health has changed, they elect a new leader to write the updated result onto chain as shown in the diagram above.

**How the External Adapter Works**

Checking the Starkware Sequencer’s health is currently a two step process

1. Call the Sequencer directly to fetch the pending block’s details.
   1. Verify that a new block has been produced within 2 minutes by checking the pending block’s `parentHash`
   2. If the pending block’s `parentHash` has not changed, then check the length of the `transactions` field to see if it has increased since the last round
2. Send an empty transaction to a dummy contract at address `0x00000000000000000000000000000000000000000000000000000000000001`

   The EA sends the empty transaction using the StarknetJS library. This transaction tries to call the dummy contract’s `initialize` function with a `maxFee` of 0

   ```jsx
   const DUMMY_ADDRESS = '0x00000000000000000000000000000000000000000000000000000000000001'
   const DEFAULT_PRIVATE_KEY = '0x0000000000000000000000000000000000000000000000000000000000000001'
   const starkKeyPair = ec.genKeyPair(DEFAULT_PRIVATE_KEY)
   const starkKeyPub = ec.getStarkKey(starkKeyPair)
   const provider = config.starkwareConfig.provider
   const account = new Account(provider, DUMMY_ADDRESS, starkKeyPair)

   account.execute(
     {
       contractAddress: DUMMY_ADDRESS,
       entrypoint: 'initialize',
       calldata: [starkKeyPub, '0'],
     },
     undefined,
     { maxFee: '0' },
   )
   ```

3. As the above transaction is expected to fail, the EA will consider the Sequencer as healthy if it receives any of the expected error statuses
   1. `StarknetErrorCode.UNINITIALIZED_CONTRACT` if the dummy contract has not been initialized
   2. `StarknetErrorCode.OUT_OF_RANGE_FEE` if the dummy contract has been initialized by accident. As Starknet is a permissionless network, we cannot guarantee that a user deploys and initializes a contract at the dummy address. As a result, the EA will set the `maxFee` to 0 so that the transaction will fail with the `StarknetErrorCode.OUT_OF_RANGE_FEE` status code.

## Contracts

[StarknetValidator](https://github.com/smartcontractkit/chainlink-starknet/blob/develop/contracts/src/chainlink/solidity/emergency/StarkNetValidator.sol)
[SequencerUptimeFeed](https://github.com/smartcontractkit/chainlink-starknet/blob/develop/contracts/src/chainlink/cairo/emergency/SequencerUptimeFeed/sequencer_uptime_feed.cairo)
