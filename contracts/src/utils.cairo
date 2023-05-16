use integer::U128IntoFelt252;
use integer::u128s_from_felt252;
use integer::U128sFromFelt252Result;
fn split_felt(felt: felt252) -> (u128, u128) {
    match u128s_from_felt252(felt) {
        U128sFromFelt252Result::Narrow(low) => (0_u128, low),
        U128sFromFelt252Result::Wide((high, low)) => (high, low),
    }
}

