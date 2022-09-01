# Equivalent to openzeppelin/Ownable except it's a two step process to transfer ownership.
%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.math import assert_not_zero

@storage_var
func Ownable_owner() -> (owner_address : felt):
end

@storage_var
func Ownable_proposed_owner() -> (proposed_owner_address : felt):
end

func Ownable_initializer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    owner : felt
):
    Ownable_owner.write(owner)
    return ()
end

func Ownable_only_owner{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    let (owner) = Ownable_owner.read()
    let (caller) = get_caller_address()
    with_attr error_message("Ownable: Caller is the zero address"):
        assert_not_zero(caller)
    end
    with_attr error_message("Ownable: caller is not the owner"):
        assert owner = caller
    end
    return ()
end

func Ownable_get_owner{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
    owner : felt
):
    let (owner) = Ownable_owner.read()
    return (owner=owner)
end

func Ownable_transfer_ownership{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    new_owner : felt
) -> ():
    with_attr error_message("Ownable: Cannot transfer to zero address"):
        assert_not_zero(new_owner)
    end
    Ownable_only_owner()
    Ownable_proposed_owner.write(new_owner)
    return ()
end

func Ownable_accept_ownership{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
    ) -> (new_owner : felt):
    let (proposed_owner) = Ownable_proposed_owner.read()
    let (caller) = get_caller_address()
    # caller cannot be zero address to avoid overwriting owner when proposed_owner is not set
    with_attr error_message("Ownable: caller is the zero address"):
        assert_not_zero(caller)
    end
    with_attr error_message("Ownable: caller is not the proposed owner"):
        assert proposed_owner = caller
    end
    Ownable_owner.write(proposed_owner)
    return (new_owner=proposed_owner)
end
