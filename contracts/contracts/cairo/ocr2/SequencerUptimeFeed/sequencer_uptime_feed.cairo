%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import get_tx_info, get_block_timestamp

from cairo.ocr2.interfaces.IAggregator import Round
from cairo.ocr2.interfaces.IAccessController import IAccessController
from cairo.ocr2.SequencerUptimeFeed.library import sequencer_uptime_feed
from cairo.SimpleReadAccessController.library import simple_read_access_controller

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    initial_status : felt, owner_address : felt
):
    sequencer_uptime_feed.initialize(initial_status, owner_address)
    return ()
end

# implements IAggregator
@l1_handler
func update_status{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    from_address : felt, status : felt, timestamp : felt
):
    sequencer_uptime_feed.update_status(from_address, status, timestamp)
    return ()
end

# implements IAggregator
@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    let (latest_round) = sequencer_uptime_feed.latest_round_data()
    return (latest_round)
end

# implements IAggregator
@view
func round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    round_id : felt
) -> (res : Round):
    let (round) = sequencer_uptime_feed.round_data(round_id)
    return (round)
end

# implements IAggregator
@view
func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    description : felt
):
    let (description) = sequencer_uptime_feed.description()
    return (description)
end

# implements IAggregator
@view
func decimals{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    decimals : felt
):
    let (decimals) = sequencer_uptime_feed.decimals()
    return (decimals)
end

# implements IAggregator
@view
func type_and_version{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    meta : felt
):
    let (meta) = sequencer_uptime_feed.type_and_version()
    return (meta)
end

# implements IAccessController
@view
func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    user : felt, data_len : felt, data : felt*
) -> (bool : felt):
    let (has_access) = simple_read_access_controller.has_access(user, data_len, data)
    return (has_access)
end

# implements IAccessController
@view
func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    simple_read_access_controller.check_access(user)
    return ()
end
