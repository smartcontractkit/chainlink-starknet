use starknet::class_hash_const;
use starknet::contract_address_const;
use starknet::syscalls::deploy_syscall;
use starknet::testing::set_caller_address;
use starknet::testing::set_contract_address;
use starknet::Felt252TryIntoClassHash;

use array::ArrayTrait;
use option::OptionTrait;
use result::ResultTrait;
use traits::TryInto;

use chainlink::multisig::assert_unique_values;
use chainlink::multisig::Multisig;
use chainlink::multisig::Multisig::MultisigImpl;

#[starknet::contract]
mod MultisigTest {
    use array::ArrayTrait;

    #[storage]
    struct Storage {}

    #[external(v0)]
    fn increment(ref self: ContractState, val1: felt252, val2: felt252) -> Array<felt252> {
        array![val1 + 1, val2 + 1]
    }
}

// TODO: test set_threshold with recursive call
// TODO: test set_signers with recursive call
// TODO: test set_signers_and_thershold with recursive call

fn STATE() -> Multisig::ContractState {
    Multisig::contract_state_for_testing()
}

fn sample_calldata() -> Array::<felt252> {
    array![1, 2, 32]
}

#[test]
#[available_gas(2000000)]
fn test_assert_unique_values_empty() {
    let a = ArrayTrait::<felt252>::new();
    assert_unique_values(@a);
}

#[test]
#[available_gas(2000000)]
fn test_assert_unique_values_no_duplicates() {
    let a = array![1, 2, 3];
    assert_unique_values(@a);
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_assert_unique_values_with_duplicate() {
    let a = array![1, 2, 3, 3];
    assert_unique_values(@a);
}

#[test]
#[available_gas(2000000)]
fn test_is_signer_true() {
    let mut state = STATE();
    let signer = contract_address_const::<1>();
    let mut signers = ArrayTrait::new();
    signers.append(signer);
    Multisig::constructor(ref state, :signers, threshold: 1);
    assert(MultisigImpl::is_signer(@state, signer), 'should be signer');
}

#[test]
#[available_gas(2000000)]
fn test_is_signer_false() {
    let mut state = STATE();
    let not_signer = contract_address_const::<2>();
    let mut signers = ArrayTrait::new();
    signers.append(contract_address_const::<1>());
    Multisig::constructor(ref state, :signers, threshold: 1);
    assert(!MultisigImpl::is_signer(@state, not_signer), 'should be signer');
}

#[test]
#[available_gas(2000000)]
fn test_signer_len() {
    let mut state = STATE();
    let mut signers = ArrayTrait::new();
    signers.append(contract_address_const::<1>());
    signers.append(contract_address_const::<2>());
    Multisig::constructor(ref state, :signers, threshold: 1);
    assert(MultisigImpl::get_signers_len(@state) == 2, 'should equal 2 signers');
}

#[test]
#[available_gas(2000000)]
fn test_get_signers() {
    let mut state = STATE();
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];

    Multisig::constructor(ref state, :signers, threshold: 1);
    let returned_signers = MultisigImpl::get_signers(@state);
    assert(returned_signers.len() == 2, 'should match signers length');
    assert(*returned_signers.at(0) == signer1, 'should match signer 1');
    assert(*returned_signers.at(1) == signer2, 'should match signer 2');
}

#[test]
#[available_gas(2000000)]
fn test_get_threshold() {
    let mut state = STATE();
    let mut signers = ArrayTrait::new();
    signers.append(contract_address_const::<1>());
    signers.append(contract_address_const::<2>());
    Multisig::constructor(ref state, :signers, threshold: 1);
    assert(MultisigImpl::get_threshold(@state) == 1, 'should equal threshold of 1');
}

#[test]
#[available_gas(2000000)]
fn test_submit_transaction() {
    let mut state = STATE();
    let signer = contract_address_const::<1>();
    let signers = array![signer];
    Multisig::constructor(ref state, :signers, threshold: 1);

    set_caller_address(signer);
    let to = contract_address_const::<42>();
    let function_selector = 10;
    MultisigImpl::submit_transaction(ref state, :to, :function_selector, calldata: sample_calldata());

    let (transaction, calldata) = MultisigImpl::get_transaction(@state, 0);
    assert(transaction.to == to, 'should match target address');
    assert(transaction.function_selector == function_selector, 'should match function selector');
    assert(transaction.calldata_len == sample_calldata().len(), 'should match calldata length');
    assert(!transaction.executed, 'should not be executed');
    assert(transaction.confirmations == 0, 'should not have confirmations');
// TODO: compare calldata when loops are supported
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_submit_transaction_not_signer() {
    let mut state = STATE();
    let signer = contract_address_const::<1>();
    let signers = array![signer];
    Multisig::constructor(ref state, :signers, threshold: 1);

    set_caller_address(contract_address_const::<3>());
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
}

#[test]
#[available_gas(2000000)]
fn test_confirm_transaction() {
    let mut state = STATE();
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);

    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);

    assert(MultisigImpl::is_confirmed(@state, nonce: 0, signer: signer1), 'should be confirmed');
    assert(!MultisigImpl::is_confirmed(@state, nonce: 0, signer: signer2), 'should not be confirmed');
    let (transaction, _) = MultisigImpl::get_transaction(@state, 0);
    assert(transaction.confirmations == 1, 'should have confirmation');
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_confirm_transaction_not_signer() {
    let mut state = STATE();
    let signer = contract_address_const::<1>();
    let not_signer = contract_address_const::<2>();
    let signers = array![signer];
    Multisig::constructor(ref state, :signers, threshold: 1);
    set_caller_address(signer);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );

    set_caller_address(not_signer);
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
}

