use sncast_std::{
    declare, deploy, DeclareResult, DeployResult, get_nonce, DisplayContractAddress,
    DisplayClassHash
};

use starknet::ContractAddress;

fn declare_and_deploy(
    contract_name: ByteArray, constructor_calldata: Array<felt252>
) -> DeployResult {
    println!("Declaring contract...");
    let max_fee = 99999999999999999;
    let declare_result = declare(contract_name, Option::Some(max_fee), Option::None);
    if declare_result.is_err() {
        println!("{:?}", declare_result.unwrap_err());
        panic_with_felt252('declare failed');
    }

    println!("Deploying contract...");
    let nonce = get_nonce('latest');
    let salt = get_nonce('pending');
    let deploy_result = deploy(
        declare_result.unwrap().class_hash,
        constructor_calldata,
        Option::Some(salt),
        true,
        Option::Some(max_fee),
        Option::Some(nonce)
    );
    if deploy_result.is_err() {
        println!("{:?}", deploy_result.unwrap_err());
        panic_with_felt252('deploy failed');
    }

    return deploy_result.unwrap();
}

fn deploy_mock_aggregator(decimals: u8) -> DeployResult {
    let mut calldata = ArrayTrait::new();
    calldata.append(decimals.into());
    return declare_and_deploy("MockAggregator", calldata);
}

fn deploy_consumer(aggregator_address: ContractAddress) -> DeployResult {
    let mut calldata = ArrayTrait::new();
    calldata.append(aggregator_address.into());
    return declare_and_deploy("AggregatorConsumer", calldata);
}

fn main() {
    println!("");

    println!("Declaring and deploying MockAggregator");
    let mock_aggregator = deploy_mock_aggregator(16);
    println!("MockAggregator deployed at address: {}", mock_aggregator.contract_address);
    println!("");

    println!("Declaring and deploying AggregatorConsumer");
    let consumer = deploy_consumer(mock_aggregator.contract_address);
    println!("AggregatorConsumer deployed at address: {}", consumer.contract_address);
    println!("");
}

