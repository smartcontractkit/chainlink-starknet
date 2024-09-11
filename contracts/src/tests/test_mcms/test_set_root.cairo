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
use chainlink::tests::test_mcms::test_set_config::setup_2_of_2_mcms_no_root;
use chainlink::tests::test_mcms::utils::{insecure_sign};

use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, start_cheat_caller_address,
    stop_cheat_caller_address, stop_cheat_caller_address_global, start_cheat_chain_id_global,
    spy_events, EventSpyAssertionsTrait, // Add for assertions on the EventSpy 
    test_address, // the contract being tested,
     start_cheat_chain_id,
    cheatcodes::{events::{EventSpy}}, start_cheat_block_timestamp_global
};

// test to add root

// 1. set up mcms contract
// 2. set up a dummy contract (like mock multisig target or a new contract)
// 3. propose Op struct (of 2 ) and metadata 
// 4. generate a root 
// 5. abi_encode(root and valid_until)
// 6. create message hash
// 7. sign the message hash (how to do?) -- i'll do it via typescript and just input the hash here
//  leaves = [...txs.map(txCoder), ...metadata.map(metadataCoder)] (metadata is the last leaf) <-- sort pairs and sort leaves
// https://www.npmjs.com/package/merkletreejs
// the proof does not include the root or the leaf

// metadata should be the last leaf
// will only work with an even number of ops
fn merkle_root(leafs: Array<u256>) -> (u256, Span<u256>) {
    let mut level: Array<u256> = ArrayTrait::new();
    let odd = leafs.len() % 2 == 1;
    let mut metadata: Option<u256> = Option::None;

    if odd {
        metadata = Option::Some(*leafs.at(leafs.len() - 1));
        let mut i = 0;

        // we assume metadata is last leaf so we exclude for now
        while i < leafs.len() - 1 {
            level.append(*leafs.at(i));
            i += 1;
        }
    }

    let mut level = level.span();

    // level length is always even (except when it's 1)
    while level
        .len() > 1 {
            let mut i = 0;
            let mut new_level: Array<u256> = ArrayTrait::new();
            while i < level
                .len() {
                    new_level.append(hash_pair(*(level.at(i)), *level.at(i + 1)));
                    i += 2
                };
            level = new_level.span();
        };

    let mut metadata_proof = *level.at(0);

    // based on merkletree.js lib we use, the odd leaf out is not hashed until the very end
    let root = match metadata {
        Option::Some(m) => hash_pair(*level.at(0), m),
        Option::None => (*level.at(0)),
    };

    (root, array![metadata_proof].span())
}

// sets up root
fn setup_2_of_2_mcms() -> (
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
    Array<Op>
) {
    let signer_address_1: EthAddress = (0x13Cf92228941e27eBce80634Eba36F992eCB148A)
        .try_into()
        .unwrap();

    let signer_address_2: EthAddress = (0xDa09C953823E1F60916E85faD44bF99A7DACa267)
        .try_into()
        .unwrap();
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
        setup_2_of_2_mcms_no_root(
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
    let selector2 = selector!("toggle");
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
        multisig: mcms_address,
        pre_op_count: 0,
        post_op_count: 2,
        override_previous_root: false,
    };
    let valid_until = 9;

    let op1_hash = hash_op(op1);
    let op2_hash = hash_op(op2);

    let metadata_hash = hash_metadata(metadata, valid_until);

    // create merkle tree
    let (root, metadata_proof) = merkle_root(array![op1_hash, op2_hash, metadata_hash]);

    let encoded_root = BytesTrait::new_empty().encode(root).encode(valid_until);

    // abi.encode(root, valid_until) 
    // output: `sign this msg hash: 80900408832728789228986457074438699885219690940027717206734875544162448258660``
    println!("sign this msg hash: {:?}", encoded_root.keccak());

    // run `yarn ts-node test/mcms-generate-signature.ts <msg_hash>

    // signature 1: 
    // r: 
    //   high: 0x547a06aecfe75c9bca6d20fb0571ca37, low: 0x1ba523139c3c00145f74f997b224c0ae
    // s:
    //   high: 0x00696f026d9371e3009588d81391174f, low: 0xb85561559edf82cbebf145a8d77681a9
    // v:
    //   27
    // signature 2: 
    // r: 
    //     high: 0x73b0625d2623d18af2cbc023511386c7, low: 0x9d04510e392300e7a69dea2b89d83dd6
    // s:
    //     high: 0x55ce476d61b1cecdad20854760996cc7, low: 0x12754fc6e8ac845779e3d22d5a443957
    // v:
    //     28

    let signature1 = signature_from_vrs(
        v: 27,
        r: u256 {
            high: 0x547a06aecfe75c9bca6d20fb0571ca37, low: 0x1ba523139c3c00145f74f997b224c0ae,
        },
        s: u256 {
            high: 0x00696f026d9371e3009588d81391174f, low: 0xb85561559edf82cbebf145a8d77681a9
        },
    );

    let signature2 = signature_from_vrs(
        v: 28,
        r: u256 {
            high: 0x73b0625d2623d18af2cbc023511386c7, low: 0x9d04510e392300e7a69dea2b89d83dd6,
        },
        s: u256 {
            high: 0x55ce476d61b1cecdad20854760996cc7, low: 0x12754fc6e8ac845779e3d22d5a443957
        },
    );

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
        ops
    )
}


// sets up root
fn new_setup_2_of_2_mcms() -> (
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
    Array<Op>
) {
    let signer_address_1: EthAddress = (0x13Cf92228941e27eBce80634Eba36F992eCB148A)
        .try_into()
        .unwrap();
    let private_key_1: u256 = 0xf366414c9042ec470a8d92e43418cbf62caabc2bbc67e82bd530958e7fcaa688;

    let signer_address_2: EthAddress = (0xDa09C953823E1F60916E85faD44bF99A7DACa267)
        .try_into()
        .unwrap();
    let private_key_2: u256 = 0xed10b7a09dd0418ab35b752caffb70ee50bbe1fe25a2ebe8bba8363201d48527;

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
        setup_2_of_2_mcms_no_root(
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
    let selector2 = selector!("toggle");
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
        multisig: mcms_address,
        pre_op_count: 0,
        post_op_count: 2,
        override_previous_root: false,
    };
    let valid_until = 9;

    let op1_hash = hash_op(op1);
    let op2_hash = hash_op(op2);

    let metadata_hash = hash_metadata(metadata, valid_until);

    // create merkle tree
    let (root, metadata_proof) = merkle_root(array![op1_hash, op2_hash, metadata_hash]);

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
        ops
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
        ops
    ) =
        new_setup_2_of_2_mcms();

    mcms.set_root(root, valid_until, metadata, metadata_proof, signatures);

    let (actual_root, actual_valid_until) = mcms.get_root();

    assert(actual_root == root, 'root returned not equal');
    assert(actual_valid_until == valid_until, 'valid until not equal');

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
        ops
    ) =
        new_setup_2_of_2_mcms();

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
        ops
    ) =
        new_setup_2_of_2_mcms();

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
        ops
    ) =
        new_setup_2_of_2_mcms();

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


