use starknet::{ContractAddress, contract_address_const};
use chainlink::{
    access_control::rbac_timelock::{
        RBACTimelock, IRBACTimelock, IRBACTimelockDispatcher, IRBACTimelockDispatcherTrait,
        IRBACTimelockSafeDispatcher, IRBACTimelockSafeDispatcherTrait,
        RBACTimelock::{ADMIN_ROLE, PROPOSER_ROLE, EXECUTOR_ROLE, CANCELLER_ROLE, BYPASSER_ROLE},
        Call
    },
    libraries::mocks::mock_multisig_target::{
        IMockMultisigTarget, IMockMultisigTargetDispatcherTrait, IMockMultisigTargetDispatcher
    }
};

use openzeppelin::{
    introspection::interface::{ISRC5, ISRC5Dispatcher, ISRC5DispatcherTrait, ISRC5_ID},
    access::accesscontrol::{
        interface::{
            IACCESSCONTROL_ID, IAccessControl, IAccessControlDispatcher,
            IAccessControlDispatcherTrait
        },
        accesscontrol::AccessControlComponent::Errors
    },
    token::{erc1155::interface::{IERC1155_RECEIVER_ID}, erc721::interface::{IERC721_RECEIVER_ID}}
};
use snforge_std::{
    declare, ContractClassTrait, spy_events, EventSpyAssertionsTrait,
    start_cheat_caller_address_global, start_cheat_block_timestamp_global
};
// 1. test supports access controller, erc1155 receiver, and erc721 receiver
// 2. test has_roles after constructor is called + min delay + event emitted
// 3. test schedule_batch, cancel, execute_batch, update_delay, bypasser_execute_batch , block, unblock with wrong roles
// 4. test schedule with too small of a delay
// 5. test schedule with blocked selector
// 6. test schedule_batch + get operation is true

// 7. test can't cancel operation that isn't scheduled 
// 8. test can't cancel operation that was already executed
// test cancel is successful

// 9. test execute call that isn't ready yet
// 10. test execute call which predecessor is not executed
// 11. test execute call successful + assert it's done

// 12. test update delay is true. 
// 13. test scheduled batch and then updated delay changes the return of is_operation_ready
// 14. test bypasser execute batch works 
// 15. test bypasser execute batch fails
// 16. test block function selector once and then twice
// 17. test unblock fx selector unsuccessful and then successful
// 18. test get block fx selector count
// 19. test get blocked fx selector index throughout a bunch of add and removals
// 20. test is_operation, pending, ready, done throughout life cycle of id
// 21. test block and unblock the max felt size

fn deploy_args() -> (
    u256, ContractAddress, ContractAddress, ContractAddress, ContractAddress, ContractAddress
) {
    let min_delay: u256 = 0x9;
    let admin = contract_address_const::<1>();
    let proposer = contract_address_const::<2>();
    let executor = contract_address_const::<3>();
    let canceller = contract_address_const::<4>();
    let bypasser = contract_address_const::<5>();
    (min_delay, admin, proposer, executor, canceller, bypasser)
}

fn setup_mock_target() -> (ContractAddress, IMockMultisigTargetDispatcher) {
    let calldata = ArrayTrait::new();
    let mock_target_contract = declare("MockMultisigTarget").unwrap();
    let (target_address, _) = mock_target_contract.deploy(@calldata).unwrap();
    (target_address, IMockMultisigTargetDispatcher { contract_address: target_address })
}

fn setup_timelock() -> (ContractAddress, IRBACTimelockDispatcher, IRBACTimelockSafeDispatcher) {
    let (min_delay, admin, proposer, executor, canceller, bypasser) = deploy_args();
    let proposers = array![proposer];
    let executors = array![executor];
    let cancellers = array![canceller];
    let bypassers = array![bypasser];

    let mut calldata = ArrayTrait::new();
    Serde::serialize(@min_delay, ref calldata);
    Serde::serialize(@admin, ref calldata);
    Serde::serialize(@proposers, ref calldata);
    Serde::serialize(@executors, ref calldata);
    Serde::serialize(@cancellers, ref calldata);
    Serde::serialize(@bypassers, ref calldata);

    let (timelock_address, _) = declare("RBACTimelock").unwrap().deploy(@calldata).unwrap();

    (
        timelock_address,
        IRBACTimelockDispatcher { contract_address: timelock_address },
        IRBACTimelockSafeDispatcher { contract_address: timelock_address }
    )
}

