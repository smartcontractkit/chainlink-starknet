use core::array::{SpanTrait, ArrayTrait};
use starknet::{ContractAddress, EthAddress, EthAddressZeroable};
use chainlink::mcms::{
    ManyChainMultiSig, IManyChainMultiSigDispatcher, IManyChainMultiSigSafeDispatcher,
    IManyChainMultiSigSafeDispatcherTrait, ManyChainMultiSig::{MAX_NUM_SIGNERS}
};

use snforge_std::{
    declare, ContractClassTrait, start_cheat_caller_address_global, stop_cheat_caller_address_global
};

// set_config tests

// 1. test if lena(signer_address) = 0 => revert 
// 2. test if lena(signer_address) > MAX_NUM_SIGNERS => revert

// 3. test if signer addresses and signer groups not same size

// 4. test if group_quorum and group_parents not same size

// 6. test if one of signer_group #'s is out of bounds NUM_GROUPS

// 7. test if group_parents[i] is greater than or equal to i (when not 0) there is revert
// 8. test if i is 0 and group_parents[i] != 0 and revert

// 9. test if there is a signer in a group where group_quorum[i] == 0 => revert
// 10. test if there are not enough signers to meet a quorum => revert
// 11. test if signer addresses are not in ascending order
// 12. successful => test without clearing root. test the state of storage variables and that event was emitted

fn setup() -> (ContractAddress, IManyChainMultiSigDispatcher, IManyChainMultiSigSafeDispatcher) {
    let calldata = array![];

    let (mcms_address, _) = declare("ManyChainMultiSig").unwrap().deploy(@calldata).unwrap();

    (
        mcms_address,
        IManyChainMultiSigDispatcher { contract_address: mcms_address },
        IManyChainMultiSigSafeDispatcher { contract_address: mcms_address }
    )
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_out_of_bound_signers() {
    // 1. test if len(signer_address) = 0 => revert 
    let (_, _, mcms_safe) = setup();

    let signer_addresses = array![];
    let signer_groups = array![];
    let group_quorums = array![];
    let group_parents = array![];
    let clear_root = false;

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'out of bound signers len'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'out of bound signers len', *panic_data.at(0));
        }
    }

    // 2. test if lena(signer_address) > MAX_NUM_SIGNERS => revert

    // todo: use fixed-size array in cairo >= 2.7.0
    // let signer_addresses = [EthAddressZeroable::zero(); 201];

    let mut signer_addresses = ArrayTrait::new();
    let mut i = 0;
    while i < 201_usize {
        signer_addresses.append(EthAddressZeroable::zero());
        i += 1;
    };

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'out of bound signers len'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'out of bound signers len', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_signer_groups_len_mismatch() {
    // 3. test if signer addresses and signer groups not same size
    let (_, _, mcms_safe) = setup();

    let signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![];
    let group_quorums = array![];
    let group_parents = array![];
    let clear_root = false;

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'signer groups len mismatch'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'signer groups len mismatch', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_group_quorums_parents_mismatch() {
    // 4. test if group_quorum and group_parents not length 32
    let (_, _, mcms_safe) = setup();

    let signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![0];
    let group_quorums = array![0];
    let group_parents = array![0];
    let clear_root = false;

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'group quorums/parents mismatch'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'group quorums/parents mismatch', *panic_data.at(0));
        }
    }

    // 5. test if group_quorum and group_parents not equal in length

    // todo: replace with [0_u8; 32] in cairo 2.7.0
    let mut group_quorums = ArrayTrait::new();
    let mut i = 0;
    while i < 32_usize {
        group_quorums.append(0);
        i += 1;
    };

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'group quorums/parents mismatch'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'group quorums/parents mismatch', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_signers_group_out_of_bounds() {
    // 6. test if one of signer_group #'s is out of bounds NUM_GROUPS
    let (_, _, mcms_safe) = setup();

    let signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![33];

    let mut group_quorums = ArrayTrait::new();
    let mut i = 0;
    while i < 32_usize {
        group_quorums.append(0);
        i += 1;
    };

    let mut group_parents = ArrayTrait::new();
    let mut i = 0;
    while i < 32_usize {
        group_parents.append(0);
        i += 1;
    };

    let clear_root = false;

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'out of bounds group'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'out of bounds group', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_group_tree_malformed() {
    // 7. test if group_parents[i] is greater than or equal to i (when not 0) there is revert
    let (_, _, mcms_safe) = setup();

    let signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![0];

    let mut group_quorums = array![
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
    ];

    let mut group_parents = array![
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        31
    ];

    let clear_root = false;

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'group tree malformed'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'group tree malformed', *panic_data.at(0));
        }
    }

    let mut group_parents = array![
        1,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
    ];

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'group tree malformed'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'group tree malformed', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_signer_in_disabled_group() {
    // 9. test if there is a signer in a group where group_quorum[i] == 0 => revert
    let (_, _, mcms_safe) = setup();

    let mut signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![0];
    let mut group_quorums = array![
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
    ];
    let mut group_parents = array![
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0,
        0
    ];
    let clear_root = false;

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'signer in disabled group'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'signer in disabled group', *panic_data.at(0));
        }
    }
}
// 10. test if there are not enough signers to meet a quorum => revert

