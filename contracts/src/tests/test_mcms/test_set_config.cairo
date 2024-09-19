use core::array::{SpanTrait, ArrayTrait};
use starknet::{
    ContractAddress, EthAddress, Felt252TryIntoEthAddress, EthAddressIntoFelt252,
    EthAddressZeroable, contract_address_const
};
use chainlink::mcms::{
    ExpiringRootAndOpCount, RootMetadata, Config, Signer, ManyChainMultiSig,
    ManyChainMultiSig::{
        InternalFunctionsTrait, contract_state_for_testing, s_signersContractMemberStateTrait,
        s_expiring_root_and_op_countContractMemberStateTrait,
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
     start_cheat_chain_id,
    cheatcodes::{events::{EventSpy}}
};
use chainlink::tests::test_mcms::utils::{
    setup_mcms_deploy, setup_mcms_deploy_and_set_config_2_of_2, ZERO_ARRAY, fill_array
};

#[test]
#[feature("safe_dispatcher")]
fn test_not_owner() {
    let (_, _, mcms_safe) = setup_mcms_deploy();

    let signer_addresses = array![];
    let signer_groups = array![];
    let group_quorums = array![];
    let group_parents = array![];
    let clear_root = false;

    // so that caller is not owner
    start_cheat_caller_address_global(contract_address_const::<123123>());

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'Caller is not the owner'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'Caller is not the owner', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_out_of_bound_signers() {
    // 1. test if len(signer_address) = 0 => revert 
    let (_, _, mcms_safe) = setup_mcms_deploy();

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
    let (_, _, mcms_safe) = setup_mcms_deploy();

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
    let (_, _, mcms_safe) = setup_mcms_deploy();

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
        Result::Ok(_) => panic!("expect 'wrong group quorums/parents len'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong group quorums/parents len', *panic_data.at(0));
        }
    }

    // 5. test if group_quorum and group_parents not equal in length

    let mut group_quorums = ZERO_ARRAY();

    let result = mcms_safe
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    match result {
        Result::Ok(_) => panic!("expect 'wrong group quorums/parents len'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong group quorums/parents len', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_set_config_signers_group_out_of_bounds() {
    // 6. test if one of signer_group #'s is out of bounds NUM_GROUPS
    let (_, _, mcms_safe) = setup_mcms_deploy();

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
    let (_, _, mcms_safe) = setup_mcms_deploy();

    let signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![0];

    let mut group_quorums = ZERO_ARRAY();
    let mut group_parents = fill_array(array![(31, 31)]);

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
    let (_, _, mcms_safe) = setup_mcms_deploy();

    let mut signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![0];
    let mut group_quorums = ZERO_ARRAY();
    let mut group_parents = ZERO_ARRAY();
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
#[test]
#[feature("safe_dispatcher")]
fn test_set_config_quorum_impossible() {
    let (_, _, mcms_safe) = setup_mcms_deploy();

    let mut signer_addresses = array![EthAddressZeroable::zero()];
    let signer_groups = array![0];
    let mut group_quorums = fill_array(array![(0, 2)]);
    let mut group_parents = ZERO_ARRAY();
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
        Result::Ok(_) => panic!("expect 'quorum impossible'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'quorum impossible', *panic_data.at(0));
        }
    }
}

// 11. test if signer addresses are not in ascending order
#[test]
#[feature("safe_dispatcher")]
fn test_set_config_signer_addresses_not_sorted() {
    let (_, _, mcms_safe) = setup_mcms_deploy();

    let mut signer_addresses: Array<EthAddress> = array![
        // 0x1 address
        u256 { high: 0, low: 1 }.into(), EthAddressZeroable::zero()
    ];
    let signer_groups = array![0, 0];
    let mut group_quorums = fill_array(array![(0, 2)]);
    let mut group_parents = ZERO_ARRAY();
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
        Result::Ok(_) => panic!("expect 'signer addresses not sorted'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'signer addresses not sorted', *panic_data.at(0));
        }
    }
}

// test success, root not cleared, event emitted
// 12. successful => test without clearing root. test the state of storage variables and that event was emitted
//
//                    ┌──────┐
//                 ┌─►│2-of-2│
//                 │  └──────┘        
//                 │        ▲         
//                 │        │         
//              ┌──┴───┐ ┌──┴───┐ 
//              signer 1 signer 2 
//              └──────┘ └──────┘ 
#[test]
fn test_set_config_success_dont_clear_root() {
    let signer_address_1: EthAddress = (0x141).try_into().unwrap();
    let signer_address_2: EthAddress = (0x2412).try_into().unwrap();
    let (
        mut spy,
        mcms_address,
        mcms,
        _,
        _,
        signer_addresses,
        signer_groups,
        group_quorums,
        group_parents,
        clear_root
    ) =
        setup_mcms_deploy_and_set_config_2_of_2(
        signer_address_1, signer_address_2
    );

    let expected_signer_1 = Signer { address: signer_address_1, index: 0, group: 0 };
    let expected_signer_2 = Signer { address: signer_address_2, index: 1, group: 0 };

    let expected_config = Config {
        signers: array![expected_signer_1, expected_signer_2].span(),
        group_quorums: group_quorums.span(),
        group_parents: group_parents.span(),
    };

    spy
        .assert_emitted(
            @array![
                (
                    mcms_address,
                    ManyChainMultiSig::Event::ConfigSet(
                        ManyChainMultiSig::ConfigSet {
                            config: expected_config, is_root_cleared: false
                        }
                    )
                )
            ]
        );
    let config = mcms.get_config();
    assert(config == expected_config, 'config should be same');

    // mock the contract state
    let test_address = test_address();
    start_cheat_caller_address(test_address, contract_address_const::<777>());

    // test internal function state
    let mut state = contract_state_for_testing();
    ManyChainMultiSig::constructor(ref state);
    state
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    let signer_1 = state.get_signer_by_address(signer_address_1);
    let signer_2 = state.get_signer_by_address(signer_address_2);

    assert(signer_1 == expected_signer_1, 'signer 1 not equal');
    assert(signer_2 == expected_signer_2, 'signer 2 not equal');

    //  test second set_config
    let new_signer_address_1: EthAddress = u256 { high: 0, low: 3 }.into();
    let new_signer_address_2: EthAddress = u256 { high: 0, low: 4 }.into();
    let new_signer_addresses = array![new_signer_address_1, new_signer_address_2];

    mcms
        .set_config(
            new_signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    let new_config = mcms.get_config();

    let new_expected_signer_1 = Signer { address: new_signer_address_1, index: 0, group: 0 };
    let new_expected_signer_2 = Signer { address: new_signer_address_2, index: 1, group: 0 };

    let new_expected_config = Config {
        signers: array![new_expected_signer_1, new_expected_signer_2].span(),
        group_quorums: group_quorums.span(),
        group_parents: group_parents.span(),
    };

    assert(new_config == new_expected_config, 'new config should be same');

    state
        .set_config(
            new_signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    let new_signer_1 = state.get_signer_by_address(new_signer_address_1);
    let new_signer_2 = state.get_signer_by_address(new_signer_address_2);

    assert(new_signer_1 == new_expected_signer_1, 'new signer 1 not equal');
    assert(new_signer_2 == new_expected_signer_2, 'new signer 2 not equal');

    // test old signers were reset
    let old_signer_1 = state.get_signer_by_address(signer_address_1);
    let old_signer_2 = state.get_signer_by_address(signer_address_2);
    assert(old_signer_1.address == EthAddressZeroable::zero(), 'old signer 1 should be reset');
    assert(old_signer_2.address == EthAddressZeroable::zero(), 'old signer 1 should be reset');
}


// test that the config was reset 
#[test]
fn test_set_config_success_and_clear_root() {
    // mock the contract state
    let test_address = test_address();
    let mock_chain_id = 990;
    start_cheat_caller_address(test_address, contract_address_const::<777>());
    start_cheat_chain_id(test_address, mock_chain_id);

    let mut state = contract_state_for_testing();
    ManyChainMultiSig::constructor(ref state);

    // initialize s_expiring_root_and_op_count & s_root_metadata
    state
        .s_expiring_root_and_op_count
        .write(
            ExpiringRootAndOpCount {
                root: u256 { high: 777, low: 777 }, valid_until: 102934894, op_count: 134
            }
        );

    state
        .s_root_metadata
        .write(
            RootMetadata {
                chain_id: 123123,
                multisig: contract_address_const::<111>(),
                pre_op_count: 20,
                post_op_count: 155,
                override_previous_root: false
            }
        );

    let signer_address_1: EthAddress = u256 { high: 0, low: 1 }.into();
    let signer_address_2: EthAddress = u256 { high: 0, low: 2 }.into();
    let signer_addresses: Array<EthAddress> = array![signer_address_1, signer_address_2];
    let signer_groups = array![0, 0];
    let group_quorums = fill_array(array![(0, 2)]);
    let group_parents = ZERO_ARRAY();
    let clear_root = true;

    state
        .set_config(
            signer_addresses.span(),
            signer_groups.span(),
            group_quorums.span(),
            group_parents.span(),
            clear_root
        );

    let expected_s_expiring_root_and_op_count = ExpiringRootAndOpCount {
        root: u256 { high: 0, low: 0 }, valid_until: 0, op_count: 134
    };
    let s_expiring_root_and_op_count = state.s_expiring_root_and_op_count.read();
    assert!(
        s_expiring_root_and_op_count == expected_s_expiring_root_and_op_count,
        "s_expiring_root_and_op_count not equal"
    );

    let expected_s_root_metadata = RootMetadata {
        chain_id: mock_chain_id.into(),
        multisig: test_address,
        pre_op_count: 134,
        post_op_count: 134,
        override_previous_root: true
    };
    let s_root_metadata = state.s_root_metadata.read();
    assert(expected_s_root_metadata == s_root_metadata, 's_root_metadata not equal');
}
