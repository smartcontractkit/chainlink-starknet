use starknet::ContractAddress;
use chainlink::libraries::mocks::mock_enumerable_set::{
    MockEnumerableSet, IMockEnumerableSet, IMockEnumerableSetDispatcher,
    IMockEnumerableSetDispatcherTrait, IMockEnumerableSetSafeDispatcher,
    IMockEnumerableSetSafeDispatcherTrait
};
use snforge_std::{declare, ContractClassTrait, DeclareResultTrait};

const MOCK_SET_ID: felt252 = 'adfasdf';
const OTHER_SET_ID: felt252 = 'fakeasdf';

fn expect_out_of_bounds<T, impl TDrop: Drop<T>>(result: Result<T, Array<felt252>>) {
    match result {
        Result::Ok(_) => panic!("expect 'index out of bounds'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'index out of bounds', *panic_data.at(0));
        }
    }
}

fn expect_set_is_1_indexed<T, impl TDrop: Drop<T>>(result: Result<T, Array<felt252>>) {
    match result {
        Result::Ok(_) => panic!("expect 'set is 1-indexed'"),
        Result::Err(panic_data) => {
            assert(*panic_data.at(0) == 'set is 1-indexed', *panic_data.at(0));
        }
    }
}

fn setup_mock() -> (
    ContractAddress, IMockEnumerableSetDispatcher, IMockEnumerableSetSafeDispatcher
) {
    let calldata = array![];
    let (mock_address, _) = declare("MockEnumerableSet")
        .unwrap()
        .contract_class()
        .deploy(@calldata)
        .unwrap();

    (
        mock_address,
        IMockEnumerableSetDispatcher { contract_address: mock_address },
        IMockEnumerableSetSafeDispatcher { contract_address: mock_address }
    )
}

#[test]
fn test_add() {
    let (_, mock, _) = setup_mock();

    // ensure that adding to other sets do not interfere with current set
    mock.add(OTHER_SET_ID, 6);

    let first_value = 12;

    assert(mock.add(MOCK_SET_ID, first_value), 'should add');

    assert(mock.contains(MOCK_SET_ID, first_value), 'should contain');
    assert(mock.length(MOCK_SET_ID) == 1, 'should equal 1');
    assert(mock.at(MOCK_SET_ID, 1) == first_value, 'should return val');
    assert(mock.values(MOCK_SET_ID) == array![first_value], 'arrays should equal');

    assert(!mock.add(MOCK_SET_ID, first_value), 'should not add');

    let second_value = 100;

    assert(mock.add(MOCK_SET_ID, second_value), 'should add');
    assert(
        mock.contains(MOCK_SET_ID, first_value) && mock.contains(MOCK_SET_ID, second_value),
        'should contain'
    );
    assert(mock.length(MOCK_SET_ID) == 2, 'should equal 2');
    assert(
        mock.at(MOCK_SET_ID, 1) == first_value && mock.at(MOCK_SET_ID, 2) == second_value,
        'should return val'
    );
    assert(mock.values(MOCK_SET_ID) == array![first_value, second_value], 'arrays should equal');
}