#[test]
fn test_supports_interfaces() {
    let (timelock_address, _, _) = setup_timelock();

    let timelock = ISRC5Dispatcher { contract_address: timelock_address };

    assert(timelock.supports_interface(ISRC5_ID), 'supports ISRC5');

    assert(timelock.supports_interface(IACCESSCONTROL_ID), 'supports IACCESSCONTROL_ID');

    assert(timelock.supports_interface(IERC1155_RECEIVER_ID), 'supports IERC1155_RECEIVER_ID');

    assert(timelock.supports_interface(IERC721_RECEIVER_ID), 'supports IERC721_RECEIVER_ID');

    assert(!timelock.supports_interface(0x0123123123), 'does not support random one');
}

#[test]
fn test_roles() {
    let (_, admin, proposer, executor, canceller, bypasser) = deploy_args();
    let (timelock_address, _, _) = setup_timelock();

    let timelock = IAccessControlDispatcher { contract_address: timelock_address };

    // admin role controls rest of roles
    assert(
        timelock.get_role_admin(ADMIN_ROLE) == ADMIN_ROLE
            && timelock.get_role_admin(PROPOSER_ROLE) == ADMIN_ROLE
            && timelock.get_role_admin(EXECUTOR_ROLE) == ADMIN_ROLE
            && timelock.get_role_admin(CANCELLER_ROLE) == ADMIN_ROLE
            && timelock.get_role_admin(BYPASSER_ROLE) == ADMIN_ROLE,
        'admin role controls all roles'
    );

    // admin address
    assert(timelock.has_role(ADMIN_ROLE, admin), 'is admin');
    assert(timelock.has_role(PROPOSER_ROLE, proposer), 'is proposer');
    assert(timelock.has_role(EXECUTOR_ROLE, executor), 'is executor');
    assert(timelock.has_role(CANCELLER_ROLE, canceller), 'is canceller');
    assert(timelock.has_role(BYPASSER_ROLE, bypasser), 'is bypasser');
}

#[test]
fn test_deploy() {
    let mut spy = spy_events();

    let (min_delay, _, _, _, _, _) = deploy_args();
    let (timelock_address, timelock, _) = setup_timelock();

    assert(timelock.get_min_delay() == min_delay, 'min delay correct');
    spy
        .assert_emitted(
            @array![
                (
                    timelock_address,
                    RBACTimelock::Event::MinDelayChange(
                        RBACTimelock::MinDelayChange { old_duration: 0, new_duration: min_delay }
                    )
                )
            ]
        );
}

#[test]
#[feature("safe_dispatcher")]
fn test_funcs_fail_wrong_role() {
    let (min_delay, _, _, _, _, _) = deploy_args();
    let (_, _, safe_timelock) = setup_timelock();

    start_cheat_caller_address_global(contract_address_const::<123123>());

    expect_missing_role(safe_timelock.schedule_batch(array![].span(), 0, 0, 0));
    expect_missing_role(safe_timelock.cancel(0));
    expect_missing_role(safe_timelock.execute_batch(array![].span(), 0, 0));
    expect_missing_role(safe_timelock.update_delay(0));
    expect_missing_role(safe_timelock.bypasser_execute_batch(array![].span()));
    expect_missing_role(safe_timelock.block_function_selector(0x0));
    expect_missing_role(safe_timelock.unblock_function_selector(0x0));
}


fn expect_missing_role(result: Result<(), Array<felt252>>) {
    match result {
        Result::Ok(_) => panic!("expect 'Caller is missing role'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == Errors::MISSING_ROLE, *panic_data.at(0));
        }
    }
}

// schedule

// 4. test schedule with too small of a delay [x]
// 5. test schedule with blocked selector [x]
// 6. test schedule_batch + get operation is true [x]
// 7 . test schedule fails when id scheduled already [x]

#[test]
#[feature("safe_dispatcher")]
fn test_schedule_delay_too_small() {
    let (min_delay, _, proposer, _, _, _) = deploy_args();
    let (_, _, safe_timelock) = setup_timelock();

    start_cheat_caller_address_global(proposer);

    let result = safe_timelock.schedule_batch(array![].span(), 0, 0, min_delay - 1);
    match result {
        Result::Ok(_) => panic!("expect 'insufficient delay'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'insufficient delay', *panic_data.at(0));
        }
    }
}

