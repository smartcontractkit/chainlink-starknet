%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin, SignatureBuiltin
from starkware.starknet.common.syscalls import (
    library_call_l1_handler,
    get_caller_address,
    deploy,
    call_contract,
)
from starkware.cairo.common.bool import TRUE, FALSE

from ocr2.interfaces.IAggregator import IAggregator, Round
from ocr2.interfaces.IAccessController import IAccessController

@storage_var
func s_uptime_feed_address() -> (address : felt):
end

@constructor
func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    uptime_feed_address : felt
):
    s_uptime_feed_address.write(uptime_feed_address)
    return ()
end

@view
func latest_round_data{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    round : Round
):
    let (address) = s_uptime_feed_address.read()
    let (latest_round) = IAggregator.latest_round_data(contract_address=address)
    return (latest_round)
end

@view
func description{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    description : felt
):
    let (address) = s_uptime_feed_address.read()
    let (description) = IAggregator.description(contract_address=address)
    return (description)
end

@view
func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    user : felt, data_len : felt, data : felt*
) -> (bool : felt):
    alloc_locals
    let (local address) = s_uptime_feed_address.read()
    let (has_access : felt) = IAccessController.has_access(
        contract_address=address, address=user, data_len=data_len, data=data
    )
    return (has_access)
end

# @external
# func l1_handle_test{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
#     class_hash : felt
# ):
#     let (caller) = get_caller_address()
#     let (contract_address) = deploy(
#         class_hash=class_hash,
#         contract_address_salt=1,
#         constructor_calldata_size=2,
#         constructor_calldata=cast(new (1, caller), felt*),
#     )
#     # const function_selector = 0x38142ae23cdc38788f60223e0d361c4a67d9a841d171c3eac2e8181fbc757d3
#     # library_call_l1_handler(
#     #     class_hash=class_hash,
#     #     function_selector=function_selector,
#     #     calldata_size=3,
#     #     calldata=cast(new (1, 1, 1), felt*),
#     # )

# # call_contract(
#     #     contract_address=contract_address,
#     #     function_selector=function_selector,
#     #     calldata_size=3,
#     #     calldata=cast(new (1, 1, 1), felt*),
#     # )

# const function_selector = 0x28ffe4ff0f226a9107253e17a904099aa4f63a02a5621de0576e5aa71bc5194
#     call_contract(
#         contract_address=contract_address,
#         function_selector=function_selector,
#         calldata_size=2,
#         calldata=cast(new (1, 1, 1), felt*),
#     )

# return ()
# end
