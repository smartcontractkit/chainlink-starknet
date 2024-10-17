use alexandria_data_structures::array_ext::ArrayTraitExt;
use alexandria_bytes::{Bytes, BytesTrait};
use alexandria_encoding::sol_abi::sol_bytes::SolBytesTrait;
use alexandria_encoding::sol_abi::encode::SolAbiEncodeTrait;
use core::array::{SpanTrait, ArrayTrait};
use starknet::{
    eth_signature::is_eth_signature_valid, ContractAddress, EthAddress, EthAddressIntoFelt252,
    EthAddressZeroable, contract_address_const, eth_signature::public_key_point_to_eth_address,
    secp256_trait::{
        Secp256Trait, Secp256PointTrait, recover_public_key, is_signature_entry_valid, Signature,
        signature_from_vrs
    },
    secp256k1::Secp256k1Point, SyscallResult, SyscallResultTrait
};
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
use chainlink::tests::test_mcms::utils::{
    insecure_sign, setup_signers, SignerMetadata, setup_mcms_deploy_and_set_config_2_of_2,
    setup_mcms_deploy_set_config_and_set_root, set_root_args, merkle_root
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

// sets up root but with wrong multisig address in metadata
fn setup_mcms_deploy_set_config_and_set_root_WRONG_MULTISIG() -> (
    EventSpy,
    ContractAddress,
    IManyChainMultiSigDispatcher,
    IManyChainMultiSigSafeDispatcher,
    Config,
    Array<EthAddress>,
    Array<u8>,
    Array<u8>,
    Array<u8>,
    bool, // clear root
    u256,
    u32,
    RootMetadata,
    Span<u256>,
    Array<Signature>,
    Array<Op>,
    Span<Span<u256>>,
) {
    let (signer_address_1, private_key_1, signer_address_2, private_key_2, signer_metadata) =
        setup_signers();

    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root
    ) =
        setup_mcms_deploy_and_set_config_2_of_2(
        signer_address_1, signer_address_2
    );

    let calldata = ArrayTrait::new();
    let mock_target_contract = declare("MockMultisigTarget").unwrap();
    let (target_address, _) = mock_target_contract.deploy(@calldata).unwrap();

    // mock chain id & timestamp
    let mock_chain_id = 732;
    start_cheat_chain_id_global(mock_chain_id);

    start_cheat_block_timestamp_global(3);

    // first operation
    let selector1 = selector!("set_value");
    let calldata1: Array<felt252> = array![1234123];
    let op1 = Op {
        chain_id: mock_chain_id.into(),
        multisig: mcms_address,
        nonce: 0,
        to: target_address,
        selector: selector1,
        data: calldata1.span()
    };

    // second operation
    // todo update
    let selector2 = selector!("flip_toggle");
    let calldata2 = array![];
    let op2 = Op {
        chain_id: mock_chain_id.into(),
        multisig: mcms_address,
        nonce: 1,
        to: target_address,
        selector: selector2,
        data: calldata2.span()
    };

    let metadata = RootMetadata {
        chain_id: mock_chain_id.into(),
        multisig: contract_address_const::<123123>(), // choose wrong multisig address
        pre_op_count: 0,
        post_op_count: 2,
        override_previous_root: false,
    };
    let valid_until = 9;

    let op1_hash = hash_op(op1);
    let op2_hash = hash_op(op2);

    let metadata_hash = hash_metadata(metadata);

    // create merkle tree
    let (root, metadata_proof, ops_proof) = merkle_root(array![op1_hash, op2_hash, metadata_hash]);

    let encoded_root = BytesTrait::new_empty().encode(root).encode(valid_until);
    let message_hash = eip_191_message_hash(encoded_root.keccak());

    let (r_1, s_1, y_parity_1) = insecure_sign(message_hash, private_key_1);
    let (r_2, s_2, y_parity_2) = insecure_sign(message_hash, private_key_2);

    let signature1 = Signature { r: r_1, s: s_1, y_parity: y_parity_1 };
    let signature2 = Signature { r: r_2, s: s_2, y_parity: y_parity_2 };

    let addr1 = recover_eth_ecdsa(message_hash, signature1).unwrap();
    let addr2 = recover_eth_ecdsa(message_hash, signature2).unwrap();

    assert(addr1 == signer_address_1, 'signer 1 not equal');
    assert(addr2 == signer_address_2, 'signer 2 not equal');

    let signatures = array![signature1, signature2];

    let ops = array![op1.clone(), op2.clone()];

    (
        spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    )
}