#[test]
fn test_schedule_success() {
    let (min_delay, _, proposer, _, _, _) = deploy_args();
    let (timelock_address, timelock, _) = setup_timelock();

    let mock_time = 3;
    let mock_ready_time = mock_time + min_delay.try_into().unwrap();

    start_cheat_caller_address_global(proposer);
    start_cheat_block_timestamp_global(mock_time);

    let call = Call {
        target: contract_address_const::<100>(),
        selector: selector!("doesnt_exist"),
        data: array![0x123].span()
    };
    let calls = array![call].span();
    let predecessor = 0;
    let salt = 1;

    let id = timelock.hash_operation_batch(calls, predecessor, salt);

    assert(!timelock.is_operation(id), 'should not exist');

    let mut spy = spy_events();

    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    spy
        .assert_emitted(
            @array![
                (
                    timelock_address,
                    RBACTimelock::Event::CallScheduled(
                        RBACTimelock::CallScheduled {
                            id: id,
                            index: 0,
                            target: call.target,
                            selector: call.selector,
                            data: call.data,
                            predecessor: predecessor,
                            salt: salt,
                            delay: min_delay
                        }
                    )
                )
            ]
        );

    assert(timelock.is_operation(id), 'should exist');
    assert(timelock.is_operation_pending(id), 'should be pending');
    assert(!timelock.is_operation_ready(id), 'should not be ready');

    start_cheat_block_timestamp_global(mock_ready_time);

    assert(timelock.is_operation_ready(id), 'should be ready');

    assert(timelock.get_timestamp(id) == mock_ready_time.into(), 'timestamps match');
}

#[test]
#[feature("safe_dispatcher")]
fn test_schedule_twice() {
    let (min_delay, _, proposer, _, _, _) = deploy_args();
    let (_, timelock, safe_timelock) = setup_timelock();

    start_cheat_caller_address_global(proposer);

    let calls = array![
        Call {
            target: contract_address_const::<100>(),
            selector: selector!("doesnt_exist"),
            data: array![0x123].span()
        }
    ]
        .span();
    let predecessor = 0;
    let salt = 1;

    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    let result = safe_timelock.schedule_batch(calls, predecessor, salt, min_delay);
    match result {
        Result::Ok(_) => panic!("expect 'operation already scheduled'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'operation already scheduled', *panic_data.at(0));
        }
    }
}

#[test]
#[feature("safe_dispatcher")]
fn test_schedule_blocked() {
    let (min_delay, admin, proposer, _, _, _) = deploy_args();
    let (_, timelock, safe_timelock) = setup_timelock();

    let selector = selector!("doesnt_exist");

    // first, block the fx
    start_cheat_caller_address_global(admin);

    timelock.block_function_selector(selector);

    start_cheat_caller_address_global(proposer);

    let calls = array![
        Call {
            target: contract_address_const::<100>(), selector: selector, data: array![0x123].span()
        }
    ]
        .span();
    let predecessor = 0;
    let salt = 1;

    let result = safe_timelock.schedule_batch(calls, predecessor, salt, min_delay);
    match result {
        Result::Ok(_) => panic!("expect 'selector is blocked'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'selector is blocked', *panic_data.at(0));
        }
    }
}

// 7. test can't cancel operation that isn't scheduled [x]
// 8. test can't cancel operation that was already executed
// test cancel is successful
#[test]
#[feature("safe_dispatcher")]
fn test_cancel_id_not_pending() {
    let (_, _, _, _, canceller, _) = deploy_args();
    let (_, _, safe_timelock) = setup_timelock();

    start_cheat_caller_address_global(canceller);

    // unscheduled id
    let result = safe_timelock.cancel(123123123123);

    match result {
        Result::Ok(_) => panic!("expect 'rbact: cant cancel operation'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'rbact: cant cancel operation', *panic_data.at(0));
        }
    }
// executed id

// todo: add case where it was already executed here
}