#[test]
#[available_gas(4000000)]
fn test_revoke_confirmation() {
    let mut state = STATE();
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);

    MultisigImpl::revoke_confirmation(ref state, nonce: 0);

    assert(!MultisigImpl::is_confirmed(@state, nonce: 0, signer: signer1), 'should not be confirmed');
    assert(!MultisigImpl::is_confirmed(@state, nonce: 0, signer: signer2), 'should not be confirmed');
    let (transaction, _) = MultisigImpl::get_transaction(@state, 0);
    assert(transaction.confirmations == 0, 'should not have confirmation');
}

#[test]
#[available_gas(4000000)]
#[should_panic]
fn test_revoke_confirmation_not_signer() {
    let mut state = STATE();
    let signer = contract_address_const::<1>();
    let not_signer = contract_address_const::<2>();
    let mut signers = array![signer];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);

    set_caller_address(not_signer);
    MultisigImpl::revoke_confirmation(ref state, nonce: 0);
}

#[test]
#[available_gas(2000000)]
#[should_panic]
fn test_execute_confirmation_below_threshold() {
    let mut state = STATE();
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    MultisigImpl::execute_transaction(ref state, nonce: 0);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('only multisig allowed', ))]
fn test_upgrade_not_multisig() {
    let mut state = STATE();
    let account = contract_address_const::<777>();
    set_caller_address(account);

    MultisigImpl::upgrade(ref state, class_hash_const::<1>())
}

#[test]
#[available_gas(80000000)]
fn test_execute() {
    let mut state = STATE();
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    let (test_address, _) = deploy_syscall(
        MultisigTest::TEST_CLASS_HASH.try_into().unwrap(), 0, ArrayTrait::new().span(), false
    )
        .unwrap();
    set_caller_address(signer1);
    let increment_calldata = array![42, 100];
    MultisigImpl::submit_transaction(
        ref state,
        to: test_address,
        // increment()
        function_selector: 0x7a44dde9fea32737a5cf3f9683b3235138654aa2d189f6fe44af37a61dc60d,
        calldata: increment_calldata,
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(signer2);
    MultisigImpl::confirm_transaction(ref state, nonce: 0);

    let response = MultisigImpl::execute_transaction(ref state, nonce: 0);
    assert(response.len() == 3, 'expected response length 3');
    assert(*response.at(0) == 2, 'expected array length 2');
    assert(*response.at(1) == 43, 'expected array value 43');
    assert(*response.at(2) == 101, 'expected array value 101');
}

#[test]
#[available_gas(8000000)]
#[should_panic(expected: ('invalid signer', ))]
fn test_execute_not_signer() {
    let mut state = STATE();
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(signer2);
    MultisigImpl::confirm_transaction(ref state, nonce: 0);

    set_caller_address(contract_address_const::<3>());
    MultisigImpl::execute_transaction(ref state, nonce: 0);
}

#[test]
#[available_gas(8000000)]
#[should_panic(expected: ('transaction invalid', ))]
fn test_execute_after_set_signers() {
    let mut state = STATE();
    let contract_address = contract_address_const::<100>();
    set_contract_address(contract_address);
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signer3 = contract_address_const::<3>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(signer2);
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(contract_address);
    let new_signers = array![signer2, signer3];
    MultisigImpl::set_signers(ref state, new_signers);

    set_caller_address(signer2);
    MultisigImpl::execute_transaction(ref state, nonce: 0);
}

#[test]
#[available_gas(8000000)]
#[should_panic(expected: ('transaction invalid', ))]
fn test_execute_after_set_signers_and_threshold() {
    let mut state = STATE();
    let contract_address = contract_address_const::<100>();
    set_contract_address(contract_address);
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signer3 = contract_address_const::<3>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(signer2);
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(contract_address);
    let new_signers = array![signer2, signer3];
    MultisigImpl::set_signers_and_threshold(ref state, new_signers, 1);

    set_caller_address(signer2);
    MultisigImpl::execute_transaction(ref state, nonce: 0);
}

#[test]
#[available_gas(8000000)]
#[should_panic(expected: ('transaction invalid', ))]
fn test_execute_after_set_threshold() {
    let mut state = STATE();
    let contract_address = contract_address_const::<100>();
    set_contract_address(contract_address);
    let signer1 = contract_address_const::<1>();
    let signer2 = contract_address_const::<2>();
    let signers = array![signer1, signer2];
    Multisig::constructor(ref state, :signers, threshold: 2);
    set_caller_address(signer1);
    MultisigImpl::submit_transaction(
        ref state, to: contract_address_const::<42>(), function_selector: 10, calldata: sample_calldata(), 
    );
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(signer2);
    MultisigImpl::confirm_transaction(ref state, nonce: 0);
    set_caller_address(contract_address);
    MultisigImpl::set_threshold(ref state, 1);

    set_caller_address(signer1);
    MultisigImpl::execute_transaction(ref state, nonce: 0);
}
