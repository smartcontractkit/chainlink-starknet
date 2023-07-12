use starknet::ContractAddress;

#[starknet::interface]
trait IMintableToken<TContractState> {
    #[external(v0)]
    fn permissionedMint(ref self: TContractState, account: ContractAddress, amount: u256);
    #[external(v0)]
    fn permissionedBurn(ref self: TContractState, account: ContractAddress, amount: u256);
}

#[starknet::contract]
mod LinkToken {
    use super::IMintableToken;

    use zeroable::Zeroable;

    use starknet::ContractAddress;
    use starknet::class_hash::ClassHash;

    use chainlink::libraries::token::erc20::ERC20;
    use chainlink::libraries::token::erc677::ERC677;
    use chainlink::libraries::ownable::Ownable;
    use chainlink::libraries::upgradeable::Upgradeable;

    const NAME: felt252 = 'ChainLink Token';
    const SYMBOL: felt252 = 'LINK';

    #[storage]
    struct Storage {
        _minter: ContractAddress, 
    }

    //
    // IMintableToken (StarkGate)
    //
    #[external(v0)]
    impl MintableToken of IMintableToken<ContractState> {
        fn permissionedMint(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            ERC20::_mint(account, amount);
        }

        fn permissionedBurn(ref self: ContractState, account: ContractAddress, amount: u256) {
            only_minter(@self);
            ERC20::_burn(account, amount);
        }
    }


    #[constructor]
    fn constructor(ref self: ContractState, minter: ContractAddress, owner: ContractAddress) {
        ERC20::initializer(NAME, SYMBOL);
        assert(!minter.is_zero(), 'minter is 0');
        self._minter.write(minter);
        Ownable::initializer(owner);
    }

    #[view]
    fn minter() -> ContractAddress {
        self._minter.read()
    }

    #[view]
    fn type_and_version() -> felt252 {
        'LinkToken 1.0.0'
    }

    // 
    // ERC677
    //

    #[external(v0)]
    fn transfer_and_call(to: ContractAddress, value: u256, data: Array<felt252>) -> bool {
        ERC677::transfer_and_call(to, value, data)
    }

    //
    //  Upgradeable
    //
    #[external(v0)]
    fn upgrade(ref self: ContractState, new_impl: ClassHash) {
        let ownable = Ownable::unsafe_new_contract_state();
        Ownable::assert_only_owner(@ownable);
        Upgradeable::upgrade(new_impl)
    }

    //
    // Ownership
    //

    #[external(v0)]
    impl OwnableImpl of IOwnable<ContractState> {
        fn owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::owner(@state)
        }

        fn proposed_owner(self: @ContractState) -> ContractAddress {
            let state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::proposed_owner(@state)
        }

        fn transfer_ownership(ref self: ContractState, new_owner: ContractAddress) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::transfer_ownership(ref state, new_owner)
        }

        fn accept_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::accept_ownership(ref state)
        }

        fn renounce_ownership(ref self: ContractState) {
            let mut state = Ownable::unsafe_new_contract_state();
            Ownable::OwnableImpl::renounce_ownership(ref state)
        }
    }

    //
    // ERC20
    //

    #[view]
    fn name() -> felt252 {
        ERC20::name()
    }

    #[view]
    fn symbol() -> felt252 {
        ERC20::symbol()
    }

    #[view]
    fn decimals() -> u8 {
        ERC20::decimals()
    }

    #[view]
    fn total_supply() -> u256 {
        ERC20::total_supply()
    }

    #[view]
    fn balance_of(account: ContractAddress) -> u256 {
        ERC20::balance_of(account)
    }

    #[view]
    fn allowance(owner: ContractAddress, spender: ContractAddress) -> u256 {
        ERC20::allowance(owner, spender)
    }

    #[external(v0)]
    fn transfer(recipient: ContractAddress, amount: u256) -> bool {
        ERC20::transfer(recipient, amount)
    }

    #[external(v0)]
    fn transfer_from(sender: ContractAddress, recipient: ContractAddress, amount: u256) -> bool {
        ERC20::transfer_from(sender, recipient, amount)
    }

    #[external(v0)]
    fn approve(spender: ContractAddress, amount: u256) -> bool {
        ERC20::approve(spender, amount)
    }

    #[external(v0)]
    fn increase_allowance(spender: ContractAddress, added_value: u256) -> bool {
        ERC20::increase_allowance(spender, added_value)
    }

    #[external(v0)]
    fn decrease_allowance(spender: ContractAddress, subtracted_value: u256) -> bool {
        ERC20::decrease_allowance(spender, subtracted_value)
    }


    //
    // Internal
    //

    fn only_minter(self: @ContractState) {
        let caller = starknet::get_caller_address();
        let minter = self._minter.read();
        assert(caller == minter, 'only minter');
    }
}