#[test]
fn test_cancel_success() {
    let (min_delay, _, proposer, _, canceller, _) = deploy_args();
    let (timelock_address, timelock, _) = setup_timelock();

    let mock_time = 3;

    start_cheat_caller_address_global(proposer);
    start_cheat_block_timestamp_global(mock_time);

    let call = Call {
        target: contract_address_const::<100>(),
        selector: selector!("doesnt_exist"),
        data: array![0x123].span()
    };
    let calls = array![call].span();
    let predecessor = 0;
    let salt = 1;

    let id = timelock.hash_operation_batch(calls, predecessor, salt);
    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    let mut spy = spy_events();

    start_cheat_caller_address_global(canceller);

    timelock.cancel(id);

    spy
        .assert_emitted(
            @array![
                (
                    timelock_address,
                    RBACTimelock::Event::Cancelled(RBACTimelock::Cancelled { id: id })
                )
            ]
        );

    assert(!timelock.is_operation(id), 'not operation');
    assert(timelock.get_timestamp(id) == 0, 'matches 0');
}

// 9. test execute call that isn't ready yet [x]
// 10. test execute call which predecessor is not executed [x]
// 11. test execute call successful + assert it's done
// 12. test the invocation fails for some reason [x]

#[test]
#[feature("safe_dispatcher")]
fn test_execute_op_not_ready() {
    let (min_delay, _, proposer, executor, _, _) = deploy_args();
    let (timelock_address, timelock, safe_timelock) = setup_timelock();

    let mock_time = 3;
    start_cheat_block_timestamp_global(mock_time);

    let call = Call {
        target: contract_address_const::<100>(),
        selector: selector!("doesnt_exist"),
        data: array![0x123].span()
    };
    let calls = array![call].span();
    let predecessor = 0;
    let salt = 1;

    start_cheat_caller_address_global(executor);

    // test not scheduled
    let result = safe_timelock.execute_batch(calls, predecessor, salt);
    match result {
        Result::Ok(_) => panic!("expect 'rbact: operation not ready'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'rbact: operation not ready', *panic_data.at(0));
        }
    }

    // test not ready

    // first, schedule it
    start_cheat_caller_address_global(proposer);
    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    // then, execute
    start_cheat_block_timestamp_global(
        mock_time + min_delay.try_into().unwrap() - 1
    ); // still not ready
    start_cheat_caller_address_global(executor);
    let result = safe_timelock.execute_batch(calls, predecessor, salt);
    match result {
        Result::Ok(_) => panic!("expect 'rbact: operation not ready'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'rbact: operation not ready', *panic_data.at(0));
        }
    }
// todo:
// test executed
}

#[test]
#[feature("safe_dispatcher")]
fn test_execute_predecessor_invalid() {
    let (min_delay, _, proposer, executor, _, _) = deploy_args();
    let (timelock_address, timelock, safe_timelock) = setup_timelock();

    let mock_time = 3;
    let mock_ready_time = mock_time + min_delay.try_into().unwrap();
    start_cheat_block_timestamp_global(mock_time);

    let call = Call {
        target: contract_address_const::<100>(),
        selector: selector!("doesnt_exist"),
        data: array![0x123].span()
    };
    let calls = array![call].span();
    let predecessor = 4;
    let salt = 1;

    start_cheat_caller_address_global(proposer);
    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    start_cheat_caller_address_global(executor);
    start_cheat_block_timestamp_global(mock_ready_time);
    let result = safe_timelock.execute_batch(calls, predecessor, salt);
    match result {
        Result::Ok(_) => panic!("expect 'rbact: missing dependency'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'rbact: missing dependency', *panic_data.at(0));
        }
    }
}
// snforge does not treat contract invocation failures as a panic (yet)
// #[test]
// #[feature("safe_dispatcher")]
// fn test_execute_invalid_call() {
//     let (min_delay, _, proposer, executor, _, _) = deploy_args();
//     let (timelock_address, timelock, safe_timelock) = setup_timelock();

//     let mock_time = 3;
//     let mock_ready_time = mock_time + min_delay.try_into().unwrap();
//     start_cheat_block_timestamp_global(mock_time);

//     let call = Call {
//         target: contract_address_const::<100>(),
//         selector: selector!("doesnt_exist"),
//         data: array![0x123].span()
//     };
//     let calls = array![call].span();
//     let predecessor = 0;
//     let salt = 1;

//     start_cheat_caller_address_global(proposer);
//     timelock.schedule_batch(calls, predecessor, salt, min_delay);