#[test]
fn test_set_root_success() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    let (actual_root, actual_valid_until) = mcms.get_root();

    assert(actual_root == root, 'root returned not equal');
    assert(actual_valid_until == valid_until, 'valid until not equal');

    let actual_root_metadata = mcms.get_root_metadata();
    assert(actual_root_metadata == metadata, 'root metadata not equal');

    spy
        .assert_emitted(
            @array![
                (
                    mcms_address,
                    ManyChainMultiSig::Event::NewRoot(
                        ManyChainMultiSig::NewRoot {
                            root: root, valid_until: valid_until, metadata: metadata
                        }
                    )
                )
            ]
        );
}
#[test]
#[feature("safe_dispatcher")]
fn test_set_root_hash_seen() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    mcms.set_root(root, valid_until, metadata, metadata_proof, signatures.clone());

    let result = safe_mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    match result {
        Result::Ok(_) => panic!("expect 'signed hash already seen'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'signed hash already seen', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_root_signatures_wrong_order() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    let unsorted_signatures = array![*signatures.at(1), *signatures.at(0)];

    let result = safe_mcms
        .set_root(root, valid_until, metadata, metadata_proof, unsorted_signatures);

    match result {
        Result::Ok(_) => panic!("expect 'signer address must increase'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'signer address must increase', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_root_signatures_invalid_signer() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    let invalid_signatures = array![
        signature_from_vrs(
            v: 27,
            r: u256 {
                high: 0x9e8df5d64fb9d2ae155b435ac37519fd, low: 0x6d1ffddf225cde953f6c97f8b3a7531d,
            },
            s: u256 {
                high: 0x21f13cc6eb1d14f6ebdc497411c57589, low: 0xea109b402fcde2cfe8f3d1b6d2bb8948
            },
        ),
        signature_from_vrs(
            v: 27,
            r: u256 {
                high: 0x7a5d64ca9b1814e15eb8df73b3c79ac2, low: 0x9b9080ac6546e07b1118b16e5651e19d,
            },
            s: u256 {
                high: 0x62794369d5bb5f5a02d2eb6805951990, low: 0xdfcd8563639dcc6668e235e1bea93303
            },
        )
    ];

    let result = safe_mcms
        .set_root(root, valid_until, metadata, metadata_proof, invalid_signatures);

    match result {
        Result::Ok(_) => panic!("expect 'invalid signer'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'invalid signer', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_insufficient_signers() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    let missing_1_signature = array![*signatures.at(0)];

    let result = safe_mcms
        .set_root(root, valid_until, metadata, metadata_proof, missing_1_signature);

    match result {
        Result::Ok(_) => panic!("expect 'insufficient signers'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'insufficient signers', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_valid_until_expired() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    // cheat block timestamp
    start_cheat_block_timestamp_global(valid_until.into() + 1);

    let result = safe_mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    match result {
        Result::Ok(_) => panic!("expect 'valid until has passed'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'valid until has passed', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_invalid_metadata_proof() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    let invalid_metadata_proof = array![*metadata_proof.at(0), *metadata_proof.at(0)];

    let result = safe_mcms
        .set_root(root, valid_until, metadata, invalid_metadata_proof.span(), signatures);

    match result {
        Result::Ok(_) => panic!("expect 'proof verification failed'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'proof verification failed', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_invalid_chain_id() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    start_cheat_chain_id_global(123123);

    let result = safe_mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    match result {
        Result::Ok(_) => panic!("expect 'wrong chain id'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong chain id', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_invalid_multisig_address() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root_WRONG_MULTISIG();

    let result = safe_mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    match result {
        Result::Ok(_) => panic!("expect 'wrong multisig address'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong multisig address', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_pending_ops_remain() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    // first time passes
    mcms.set_root(root, valid_until, metadata, metadata_proof, signatures.clone());

    // sign a different set of operations with same signers
    let (signer_address_1, private_key_1, signer_address_2, private_key_2, signer_metadata) =
        setup_signers();
    let (root, valid_until, metadata, metadata_proof, signatures, ops, ops_proof) = set_root_args(
        mcms_address, contract_address_const::<123123>(), signer_metadata, 0, 2
    );

    // second time fails
    let result = safe_mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    match result {
        Result::Ok(_) => panic!("expect 'pending operations remain'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'pending operations remain', *panic_data.at(0));
        }
    }
}


#[test]
#[feature("safe_dispatcher")]
fn test_wrong_pre_op_count() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        _
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    // sign a different set of operations with same signers
    let (signer_address_1, private_key_1, signer_address_2, private_key_2, signer_metadata) =
        setup_signers();
    let wrong_pre_op_count = 1;
    let (root, valid_until, metadata, metadata_proof, signatures, _, _) = set_root_args(
        mcms_address,
        contract_address_const::<123123>(),
        signer_metadata,
        wrong_pre_op_count,
        wrong_pre_op_count + 2
    );

    // first time passes
    let result = safe_mcms
        .set_root(root, valid_until, metadata, metadata_proof, signatures.clone());

    match result {
        Result::Ok(_) => panic!("expect 'wrong pre-operation count'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong pre-operation count', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_wrong_post_ops_count() {
    let (
        mut spy,
        mcms_address,
        mcms,
        safe_mcms,
        config,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root,
        root,
        valid_until,
        metadata,
        metadata_proof,
        signatures,
        ops,
        ops_proof
    ) =
        setup_mcms_deploy_set_config_and_set_root();

    mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    // sign a different set of operations with same signers

    let (signer_address_1, private_key_1, signer_address_2, private_key_2, signer_metadata) =
        setup_signers();

    let op1 = *ops.at(0);
    let op1_proof = *ops_proof.at(0);

    let op2 = *ops.at(1);
    let op2_proof = *ops_proof.at(1);

    mcms.execute(op1, op1_proof);
    mcms.execute(op2, op2_proof);

    let (root, valid_until, metadata, metadata_proof, signatures, ops, ops_proof) = set_root_args(
        mcms_address,
        contract_address_const::<123123>(),
        signer_metadata,
        2, // correct pre-op count
        1 // wrong post-op count
    );

    let result = safe_mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);
    match result {
        Result::Ok(_) => panic!("expect 'wrong post-operation count'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong post-operation count', *panic_data.at(0));
        }
    }
}
