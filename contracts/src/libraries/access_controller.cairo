use starknet::ContractAddress;

trait AccessController {
  fn has_access(user: ContractAddress, data: Array<felt252>) -> bool;
  fn check_access(user: ContractAddress);
}
