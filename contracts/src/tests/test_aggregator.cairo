use chainlink::ocr2::aggregator::pow;

// TODO: aggregator tests

#[test]
#[available_gas(10000000)]
fn test_pow_2_0() {
    assert(pow(2, 0) == 0x1, 'expected 0x1');
    assert(pow(2, 1) == 0x2, 'expected 0x2');
    assert(pow(2, 2) == 0x4, 'expected 0x4');
    assert(pow(2, 3) == 0x8, 'expected 0x8');
    assert(pow(2, 4) == 0x10, 'expected 0x10');
    assert(pow(2, 5) == 0x20, 'expected 0x20');
    assert(pow(2, 6) == 0x40, 'expected 0x40');
    assert(pow(2, 7) == 0x80, 'expected 0x80');
    assert(pow(2, 8) == 0x100, 'expected 0x100');
    assert(pow(2, 9) == 0x200, 'expected 0x200');
    assert(pow(2, 10) == 0x400, 'expected 0x400');
    assert(pow(2, 11) == 0x800, 'expected 0x800');
    assert(pow(2, 12) == 0x1000, 'expected 0x1000');
    assert(pow(2, 13) == 0x2000, 'expected 0x2000');
    assert(pow(2, 14) == 0x4000, 'expected 0x4000');
    assert(pow(2, 15) == 0x8000, 'expected 0x8000');
    assert(pow(2, 16) == 0x10000, 'expected 0x10000');
    assert(pow(2, 17) == 0x20000, 'expected 0x20000');
    assert(pow(2, 18) == 0x40000, 'expected 0x40000');
    assert(pow(2, 19) == 0x80000, 'expected 0x80000');
    assert(pow(2, 20) == 0x100000, 'expected 0x100000');
    assert(pow(2, 21) == 0x200000, 'expected 0x200000');
    assert(pow(2, 22) == 0x400000, 'expected 0x400000');
    assert(pow(2, 23) == 0x800000, 'expected 0x800000');
    assert(pow(2, 24) == 0x1000000, 'expected 0x1000000');
    assert(pow(2, 25) == 0x2000000, 'expected 0x2000000');
    assert(pow(2, 26) == 0x4000000, 'expected 0x4000000');
    assert(pow(2, 27) == 0x8000000, 'expected 0x8000000');
    assert(pow(2, 28) == 0x10000000, 'expected 0x10000000');
    assert(pow(2, 29) == 0x20000000, 'expected 0x20000000');
    assert(pow(2, 30) == 0x40000000, 'expected 0x40000000');
    assert(pow(2, 31) == 0x80000000, 'expected 0x80000000');
}
