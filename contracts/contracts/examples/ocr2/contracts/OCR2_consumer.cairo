%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from interfaces.IAggregator import IAggregator, Round

@storage_var
func store_latest_round_() -> (round : Round):
end

@storage_var
func decimals_() -> (decimals : felt):
end

@storage_var
func ocr_address_() -> (address : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        address : felt):
    ocr_address_.write(address)
    return ()
end

@view
func readStoredRound{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (round : Round):
    let (round : Round) = store_latest_round_.read()
    return (round)
end

@external
func storeLatestRound{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (address) = ocr_address_.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=address)
    store_latest_round_.write(round)
    return ()
end

@view
func readDecimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (decimals : felt):
    let (decimals) = decimals_.read()
    return (decimals)
end

@external
func storeDecimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (address) = ocr_address_.read()
    let (decimals) = IAggregator.decimals(contract_address=address)
    decimals_.write(decimals)
    return ()
end