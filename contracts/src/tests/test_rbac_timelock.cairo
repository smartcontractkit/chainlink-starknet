use starknet::ContractAddress;
use chainlink::access_control::rbac_timelock::{
    RBACTimelock, IRBACTimelock, IRBACTimelockDispatcher, IRBACTimelockDispatcherTrait,
    IRBACTimelockSafeDispatcher, IRBACTimelockSafeDispatcherTrait
};
// 1. test supports access controller, erc1155 receiver, and erc721 receiver
// 2. test has_roles after constructor is called + min delay + event emitted
// 3. test schedule_batch, cancel, execute_batch, update_delay, bypasser_execute_batch , block, unblock
// 4. test schedule with too small of a delay
// 5. test schedule with blocked selector
// 6. test schedule_batch + get operation is true
// 7. test can't cancel operation that isn't scheduled 
// 8. test can't cancel operation that was already executed
// 9. test execute call that isn't ready yet
// 10. test execute call which predecessor is not executed
// 11. test execute call successful + assert it's done
// 12. test update delay is true. 
// 13. test scheduled batch and then updated delay changes the return of is_operation_ready
// 14. test bypasser execute batch works 
// 15. test bypasser execute batch fails
// 16. test block function selector once and then twice
// 17. test unblock fx selector unsuccessful and then successful
// 18. test get block fx selector count
// 19. test get blocked fx selector index throughout a bunch of add and removals
// 20. test is_operation, pending, ready, done throughout life cycle of id
// 21. test block and unblock the max felt size

// fn setup_timelock() -> (ContractAddress, IRBACTimelockDispatcher, IRBACTimelockSafeDispatcher) {

// }


