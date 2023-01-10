%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.math import assert_not_zero
from chainlink.cairo.ocr2.IAggregator import IAggregator, Round

@storage_var
func consumer_proxy_address() -> (address: felt) {
}

@storage_var
func feed_data() -> (data: Round) {
}

@constructor
func constructor{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
    proxy_address: felt
) {
    assert_not_zero(proxy_address);
    consumer_proxy_address.write(proxy_address);
    get_latest_round_data();
    return ();
}

// Get the latest data and store it.
@external
func get_latest_round_data{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    round: Round
) {
    let (proxy: felt) = consumer_proxy_address.read();
    let (round: Round) = IAggregator.latest_round_data(contract_address=proxy);
    feed_data.write(round);
    return (round=round);
}

@view
func get_stored_round{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    res: Round
) {
    let (res) = feed_data.read();
    return (res=res);
}

@view
func get_stored_feed_address{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}() -> (
    res: felt
) {
    let (res) = consumer_proxy_address.read();
    return (res=res);
}