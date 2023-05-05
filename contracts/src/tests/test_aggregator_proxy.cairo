use starknet::testing::set_caller_address;
use starknet::ContractAddress;
use starknet::contract_address_const;
use starknet::class_hash::class_hash_const;

use chainlink::ocr2::aggregator_proxy::AggregatorProxy;

fn setup() -> ContractAddress {
    let account: ContractAddress = contract_address_const::<777>();
    set_caller_address(account);
    account
}

#[test]
#[available_gas(2000000)]
#[should_panic(expected: ('Ownable: caller is not owner', ))]
fn test_upgrade_non_owner() {
    let sender = setup();
    AggregatorProxy::upgrade(class_hash_const::<123>());

}
