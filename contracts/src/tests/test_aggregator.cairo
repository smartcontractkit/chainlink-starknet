use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;
use starknet::class_hash::Felt252TryIntoClassHash;
use starknet::syscalls::deploy_syscall;

use array::ArrayTrait;
use traits::Into;
use traits::TryInto;
use option::OptionTrait;
use core::result::ResultTrait;

use chainlink::ocr2::aggregator::pow;
use chainlink::ocr2::aggregator::Aggregator;
use chainlink::ocr2::aggregator::Aggregator::Billing;
use chainlink::access_control::access_controller::AccessController;
use chainlink::tests::test_ownable::should_implement_ownable;
use chainlink::tests::test_access_controller::should_implement_access_control;

// TODO: aggregator tests

#[test]
#[available_gas(10000000)]
fn test_pow_2_0() {
    assert(pow(2, 0) == 0x1, 'expected 0x1');
    assert(pow(2, 1) == 0x2, 'expected 0x2');
    assert(pow(2, 2) == 0x4, 'expected 0x4');
    assert(pow(2, 3) == 0x8, 'expected 0x8');
    assert(pow(2, 4) == 0x10, 'expected 0x10');
    assert(pow(2, 5) == 0x20, 'expected 0x20');
    assert(pow(2, 6) == 0x40, 'expected 0x40');
    assert(pow(2, 7) == 0x80, 'expected 0x80');
    assert(pow(2, 8) == 0x100, 'expected 0x100');
    assert(pow(2, 9) == 0x200, 'expected 0x200');
    assert(pow(2, 10) == 0x400, 'expected 0x400');
    assert(pow(2, 11) == 0x800, 'expected 0x800');
    assert(pow(2, 12) == 0x1000, 'expected 0x1000');
    assert(pow(2, 13) == 0x2000, 'expected 0x2000');
    assert(pow(2, 14) == 0x4000, 'expected 0x4000');
    assert(pow(2, 15) == 0x8000, 'expected 0x8000');
    assert(pow(2, 16) == 0x10000, 'expected 0x10000');
    assert(pow(2, 17) == 0x20000, 'expected 0x20000');
    assert(pow(2, 18) == 0x40000, 'expected 0x40000');
    assert(pow(2, 19) == 0x80000, 'expected 0x80000');
    assert(pow(2, 20) == 0x100000, 'expected 0x100000');
    assert(pow(2, 21) == 0x200000, 'expected 0x200000');
    assert(pow(2, 22) == 0x400000, 'expected 0x400000');
    assert(pow(2, 23) == 0x800000, 'expected 0x800000');
    assert(pow(2, 24) == 0x1000000, 'expected 0x1000000');
    assert(pow(2, 25) == 0x2000000, 'expected 0x2000000');
    assert(pow(2, 26) == 0x4000000, 'expected 0x4000000');
    assert(pow(2, 27) == 0x8000000, 'expected 0x8000000');
    assert(pow(2, 28) == 0x10000000, 'expected 0x10000000');
    assert(pow(2, 29) == 0x20000000, 'expected 0x20000000');
    assert(pow(2, 30) == 0x40000000, 'expected 0x40000000');
    assert(pow(2, 31) == 0x80000000, 'expected 0x80000000');
}

#[abi]
trait IAccessController { // importing from access_controller.cairo doesnt work
    fn has_access(user: ContractAddress, data: Array<felt252>) -> bool;
    fn add_access(user: ContractAddress);
    fn remove_access(user: ContractAddress);
    fn enable_access_check();
    fn disable_access_check();
}

