use starknet::{ContractAddress, contract_address_const};
use chainlink::access_control::rbac_timelock::{
    RBACTimelock, IRBACTimelock, IRBACTimelockDispatcher, IRBACTimelockDispatcherTrait,
    IRBACTimelockSafeDispatcher, IRBACTimelockSafeDispatcherTrait,
    RBACTimelock::{ADMIN_ROLE, PROPOSER_ROLE, EXECUTOR_ROLE, CANCELLER_ROLE, BYPASSER_ROLE}
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
    start_cheat_caller_address_global
};
// 1. test supports access controller, erc1155 receiver, and erc721 receiver
// 2. test has_roles after constructor is called + min delay + event emitted
// 3. test schedule_batch, cancel, execute_batch, update_delay, bypasser_execute_batch , block, unblock with wrong roles
// 4. test schedule with too small of a delay
// 5. test schedule with blocked selector
// 6. test schedule_batch + get operation is true
// 7. test can't cancel operation that isn't scheduled 
// 8. test can't cancel operation that was already executed
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
    let (min_delay, admin, proposer, executor, canceller, bypasser) = deploy_args();
    let (timelock_address, timelock, safe_timelock) = setup_timelock();

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

