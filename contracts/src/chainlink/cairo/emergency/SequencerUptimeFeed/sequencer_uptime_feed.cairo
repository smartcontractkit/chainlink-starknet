%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info, get_block_timestamp

from chainlink.cairo.access.IAccessController import IAccessController
from chainlink.cairo.access.SimpleReadAccessController.library import SimpleReadAccessController
from chainlink.cairo.ocr2.IAggregator import Round
from chainlink.cairo.emergency.SequencerUptimeFeed.library import SequencerUptimeFeed

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    initial_status : felt, owner_address : felt
):
    SequencerUptimeFeed.initialize(initial_status, owner_address)
    return ()
end

# implements IAggregator
@l1_handler
func update_status{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    from_address : felt, status : felt, timestamp : felt
):
    SequencerUptimeFeed.update_status(from_address, status, timestamp)
    return ()
end

# implements IAggregator
@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    let (latest_round) = SequencerUptimeFeed.latest_round_data()
    return (latest_round)
end

# implements IAggregator
@view
func round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (res : Round):
    let (round) = SequencerUptimeFeed.round_data(round_id)
    return (round)
end

# implements IAggregator
@view
func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    description : felt
):
    let (description) = SequencerUptimeFeed.description()
    return (description)
end

# implements IAggregator
@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    let (decimals) = SequencerUptimeFeed.decimals()
    return (decimals)
end

# implements IAggregator
@view
func type_and_version{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    meta : felt
):
    let (meta) = SequencerUptimeFeed.type_and_version()
    return (meta)
end

# implements IAccessController
@view
func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    user : felt, data_len : felt, data : felt*
) -> (bool : felt):
    let (has_access) = SimpleReadAccessController.has_access(user, data_len, data)
    return (has_access)
end

# implements IAccessController
@view
func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    SimpleReadAccessController.check_access(user)
    return ()
end
