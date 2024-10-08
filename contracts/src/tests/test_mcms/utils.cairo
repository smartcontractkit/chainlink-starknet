use core::integer::{u512, u256_wide_mul};
use alexandria_bytes::{Bytes, BytesTrait};
use alexandria_encoding::sol_abi::sol_bytes::SolBytesTrait;
use alexandria_encoding::sol_abi::encode::SolAbiEncodeTrait;
use alexandria_math::u512_arithmetics;
use core::math::{u256_mul_mod_n, u256_div_mod_n};
use core::zeroable::{IsZeroResult, NonZero, zero_based};
use alexandria_math::u512_arithmetics::{u512_add, u512_sub, U512Intou256X2,};

use starknet::{
    ContractAddress, EthAddress, EthAddressIntoFelt252, EthAddressZeroable, contract_address_const,
    eth_signature::public_key_point_to_eth_address,
    secp256_trait::{
        Secp256Trait, Secp256PointTrait, recover_public_key, is_signature_entry_valid, Signature,
        signature_from_vrs
    },
    secp256k1::{Secp256k1Point, Secp256k1Impl}, SyscallResult, SyscallResultTrait
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
use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, start_cheat_caller_address,
    stop_cheat_caller_address, stop_cheat_caller_address_global, spy_events,
    EventSpyAssertionsTrait, // Add for assertions on the EventSpy 
    test_address, // the contract being tested,
     start_cheat_chain_id, start_cheat_chain_id_global,
    start_cheat_block_timestamp_global, cheatcodes::{events::{EventSpy}}
};

//
// setup helpers
//

// returns a length 32 array
// give (index, value) tuples to fill array with
// 
// ex: fill_array(array!(0, 1)) will fill the 0th index with value 1
// 
// assumes that values array is sorted in ascending order of the index
fn fill_array(mut values: Array<(u32, u8)>) -> Array<u8> {
    let mut result: Array<u8> = ArrayTrait::new();

    let mut maybe_next = values.pop_front();

    let mut i = 0;
    while i < 32_u32 {
        match maybe_next {
            Option::Some(next) => {
                let (next_index, next_value) = next;

                if i == next_index {
                    result.append(next_value);
                    maybe_next = values.pop_front();
                } else {
                    result.append(0);
                }
            },
            Option::None(_) => { result.append(0); },
        }

        i += 1;
    };

    result
}

fn ZERO_ARRAY() -> Array<u8> {
    // todo: replace with [0_u8; 32] in cairo 2.7.0+
    fill_array(array![])
}

#[derive(Copy, Drop, Serde)]
struct SignerMetadata {
    address: EthAddress,
    private_key: u256
}

fn setup_signers() -> (EthAddress, u256, EthAddress, u256, Array<SignerMetadata>) {
    let signer_address_1: EthAddress = (0x13Cf92228941e27eBce80634Eba36F992eCB148A)
        .try_into()
        .unwrap();
    let private_key_1: u256 = 0xf366414c9042ec470a8d92e43418cbf62caabc2bbc67e82bd530958e7fcaa688;

    let signer_address_2: EthAddress = (0xDa09C953823E1F60916E85faD44bF99A7DACa267)
        .try_into()
        .unwrap();
    let private_key_2: u256 = 0xed10b7a09dd0418ab35b752caffb70ee50bbe1fe25a2ebe8bba8363201d48527;

    let signer_metadata = array![
        SignerMetadata { address: signer_address_1, private_key: private_key_1 },
        SignerMetadata { address: signer_address_2, private_key: private_key_2 }
    ];
    (signer_address_1, private_key_1, signer_address_2, private_key_2, signer_metadata)
}

impl U512PartialOrd of PartialOrd<u512> {
    #[inline(always)]
    fn le(lhs: u512, rhs: u512) -> bool {
        !(rhs < lhs)
    }
    #[inline(always)]
    fn ge(lhs: u512, rhs: u512) -> bool {
        !(lhs < rhs)
    }
    fn lt(lhs: u512, rhs: u512) -> bool {
        if lhs.limb3 < rhs.limb3 {
            true
        } else if lhs.limb3 == rhs.limb3 {
            if lhs.limb2 < rhs.limb2 {
                true
            } else if lhs.limb2 == rhs.limb2 {
                if lhs.limb1 < rhs.limb1 {
                    true
                } else {
                    false
                }
            } else {
                false
            }
        } else {
            false
        }
    }
    #[inline(always)]
    fn gt(lhs: u512, rhs: u512) -> bool {
        rhs < lhs
    }
}

