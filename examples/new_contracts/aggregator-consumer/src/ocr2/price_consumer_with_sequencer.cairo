
#[contract]
mod AggregatorPriceConsumerWithSequencer {
    use box::BoxTrait;
    use starknet::ContractAddress;
    use zeroable::Zeroable;
    use traits::Into;

    use chainlink::ocr2::aggregator::Round;
    use chainlink::ocr2::aggregator_proxy::IAggregator;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcher;
    use chainlink::ocr2::aggregator_proxy::IAggregatorDispatcherTrait;

    struct Storage {
        _uptime_feed_address: ContractAddress,
        _aggregator_address: ContractAddress,
    }

    // Sequencer-aware aggregator consumer
    // retrieves the latest price from the data feed only if
    // the uptime feed is not stale (stale = older than 60 seconds)
    #[constructor]
    fn constructor(uptime_feed_address: ContractAddress, aggregator_address: ContractAddress) {
        assert(!uptime_feed_address.is_zero(), 'uptime feed is 0');
        assert(!aggregator_address.is_zero(), 'aggregator is 0');
        _uptime_feed_address::write(uptime_feed_address);
        _aggregator_address::write(aggregator_address);
    }


    #[view]
    fn get_latest_price() -> u128 {
        assert_sequencer_healthy();
        let round = IAggregatorDispatcher{
            contract_address: _aggregator_address::read()
        }.latest_round_data();
        round.answer
    }

    #[external]
    fn assert_sequencer_healthy() {
        let round = IAggregatorDispatcher{
            contract_address: _uptime_feed_address::read()
        }.latest_round_data();
        let timestamp = starknet::info::get_block_info().unbox().block_timestamp;

        // After 60 sec the report is considered stale
        let report_stale = timestamp - round.updated_at > 60_u64;

        // 0 if the sequencer is up and 1 if it is down. No other options besides 1 and 0
        match round.answer.into() {
            0 => {
                assert(!report_stale, 'L2 seq up & report stale');
            },
            _ => {
                assert(!report_stale, 'L2 seq down & report stale');
                assert(false, 'L2 seq down & report ok');
            }
        }
    }



}

