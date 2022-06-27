%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.cairo.common.bool import TRUE, FALSE

@storage_var
func access_list_(address: felt) -> (bool: felt):
end

from contracts.ownable import (
    Ownable_initializer,
    Ownable_only_owner,
    Ownable_get_owner,
    Ownable_transfer_ownership
)

@constructor
func constructor{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    owner: felt,
):
    Ownable_initializer(owner)
    return ()
end

@view
func has_access{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
) -> (bool: felt):
    let (bool) = access_list_.read(address)
    return (bool)
end


@view
func check_access{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
):
    let (bool) = access_list_.read(address)
    with_attr error_message("AccessController: address does not have access"):
        assert bool = TRUE
    end
    return ()
end

@external
func add_access{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
):
    Ownable_only_owner()
    access_list_.write(address, TRUE)
    return ()
end

@external
func remove_access{
    syscall_ptr : felt*,
    pedersen_ptr : HashBuiltin*,
    range_check_ptr,
}(
    address: felt
):
    Ownable_only_owner()
    access_list_.write(address, FALSE)
    return ()
end

@view
func type_and_version() -> (meta: felt):
    return ('access_controller.cairo 1.0.0')
end