//  *** THIS IS CRYPTOGRAPHICALLY INSECURE ***
// the usage of a constant random target means that anyone can reverse engineer the private keys
// therefore this method is only meant to be used for tests
// arg z: message hash, arg e: private key
fn insecure_sign(z: u256, e: u256) -> (u256, u256, bool) {
    let z_u512: u512 = u256_wide_mul(z, (0x1).into());

    // order of the finite group
    let N = Secp256k1Impl::get_curve_size().try_into().unwrap();
    let n_u512: u512 = u256_wide_mul(N, (0x1).into());

    // "random" number k would be generated by a pseudo-random number generator
    // in secure applications it's important that k is random, or else the private key can 
    // be derived from r and s
    let k = 777;

    // random target
    let R = Secp256k1Impl::get_generator_point().mul(k).unwrap();
    let (r_x, r_y) = R.get_coordinates().unwrap();

    // calculate s = ( z + r*e ) / k (finite element operations)
    // where product = r*e and sum = z + r*re
    let product = u256_mul_mod_n(r_x, e, N.try_into().unwrap());
    let product_u512: u512 = u256_wide_mul(product, (0x1).into());

    // sum = z + product (finite element operations)
    // avoid u256 overflow by casting to u512
    let mut sum_u512 = u512_add(z_u512, product_u512);
    while sum_u512 >= n_u512 {
        sum_u512 = u512_sub(sum_u512, n_u512);
    };
    let sum: u256 = sum_u512.try_into().unwrap();

    let s = u256_div_mod_n(sum, k, N.try_into().unwrap()).unwrap();

    let v = 27 + (r_y % 2);

    let y_parity = v % 2 == 0;

    (r_x, s, y_parity)
}

// simplified logic will only work when len(ops) = 2
// metadata nodes is the last leaf so that len(leafs) = 3
fn merkle_root(leafs: Array<u256>) -> (u256, Span<u256>, Span<Span<u256>>) {
    let mut level: Array<u256> = ArrayTrait::new();

    let metadata = *leafs.at(leafs.len() - 1);
    let mut i = 0;

    // we assume metadata is last leaf so we exclude for now
    while i < leafs.len() - 1 {
        level.append(*leafs.at(i));
        i += 1;
    };

    let mut level = level.span(); // [leaf1, leaf2]

    let proof1 = array![*level.at(1), metadata];
    let proof2 = array![*level.at(0), metadata];

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
    let root = hash_pair(*level.at(0), metadata);

    (root, array![metadata_proof].span(), array![proof1.span(), proof2.span()].span())
}

fn set_root_args(
    mcms_address: ContractAddress,
    target_address: ContractAddress,
    mut signers_metadata: Array<SignerMetadata>,
    pre_op_count: u64,
    post_op_count: u64
) -> (u256, u32, RootMetadata, Span<u256>, Array<Signature>, Array<Op>, Span<Span<u256>>) {
    let mock_chain_id = 732;

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
        multisig: mcms_address,
        pre_op_count: pre_op_count,
        post_op_count: post_op_count,
        override_previous_root: false,
    };
    let valid_until = 9;

    let op1_hash = hash_op(op1);
    let op2_hash = hash_op(op2);

    let metadata_hash = hash_metadata(metadata, valid_until);

    // create merkle tree
    let (root, metadata_proof, ops_proof) = merkle_root(array![op1_hash, op2_hash, metadata_hash]);

    let encoded_root = BytesTrait::new_empty().encode(root).encode(valid_until);
    let message_hash = eip_191_message_hash(encoded_root.keccak());

    let mut signatures: Array<Signature> = ArrayTrait::new();

    while let Option::Some(signer_metadata) = signers_metadata
        .pop_front() {
            let (r, s, y_parity) = insecure_sign(message_hash, signer_metadata.private_key);
            let signature = Signature { r: r, s: s, y_parity: y_parity };
            let address = recover_eth_ecdsa(message_hash, signature).unwrap();

            // sanity check
            assert(address == signer_metadata.address, 'signer not equal');

            signatures.append(signature);
        };

    let ops = array![op1.clone(), op2.clone()];

    (root, valid_until, metadata, metadata_proof, signatures, ops, ops_proof)
}

//
// setup functions
//

fn setup_mcms_deploy() -> (
    ContractAddress, IManyChainMultiSigDispatcher, IManyChainMultiSigSafeDispatcher
) {
    let owner = contract_address_const::<213123123>();
    start_cheat_caller_address_global(owner);

    let calldata = array![owner.into()];

    let (mcms_address, _) = declare("ManyChainMultiSig").unwrap().deploy(@calldata).unwrap();

    (
        mcms_address,
        IManyChainMultiSigDispatcher { contract_address: mcms_address },
        IManyChainMultiSigSafeDispatcher { contract_address: mcms_address }
    )
}

fn setup_mcms_deploy_and_set_config_2_of_2(
    signer_address_1: EthAddress, signer_address_2: EthAddress
) -> (
    EventSpy,
    ContractAddress,
    IManyChainMultiSigDispatcher,
    IManyChainMultiSigSafeDispatcher,
    Config,
    Array<EthAddress>,
    Array<u8>,
    Array<u8>,
    Array<u8>,
    bool
) {
    let (mcms_address, mcms, safe_mcms) = setup_mcms_deploy();

    let signer_addresses: Array<EthAddress> = array![signer_address_1, signer_address_2];
    let signer_groups = array![0, 0];
    let group_quorums = fill_array(array![(0, 2)]);
    let group_parents = ZERO_ARRAY();

    let clear_root = false;

    let mut spy = spy_events();

    mcms
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    let config = mcms.get_config();

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
        clear_root
    )
}

// sets up root
fn setup_mcms_deploy_set_config_and_set_root() -> (
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
    Span<Span<u256>>
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

    let (root, valid_until, metadata, metadata_proof, signatures, ops, ops_proof) = set_root_args(
        mcms_address, target_address, signer_metadata, 0, 2
    );

    // mock chain id & timestamp
    start_cheat_chain_id_global(metadata.chain_id.try_into().unwrap());

    let mock_timestamp = 3;
    start_cheat_block_timestamp_global(mock_timestamp);

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