fn setup() -> (ContractAddress, ContractAddress, ContractAddress, IAccessControllerDispatcher) {
    let acc1: ContractAddress = contract_address_const::<777>();
    let acc2: ContractAddress = contract_address_const::<888>();
    // set acc1 as default caller
    set_caller_address(acc1);

    // deploy billing access controller
    let mut calldata = ArrayTrait::new();
    calldata.append(acc1.into()); // owner = acc1;
    let (billingAccessControllerAddr, _) = deploy_syscall(
        AccessController::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();
    let billingAccessController = IAccessControllerDispatcher {
        contract_address: billingAccessControllerAddr
    };

    // return accounts and billing access controller
    (acc1, acc2, billingAccessControllerAddr, billingAccessController)
}

#[test]
#[available_gas(2000000)]
fn test_ownable() {
    let (account, _, _, _) = setup();
    // Deploy aggregator
    let mut calldata = ArrayTrait::new();
    calldata.append(account.into()); // owner
    calldata.append(contract_address_const::<777>().into()); // link token
    calldata.append(0); // min_answer
    calldata.append(100); // max_answer
    calldata.append(contract_address_const::<999>().into()); // billing access controller
    calldata.append(8); // decimals
    calldata.append(123); // description
    let (aggregatorAddr, _) = deploy_syscall(
        Aggregator::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();

    should_implement_ownable(aggregatorAddr, account);
}

#[test]
#[available_gas(2000000)]
fn test_access_control() {
    let (account, _, _, _) = setup();
    // Deploy aggregator
    let mut calldata = ArrayTrait::new();
    calldata.append(account.into()); // owner
    calldata.append(contract_address_const::<777>().into()); // link token
    calldata.append(0); // min_answer
    calldata.append(100); // max_answer
    calldata.append(contract_address_const::<999>().into()); // billing access controller
    calldata.append(8); // decimals
    calldata.append(123); // description
    let (aggregatorAddr, _) = deploy_syscall(
        Aggregator::TEST_CLASS_HASH.try_into().unwrap(), 0, calldata.span(), false
    )
        .unwrap();

    should_implement_access_control(aggregatorAddr, account);
}


#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_non_owner() {
    let sender = setup();
    Aggregator::upgrade(class_hash_const::<123>());
}

// --- Billing tests ---

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_set_billing_access_controller_not_owner() {
    let (owner, acc2, billingAccessControllerAddr, _) = setup();
    // Aggregator initialization
    Aggregator::constructor(owner, contract_address_const::<777>(), 0, 100, acc2, 8, 123);

    // set billing access controller should revert if caller is not owner
    set_caller_address(acc2);
    Aggregator::set_billing_access_controller(billingAccessControllerAddr);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('caller does not have access', ))]
fn test_set_billing_config_no_access() {
    let (owner, acc2, billingAccessControllerAddr, _) = setup();
    // Aggregator initialization
    Aggregator::constructor(
        owner, contract_address_const::<777>(), 0, 100, billingAccessControllerAddr, 8, 123
    );

    // set billing config as acc2 with no access
    let config: Billing = Billing {
        observation_payment_gjuels: 1,
        transmission_payment_gjuels: 5,
        gas_base: 1,
        gas_per_signature: 1,
    };
    set_caller_address(acc2);
    Aggregator::set_billing(config);
}

#[test]
#[available_gas(2000000)]
fn test_set_billing_config_as_owner() {
    let (owner, _, billingAccessControllerAddr, _) = setup();
    // Aggregator initialization
    Aggregator::constructor(
        owner, contract_address_const::<777>(), 0, 100, billingAccessControllerAddr, 8, 123
    );

    // set billing config as owner
    let config: Billing = Billing {
        observation_payment_gjuels: 1,
        transmission_payment_gjuels: 5,
        gas_base: 1,
        gas_per_signature: 1,
    };
    Aggregator::set_billing(config);

    // check billing config
    let billing = Aggregator::billing();
    assert(billing.observation_payment_gjuels == 1, 'should be 1');
    assert(billing.transmission_payment_gjuels == 5, 'should be 5');
    assert(billing.gas_base == 1, 'should be 1');
    assert(billing.gas_per_signature == 1, 'should be 1');
}

#[test]
#[available_gas(2000000)]
fn test_set_billing_config_as_acc_with_access() {
    let (owner, acc2, billingAccessControllerAddr, billingAccessController) = setup();
    // grant acc2 access on access controller
    set_contract_address(owner);
    billingAccessController.add_access(acc2);

    // Aggregator initialization
    Aggregator::constructor(
        owner, contract_address_const::<777>(), 0, 100, billingAccessControllerAddr, 8, 123
    );

    // set billing config as acc2 with access
    let config: Billing = Billing {
        observation_payment_gjuels: 1,
        transmission_payment_gjuels: 5,
        gas_base: 1,
        gas_per_signature: 1,
    };
    set_caller_address(acc2);
    Aggregator::set_billing(config);

    // check billing config
    let billing = Aggregator::billing();
    assert(billing.observation_payment_gjuels == 1, 'should be 1');
    assert(billing.transmission_payment_gjuels == 5, 'should be 5');
    assert(billing.gas_base == 1, 'should be 1');
    assert(billing.gas_per_signature == 1, 'should be 1');
}
