// set_config tests

// 1. test if lena(signer_address) = 0 => revert 
// 2. test if lena(signer_address) > MAX_NUM_SIGNERS => revert
// 3. test if signer addresses and signer groups not same size
// 4. test if group_quorum and group_parents not same size
// 5. test if one of signer_group #'s is out of bounds NUM_GROUPS
// 6. test if group_parents[i] is greater than or equal to i (when not 0) there is revert
// 7. test if i is 0 and group_parents[i] != 0 and revert
// 8. test if there is a signer in a group where group_quorum[i] == 0 => revert
// 9. test if there are not enough signers to meet a quorum => revert
// 10. test if signer addresses are not in ascending order
// 11. successful => test without clearing root. test the state of storage variables and that event was emitted


