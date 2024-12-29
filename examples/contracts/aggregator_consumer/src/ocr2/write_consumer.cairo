use starknet::ContractAddress;


#[starknet::interface]
pub trait IWriteConsumer<TContractState> {
    fn write(ref self: TContractState, proxy: ContractAddress);
}

#[starknet::contract]
mod WriteConsumer {
    use starknet::ContractAddress;
    use chainlink::ocr2::aggregator_proxy::{
        IAggregatorProxy, IAggregatorProxyDispatcher, IAggregatorProxyDispatcherTrait
    };

    #[storage]
    struct Storage {
        answer: u128,
        proxy: ContractAddress
    }

    #[constructor]
    fn constructor(ref self: ContractState) {}

    #[derive(Drop, starknet::Event)]
    struct AnswerWrite {
        answer: u128,
        proxy: ContractAddress
    }

    #[event]
    #[derive(Drop, starknet::Event)]
    enum Event {
        AnswerWrite: AnswerWrite
    }

    #[abi(embed_v0)]
    impl WriteConsumerImpl of super::IWriteConsumer<ContractState> {
        fn write(ref self: ContractState, proxy: ContractAddress) {
            let answer = IAggregatorProxyDispatcher { contract_address: proxy }.latest_answer();
            self.answer.write(answer);
            self.proxy.write(proxy);

            self.emit(Event::AnswerWrite(AnswerWrite { answer: answer, proxy: proxy }))
        }
    }
}
