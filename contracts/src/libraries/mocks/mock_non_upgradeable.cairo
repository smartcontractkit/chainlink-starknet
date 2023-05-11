#[contract]
mod MockNonUpgradeable {

    #[constructor]
    fn constructor() {}

    #[view]
    fn bar() -> bool {
        true
    }
}
