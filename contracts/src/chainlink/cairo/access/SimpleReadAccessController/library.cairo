%lang starknet

from starkware.cairo.common.alloc import alloc
from starkware.cairo.common.cairo_builtins import HashBuiltin
from starkware.starknet.common.syscalls import get_tx_info
from starkware.cairo.common.bool import TRUE, FALSE

from chainlink.cairo.access.SimpleWriteAccessController.library import (
    SimpleWriteAccessController,
    owner,
    proposed_owner,
    transfer_ownership,
    accept_ownership,
    add_access,
    remove_access,
    enable_access_check,
    disable_access_check,
)

namespace SimpleReadAccessController {
    func initialize{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
        owner_address: felt
    ) {
        SimpleWriteAccessController.initialize(owner_address);
        return ();
    }

    // Gives access to:
    // - any externally owned account (note that offchain actors can always read
    // any contract storage regardless of onchain access control measures, so this
    // does not weaken the access control while improving usability)
    // - accounts explicitly added to an access list
    func has_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(
        user: felt, data_len: felt, data: felt*
    ) -> (bool: felt) {
        let (has_access) = SimpleWriteAccessController.has_access(user, data_len, data);
        if (has_access == TRUE) {
            return (TRUE,);
        }

        // NOTICE: access is granted to direct calls, to enable off-chain reads.
        if (user == 0) {
            return (TRUE,);
        }

        return (FALSE,);
    }

    // TODO: remove when starkware adds get_class_hash_at
    func check_access{syscall_ptr: felt*, pedersen_ptr: HashBuiltin*, range_check_ptr}(user: felt) {
        alloc_locals;

        let empty_data_len = 0;
        let (empty_data) = alloc();

        let (bool) = SimpleReadAccessController.has_access(user, empty_data_len, empty_data);
        with_attr error_message("SimpleReadAccessController: address does not have access") {
            assert bool = TRUE;
        }

        return ();
    }
}
