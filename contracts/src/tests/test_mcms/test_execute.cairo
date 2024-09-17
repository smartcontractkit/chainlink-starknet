use chainlink::mcms::{
    recover_eth_ecdsa, hash_pair, hash_op, hash_metadata, ExpiringRootAndOpCount, RootMetadata,
    Config, Signer, eip_191_message_hash, ManyChainMultiSig, Op,
    ManyChainMultiSig::{
        NewRoot, InternalFunctionsTrait, contract_state_for_testing,
        s_signersContractMemberStateTrait, s_expiring_root_and_op_countContractMemberStateTrait,
        s_root_metadataContractMemberStateTrait
    },
    IManyChainMultiSigDispatcher, IManyChainMultiSigDispatcherTrait,
    IManyChainMultiSigSafeDispatcher, IManyChainMultiSigSafeDispatcherTrait, IManyChainMultiSig,
    ManyChainMultiSig::{MAX_NUM_SIGNERS},
};
use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, start_cheat_caller_address,
    stop_cheat_caller_address, stop_cheat_caller_address_global, start_cheat_chain_id_global,
    spy_events, EventSpyAssertionsTrait, // Add for assertions on the EventSpy 
    test_address, // the contract being tested,
     start_cheat_chain_id,
    cheatcodes::{events::{EventSpy}}, start_cheat_block_timestamp_global,
    start_cheat_block_timestamp, start_cheat_account_contract_address_global,
    start_cheat_account_contract_address
};
use chainlink::tests::test_mcms::test_set_config::{setup_2_of_2_mcms_no_root, setup};
use chainlink::tests::test_mcms::test_set_root::{new_setup_2_of_2_mcms};
// 1. test no more operations to execute
// 2. test wrong chain id
// 3. test wrong multisig address
// 4. test root has expired 
// 5. test wrong nonce
// 6. test wrong proof
// 7. test contract call fails (it doesn't exist)
// 8. test success

// #[test]
// fn test_success() {
//     let (
//         mut spy,
//         mcms_address,
//         mcms,
//         safe_mcms,
//         config,
//         signer_addresses,
//         signer_groups,
//         group_quorums,
//         group_parents,
//         clear_root,
//         root,
//         valid_until,
//         metadata,
//         metadata_proof,
//         signatures,
//         ops
//     ) =
//         new_setup_2_of_2_mcms();

//     mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

//     // execute 
// }

