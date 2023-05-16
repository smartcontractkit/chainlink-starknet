use starknet::ContractAddress;

#[abi]
trait IMintableToken {
    #[external]
    fn permissionedMint(account: ContractAddress, amount: u256);
    #[external]
    fn permissionedBurn(account: ContractAddress, amount: u256);
}

#[contract]
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

    struct Storage {
        _minter: ContractAddress, 
    }

    impl MintableToken of IMintableToken {
        fn permissionedMint(account: ContractAddress, amount: u256) {
            only_minter();
            ERC20::_mint(account, amount);
        }

        fn permissionedBurn(account: ContractAddress, amount: u256) {
            only_minter();
            ERC20::_burn(account, amount);
        }
    }


    #[constructor]
    fn constructor(minter: ContractAddress, owner: ContractAddress) {
        ERC20::initializer(NAME, SYMBOL);
        assert(!minter.is_zero(), 'minter is 0');
        _minter::write(minter);
        Ownable::initializer(owner);
    }

    #[view]
    fn minter() -> ContractAddress {
        _minter::read()
    }

    #[view]
    fn type_and_version() -> felt252 {
        'LinkToken 1.0.0'
    }

    // 
    // ERC677
    //

    #[external]
    fn transfer_and_call(to: ContractAddress, value: u256, data: Array<felt252>) -> bool {
        ERC677::transfer_and_call(to, value, data)
    }

    //
    // IMintableToken (StarkGate)
    //

    #[external]
    fn permissionedMint(account: ContractAddress, amount: u256) {
        MintableToken::permissionedMint(account, amount)
    }

    #[external]
    fn permissionedBurn(account: ContractAddress, amount: u256) {
        MintableToken::permissionedBurn(account, amount)
    }

    //
    //  Upgradeable
    //
    #[external]
    fn upgrade(new_impl: ClassHash) {
        Ownable::assert_only_owner();
        Upgradeable::upgrade(new_impl)
    }

    //
    // Ownership
    //

    #[view]
    fn owner() -> ContractAddress {
        Ownable::owner()
    }

    #[view]
    fn proposed_owner() -> ContractAddress {
        Ownable::proposed_owner()
    }

    #[external]
    fn transfer_ownership(new_owner: ContractAddress) {
        Ownable::transfer_ownership(new_owner)
    }

    #[external]
    fn accept_ownership() {
        Ownable::accept_ownership()
    }

    #[external]
    fn renounce_ownership() {
        Ownable::renounce_ownership()
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

    #[external]
    fn transfer(recipient: ContractAddress, amount: u256) -> bool {
        ERC20::transfer(recipient, amount)
    }

    #[external]
    fn transfer_from(sender: ContractAddress, recipient: ContractAddress, amount: u256) -> bool {
        ERC20::transfer_from(sender, recipient, amount)
    }

    #[external]
    fn approve(spender: ContractAddress, amount: u256) -> bool {
        ERC20::approve(spender, amount)
    }

    #[external]
    fn increase_allowance(spender: ContractAddress, added_value: u256) -> bool {
        ERC20::increase_allowance(spender, added_value)
    }

    #[external]
    fn decrease_allowance(spender: ContractAddress, subtracted_value: u256) -> bool {
        ERC20::decrease_allowance(spender, subtracted_value)
    }


    //
    // Internal
    //

    fn only_minter() {
        let caller = starknet::get_caller_address();
        let minter = minter();
        assert(caller == minter, 'only minter');
    }
}
