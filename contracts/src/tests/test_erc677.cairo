use chainlink::libraries::token::erc677::ERC677;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::testing::set_caller_address;
use array::ArrayTrait;
use traits::Into;
use zeroable::Zeroable;
use chainlink::token::mock::valid_erc667_receiver::ValidReceiver;
use chainlink::token::mock::invalid_erc667_receiver::InvalidReceiver;


// Ignored tests are dependent on upgrading our version of cairo to include this PR https://github.com/starkware-libs/cairo/pull/2912/files

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<1>();
    // Set account as default caller
    set_caller_address(account);
    account
}

// todo
fn setup_valid_receiver() -> ContractAddress {
    contract_address_const::<234>()
}

// todo
fn setup_invalid_receiver() -> ContractAddress {
    contract_address_const::<123>()
}

fn transfer_and_call(receiver: ContractAddress) {
    let data = ArrayTrait::<felt252>::new();
    // have to send 0 because ERC20 is not initialized with starting supply when using this library by itself
    ERC677::transfer_and_call(receiver, u256 { high: 0, low: 0 }, data);
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('ERC20: transfer to 0', ))]
fn test_to_zero_address() {
    setup();
    transfer_and_call(Zeroable::zero());
}

#[test]
#[available_gas(2000000)]
#[ignore]
fn test_valid_transfer_and_call() {
    setup();
    let receiver = setup_valid_receiver();

    transfer_and_call(receiver);
// assert Transfer event emited (not supported yet)

// assert that receiver's onTokenTransfer called was called
}

#[test]
#[available_gas(2000000)]
#[ignore]
fn test_invalid_receiver_supports_interface_true() {
    setup();
    let receiver = setup_invalid_receiver();

    transfer_and_call(receiver);
// should panic
}

#[test]
#[available_gas(2000000)]
#[ignore]
fn test_invalid_receiver_supports_interface_false() {
    setup();
    let receiver = setup_invalid_receiver();

    transfer_and_call(receiver);
// should panic
}


#[test]
#[available_gas(2000000)]
#[ignore]
fn test_nonexistent_receiver() {
    setup();

    transfer_and_call(contract_address_const::<777>());
// should panic
}