//     // will fail because target contract does not exist
//     start_cheat_block_timestamp_global(mock_ready_time);
//     start_cheat_caller_address_global(executor);
//     let result = safe_timelock.execute_batch(calls, predecessor, salt);
// }

#[test]
fn test_execute_successful() {
    let (target_address, target) = setup_mock_target();
    let (min_delay, _, proposer, executor, _, _) = deploy_args();
    let (timelock_address, timelock, _) = setup_timelock();

    let mock_time = 3;
    let mock_ready_time = mock_time + min_delay.try_into().unwrap();

    let call_1 = Call {
        target: target_address, selector: selector!("set_value"), data: array![0x56162].span()
    };

    let call_2 = Call {
        target: target_address, selector: selector!("flip_toggle"), data: array![].span()
    };

    let calls = array![call_1, call_2].span();
    let predecessor = 0;
    let salt = 1;

    start_cheat_caller_address_global(proposer);
    start_cheat_block_timestamp_global(mock_time);

    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    let id = timelock.hash_operation_batch(calls, predecessor, salt);

    start_cheat_caller_address_global(executor);
    start_cheat_block_timestamp_global(mock_ready_time);

    let mut spy = spy_events();

    timelock.execute_batch(calls, predecessor, salt);

    spy
        .assert_emitted(
            @array![
                (
                    timelock_address,
                    RBACTimelock::Event::CallExecuted(
                        RBACTimelock::CallExecuted {
                            id: id,
                            index: 0,
                            target: call_1.target,
                            selector: call_1.selector,
                            data: call_1.data
                        }
                    )
                ),
                (
                    timelock_address,
                    RBACTimelock::Event::CallExecuted(
                        RBACTimelock::CallExecuted {
                            id: id,
                            index: 1,
                            target: call_2.target,
                            selector: call_2.selector,
                            data: call_2.data
                        }
                    )
                )
            ]
        );

    let (actual_value, actual_toggle) = target.read();
    assert(actual_value == 0x56162, 'value equal');
    assert(actual_toggle, 'toggle true');
}


// 12. test update delay is true. 
// 13. test scheduled batch and then updated delay changes the return of is_operation_ready *not

#[test]
fn test_update_delay_success() {
    let (min_delay, admin, _, _, _, _) = deploy_args();
    let (timelock_address, timelock, _) = setup_timelock();

    start_cheat_caller_address_global(admin);

    let mut spy = spy_events();

    timelock.update_delay(0x92289);
    spy
        .assert_emitted(
            @array![
                (
                    timelock_address,
                    RBACTimelock::Event::MinDelayChange(
                        RBACTimelock::MinDelayChange {
                            old_duration: min_delay, new_duration: 0x92289
                        }
                    )
                )
            ]
        );

    assert(timelock.get_min_delay() == 0x92289, 'new min delay');
}

#[test]
fn test_update_delay_no_affect_op_readiness() {
    let (min_delay, admin, proposer, _, _, _) = deploy_args();
    let (_, timelock, _) = setup_timelock();

    let mock_time = 3;
    let mock_ready_time = mock_time + min_delay.try_into().unwrap();
    start_cheat_block_timestamp_global(mock_time);

    let call = Call {
        target: contract_address_const::<100>(),
        selector: selector!("doesnt_exist"),
        data: array![0x123].span()
    };
    let calls = array![call].span();
    let predecessor = 0;
    let salt = 1;

    let id = timelock.hash_operation_batch(calls, predecessor, salt);

    start_cheat_caller_address_global(proposer);

    timelock.schedule_batch(calls, predecessor, salt, min_delay);

    start_cheat_block_timestamp_global(mock_ready_time);

    assert(timelock.is_operation_ready(id), 'confirm op ready');

    start_cheat_caller_address_global(admin);
    timelock.update_delay(0x92289);

    assert(timelock.is_operation_ready(id), 'op still ready');
}

// 14. test bypasser execute batch works 
// 15. test bypasser execute batch fails
// 16. test block function selector once and then twice
// 17. test unblock fx selector unsuccessful and then successful
// 18. test get block fx selector count
// 19. test get blocked fx selector index throughout a bunch of add and removals
// 20. test is_operation, pending, ready, done throughout life cycle of id
// 21. test block and unblock the max felt size

fn test_bypasser_execute_success() {}

