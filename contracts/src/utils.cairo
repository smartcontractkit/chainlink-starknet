use integer::U128IntoFelt252;
use integer::u128s_from_felt252;
use integer::U128sFromFelt252Result;
use core::integer::u128_byte_reverse;
use core::keccak::compute_keccak_byte_array;
use alexandria_bytes::{Bytes, BytesTrait};

fn split_felt(felt: felt252) -> (u128, u128) {
    match u128s_from_felt252(felt) {
        U128sFromFelt252Result::Narrow(low) => (0_u128, low),
        U128sFromFelt252Result::Wide((high, low)) => (high, low),
    }
}

pub impl ByteArrayUtil of Into<Bytes, ByteArray> {
    fn into(self: Bytes) -> ByteArray {
        let mut res: ByteArray = Default::default();
        let mut offset = 0;
        while offset < self.size() {
            if offset + 16 <= self.size() {
                let (new_offset, value) = self.read_u128(offset);
                res.append_word(value.into(), 16);
                offset = new_offset;
            } else {
                let (new_offset, value) = self.read_u8(offset);
                res.append_byte(value);
                offset = new_offset;
            }
        };
        res
    }
}

pub fn u256_reverse_endian(input: u256) -> u256 {
    let low = u128_byte_reverse(input.high);
    let high = u128_byte_reverse(input.low);
    u256 { low, high }
}

// never use compute_keccak_byte_array directly because it
// returns little-endian while evm implementations use big-endian
pub fn keccak(input: @ByteArray) -> u256 {
    u256_reverse_endian(compute_keccak_byte_array(input))
}