#[test]
#[feature("safe_dispatcher")]
fn test_remove() {
    let (_, mock, safe_mock) = setup_mock();
    let first_value = 12;

    // ensure that removing other sets do not interfere with current set
    mock.add(OTHER_SET_ID, 6);
    mock.add(OTHER_SET_ID, 7);
    mock.remove(OTHER_SET_ID, 7);

    assert(!mock.remove(MOCK_SET_ID, first_value), 'should not remove');

    // [12]
    mock.add(MOCK_SET_ID, first_value);

    // []
    assert(mock.remove(MOCK_SET_ID, first_value), 'should remove');

    assert(!mock.contains(MOCK_SET_ID, first_value), 'should not contain');
    assert(mock.length(MOCK_SET_ID) == 0, 'len should == 0');
    assert(mock.values(MOCK_SET_ID) == array![], 'should be empty array');

    // [100, 200, 300]
    mock.add(MOCK_SET_ID, 100);
    mock.add(MOCK_SET_ID, 200);
    mock.add(MOCK_SET_ID, 300);

    // [100, 200]
    assert(mock.remove(MOCK_SET_ID, 300), 'remove 300 from end');
    assert(mock.length(MOCK_SET_ID) == 2, 'length should equal 2');
    assert(!mock.contains(MOCK_SET_ID, 300), 'does not contain 300');
    assert(
        mock.contains(MOCK_SET_ID, 100) && mock.contains(MOCK_SET_ID, 200), 'contains 100 & 200'
    );
    assert(mock.at(MOCK_SET_ID, 1) == 100 && mock.at(MOCK_SET_ID, 2) == 200, 'indexes match');
    expect_out_of_bounds(safe_mock.at(MOCK_SET_ID, 3));
    assert(mock.values(MOCK_SET_ID) == array![100, 200], 'values should match');

    // [100, 200, 300]
    mock.add(MOCK_SET_ID, 300);

    // [300, 200]
    assert(mock.remove(MOCK_SET_ID, 100), 'remove 100');
    assert(mock.length(MOCK_SET_ID) == 2, 'length should equal 2');
    assert(!mock.contains(MOCK_SET_ID, 100), 'does not contain 100');
    assert(
        mock.contains(MOCK_SET_ID, 300) && mock.contains(MOCK_SET_ID, 200), 'contains 300 & 200'
    );
    assert(mock.at(MOCK_SET_ID, 1) == 300 && mock.at(MOCK_SET_ID, 2) == 200, 'indexes match');
    expect_out_of_bounds(safe_mock.at(MOCK_SET_ID, 3));
    assert(mock.values(MOCK_SET_ID) == array![300, 200], 'values should match');

    // [200]
    assert(mock.remove(MOCK_SET_ID, 300), 'remove 300');
    assert(mock.length(MOCK_SET_ID) == 1, 'length should equal 1');
    assert(!mock.contains(MOCK_SET_ID, 300), 'does not contain 300');
    assert(mock.contains(MOCK_SET_ID, 200), 'contains 200');
    assert(mock.at(MOCK_SET_ID, 1) == 200, 'indexes match');
    expect_out_of_bounds(safe_mock.at(MOCK_SET_ID, 2));
    assert(mock.values(MOCK_SET_ID) == array![200], 'values should match');

    // []
    assert(mock.remove(MOCK_SET_ID, 200), 'remove 200');

    assert(mock.length(MOCK_SET_ID) == 0, 'empty list');
    assert(mock.values(MOCK_SET_ID) == array![], 'empty list');
}

#[test]
fn test_contains() {
    let (_, mock, _) = setup_mock();

    assert(!mock.contains(MOCK_SET_ID, 6), 'should not contain');

    mock.add(MOCK_SET_ID, 7);

    assert(mock.contains(MOCK_SET_ID, 7), 'should contain');

    assert(!mock.contains(OTHER_SET_ID, 7), 'should not contain');
}

#[test]
fn test_length() {
    let (_, mock, _) = setup_mock();

    assert(mock.length(MOCK_SET_ID) == 0, 'should be 0');

    mock.add(MOCK_SET_ID, 7);

    assert(mock.length(MOCK_SET_ID) == 1, 'should be 1');

    assert(mock.length(OTHER_SET_ID) == 0, 'should be 0');
}

#[test]
#[feature("safe_dispatcher")]
fn test_zero() {
    let (_, mock, safe_mock) = setup_mock();

    expect_set_is_1_indexed(safe_mock.at(MOCK_SET_ID, 0));

    // [0]
    assert(mock.add(MOCK_SET_ID, 0), 'should add 0');
    assert(mock.contains(MOCK_SET_ID, 0), 'contains 0');

    assert(mock.length(MOCK_SET_ID) == 1, 'should be 1');

    // [0, 1]
    assert(mock.add(MOCK_SET_ID, 1), 'should add 1');
    assert(!mock.add(MOCK_SET_ID, 1), 'shouldnt add 1');

    assert(mock.length(MOCK_SET_ID) == 2, 'should be 2');

    assert(mock.at(MOCK_SET_ID, 1) == 0, 'set[1] = 0');
    assert(mock.at(MOCK_SET_ID, 2) == 1, 'set[2] = 0');

    // [1]
    assert(mock.remove(MOCK_SET_ID, 0), 'should remove 0');
    assert(!mock.remove(MOCK_SET_ID, 0), 'shouldnt remove 0');

    assert(mock.at(MOCK_SET_ID, 1) == 1, 'set[1] = 1');
    assert(!mock.contains(MOCK_SET_ID, 0), '0 is gone');
    assert(mock.length(MOCK_SET_ID) == 1, 'length 1');

    // []
    assert(mock.remove(MOCK_SET_ID, 1), '1 removed');

    // [0]
    mock.add(MOCK_SET_ID, 0);

    assert(mock.at(MOCK_SET_ID, 1) == 0, 'set[1] = 0');

    // []
    mock.remove(MOCK_SET_ID, 0);

    expect_out_of_bounds(safe_mock.at(MOCK_SET_ID, 1));
}

