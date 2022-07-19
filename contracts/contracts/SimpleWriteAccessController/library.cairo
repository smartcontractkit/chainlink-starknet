%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.bool import TRUE, FALSE
from starkware.cairo.common.cairo_builtins import HashBuiltin

from ownable import Ownable_only_owner, Ownable_initializer

struct Phase:
    member id : felt
    member aggregator_address : felt
end

@event
func AddedAccess(user : felt):
end

@event
func RemovedAccess(user : felt):
end

@event
func CheckAccessEnabled():
end

@event
func CheckAccessDisabled():
end

@storage_var
func s_check_enabled() -> (checkEnabled : felt):
end

@storage_var
func s_access_list(address : felt) -> (bool : felt):
end

# Adds an address to the access list
@external
func add_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    Ownable_only_owner()
    _add_access(user)
    return ()
end

func _add_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    let (has_access) = s_access_list.read(user)
    if has_access == FALSE:
        s_access_list.write(user, TRUE)
        AddedAccess.emit(user)
        return ()
    end

    return ()
end

# # Removes an address from the access list
@external
func remove_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(user : felt):
    Ownable_only_owner()
    let (has_access) = s_access_list.read(user)
    if has_access == TRUE:
        s_access_list.write(user, FALSE)
        RemovedAccess.emit(user)
        return ()
    end

    return ()
end

# makes the access check enforced
@external
func enable_access_check{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    Ownable_only_owner()
    let (check_enabled) = s_check_enabled.read()
    if check_enabled == FALSE:
        s_check_enabled.write(TRUE)
        CheckAccessEnabled.emit()
        return ()
    end

    return ()
end

# makes the access check unenforced
@external
func disable_access_check{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}():
    Ownable_only_owner()
    let (check_enabled) = s_check_enabled.read()
    if check_enabled == TRUE:
        s_check_enabled.write(FALSE)
        CheckAccessDisabled.emit()
        return ()
    end

    return ()
end

namespace simple_write_access_controller:
    func constructor{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        owner_address : felt
    ):
        Ownable_initializer(owner_address)
        s_check_enabled.write(TRUE)

        return ()
    end

    # Virtual method needs wrapper
    func has_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        user : felt, data_len : felt, data : felt*
    ) -> (bool : felt):
        let (has_access) = s_access_list.read(user)
        let (check_enabled) = s_check_enabled.read()

        if has_access == TRUE:
            return (TRUE)
        end

        if check_enabled == FALSE:
            return (TRUE)
        end

        return (FALSE)
    end

    # Why having address here as an argument
    func check_access{syscall_ptr : felt*, pedersen_ptr : HashBuiltin*, range_check_ptr}(
        address : felt
    ):
        alloc_locals

        let empty_data_len = 0
        let (empty_data) = alloc()

        let (bool) = simple_write_access_controller.has_access(address, empty_data_len, empty_data)
        with_attr error_message("AccessController: address does not have access"):
            assert bool = TRUE
        end

        return ()
    end
end
