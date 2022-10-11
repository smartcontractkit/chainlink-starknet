# Equivalent to openzeppelin/Ownable except it's a two step process to transfer ownership.
%lang starknet

from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.starknet.common.syscalls import get_caller_address
from starkware.cairo.common.math import assert_not_zero

#
# Events
#

@event
func OwnershipTransferred(previousOwner : felt, newOwner : felt):
end

@event
func OwnershipTransferRequested(from_address : felt, to : felt):
end

#
# Storage
#

@storage_var
func Ownable_owner() -> (owner : felt):
end

@storage_var
func Ownable_proposed_owner() -> (proposed_owner : felt):
end

namespace Ownable:
    #
    # Constructor
    #
    func initializer{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        owner : felt
    ):
        with_attr error_message("Ownable: Cannot transfer to zero address"):
            assert_not_zero(owner)
        end
        _accept_ownership_transfer(owner)
        return ()
    end

    func assert_only_owner{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
        let (owner) = Ownable_owner.read()
        let (caller) = get_caller_address()
        # caller is the zero address should not be possible anymore with introduction of fees
        with_attr error_message("Ownable: Caller is the zero address"):
            assert_not_zero(caller)
        end
        with_attr error_message("Ownable: caller is not the owner"):
            assert owner = caller
        end
        return ()
    end

    func get_owner{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}() -> (
        owner : felt
    ):
        let (owner) = Ownable_owner.read()
        return (owner=owner)
    end

    func get_proposed_owner{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        ) -> (proposed_owner : felt):
        let (proposed_owner) = Ownable_proposed_owner.read()
        return (proposed_owner=proposed_owner)
    end

    func transfer_ownership{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        new_owner : felt
    ) -> ():
        with_attr error_message("Ownable: Cannot transfer to zero address"):
            assert_not_zero(new_owner)
        end
        assert_only_owner()
        Ownable_proposed_owner.write(new_owner)
        let (previous_owner : felt) = Ownable_owner.read()
        OwnershipTransferRequested.emit(previous_owner, new_owner)
        return ()
    end

    func accept_ownership{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
        let (proposed_owner) = Ownable_proposed_owner.read()
        let (caller) = get_caller_address()
        # caller cannot be zero address to avoid overwriting owner when proposed_owner is not set
        with_attr error_message("Ownable: caller is the zero address"):
            assert_not_zero(caller)
        end
        with_attr error_message("Ownable: caller is not the proposed owner"):
            assert proposed_owner = caller
        end
        _accept_ownership_transfer(proposed_owner)
        return ()
    end

    func renounce_ownership{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
        assert_only_owner()
        _accept_ownership_transfer(0)
        return ()
    end

    #
    # Internal
    #

    func _accept_ownership_transfer{
        syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr
    }(new_owner : felt):
        let (previous_owner : felt) = Ownable_owner.read()
        Ownable_owner.write(new_owner)
        Ownable_proposed_owner.write(0)
        OwnershipTransferred.emit(previous_owner, new_owner)
        return ()
    end
end
