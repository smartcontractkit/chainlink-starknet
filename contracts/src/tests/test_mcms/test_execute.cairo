use starknet::{contract_address_const, EthAddress};
use chainlink::libraries::mocks::mock_multisig_target::{
    IMockMultisigTarget, IMockMultisigTargetDispatcherTrait, IMockMultisigTargetDispatcher
};
use chainlink::mcms::{
    ExpiringRootAndOpCount, RootMetadata, Config, Signer, ManyChainMultiSig, Op,
    IManyChainMultiSigDispatcher, IManyChainMultiSigDispatcherTrait,
    IManyChainMultiSigSafeDispatcher, IManyChainMultiSigSafeDispatcherTrait, IManyChainMultiSig,
};
use snforge_std::{
    declare, ContractClassTrait, start_cheat_chain_id_global, spy_events,
    EventSpyAssertionsTrait, // Add for assertions on the EventSpy 
    cheatcodes::{events::{EventSpy}}, start_cheat_block_timestamp_global,
};
use chainlink::tests::test_mcms::utils::{setup_mcms_deploy_set_config_and_set_root};


#[test]
fn test_success() {
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

    let op1 = *ops.at(0);
    let op1_proof = *ops_proof.at(0);

    let target_address = op1.to;
    let target = IMockMultisigTargetDispatcher { contract_address: target_address };

    let (value, toggle) = target.read();
    assert(value == 0, 'should be 0');
    assert(toggle == false, 'should be false');

    mcms.execute(op1, op1_proof);

    spy
        .assert_emitted(
            @array![
                (
                    mcms_address,
                    ManyChainMultiSig::Event::OpExecuted(
                        ManyChainMultiSig::OpExecuted {
                            nonce: op1.nonce, to: op1.to, selector: op1.selector, data: op1.data
                        }
                    )
                )
            ]
        );

    assert(mcms.get_op_count() == 1, 'op count should be 1');

    let (new_value, _) = target.read();
    assert(new_value == 1234123, 'value should be updated');

    let op2 = *ops.at(1);
    let op2_proof = *ops_proof.at(1);

    mcms.execute(op2, op2_proof);

    spy
        .assert_emitted(
            @array![
                (
                    mcms_address,
                    ManyChainMultiSig::Event::OpExecuted(
                        ManyChainMultiSig::OpExecuted {
                            nonce: op2.nonce, to: op2.to, selector: op2.selector, data: op2.data
                        }
                    )
                )
            ]
        );

    assert(mcms.get_op_count() == 2, 'op count should be 2');

    let (_, new_toggle) = target.read();
    assert(new_toggle == true, 'toggled should be true');
}

#[test]
#[feature("safe_dispatcher")]
fn test_no_more_ops_to_execute() {
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

    let op1 = *ops.at(0);
    let op1_proof = *ops_proof.at(0);

    let op2 = *ops.at(1);
    let op2_proof = *ops_proof.at(1);

    mcms.execute(op1, op1_proof);
    mcms.execute(op2, op2_proof);

    let result = safe_mcms.execute(op1, op1_proof);
    match result {
        Result::Ok(_) => panic!("expect 'post-operation count reached'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'post-operation count reached', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_wrong_chain_id() {
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

    let op1 = *ops.at(0);
    let op1_proof = *ops_proof.at(0);

    start_cheat_chain_id_global(1231);
    let result = safe_mcms.execute(op1, op1_proof);

    match result {
        Result::Ok(_) => panic!("expect 'wrong chain id'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong chain id', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_wrong_multisig_address() {
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

    let mut op1 = *ops.at(0);
    op1.multisig = contract_address_const::<119922>();
    let op1_proof = *ops_proof.at(0);

    let result = safe_mcms.execute(op1, op1_proof);

    match result {
        Result::Ok(_) => panic!("expect 'wrong multisig address'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong multisig address', *panic_data.at(0));
        }
    }
}


#[test]
#[feature("safe_dispatcher")]
fn test_root_expired() {
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

    let op1 = *ops.at(0);
    let op1_proof = *ops_proof.at(0);

    start_cheat_block_timestamp_global(valid_until.into() + 1);
    let result = safe_mcms.execute(op1, op1_proof);

    match result {
        Result::Ok(_) => panic!("expect 'root has expired'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'root has expired', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_wrong_nonce() {
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

    let mut op1 = *ops.at(0);
    op1.nonce = 100;
    let op1_proof = *ops_proof.at(0);

    let result = safe_mcms.execute(op1, op1_proof);

    match result {
        Result::Ok(_) => panic!("expect 'wrong nonce'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'wrong nonce', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_proof_verification_failed() {
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

    let op1 = *ops.at(0);
    let bad_proof = array![0x12312312312321];

    let result = safe_mcms.execute(op1, bad_proof.span());

    match result {
        Result::Ok(_) => panic!("expect 'proof verification failed'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'proof verification failed', *panic_data.at(0));
        }
    }
}

