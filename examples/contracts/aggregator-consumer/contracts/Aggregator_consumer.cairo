%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from chainlink.cairo.ocr2.IAggregator import IAggregator, Round

@storage_var
func AggregatorConsumer_ocr_address() -> (address : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(address : felt):
    AggregatorConsumer_ocr_address.write(address)
    return ()
end

@view
func readLatestRound{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    let (address) = AggregatorConsumer_ocr_address.read()
    let (round : Round) = IAggregator.latest_round_data(contract_address=address)
    return (round)
end

@view
func readDecimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    let (address) = AggregatorConsumer_ocr_address.read()
    let (decimals) = IAggregator.decimals(contract_address=address)
    return (decimals)
end
