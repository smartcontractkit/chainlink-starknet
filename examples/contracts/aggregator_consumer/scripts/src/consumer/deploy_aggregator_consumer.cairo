use sncast_std::{
    declare, deploy, DeclareResult, DeployResult, get_nonce, DisplayContractAddress,
    DisplayClassHash
};

use starknet::ContractAddress;

fn declare_and_deploy(
    contract_name: ByteArray, constructor_calldata: Array<felt252>
) -> DeployResult {
    println!("Declaring contract...");
    let declare_result = declare(contract_name, Option::None, Option::None);
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
        Option::None,
        Option::Some(nonce)
    );
    if deploy_result.is_err() {
        println!("{:?}", deploy_result.unwrap_err());
        panic_with_felt252('deploy failed');
    }

    return deploy_result.unwrap();
}

fn main() {
    // Point this to the address of the aggregator contract you'd like to use
    let aggregator_address: ContractAddress =
        0x376b1abf788737bded2011a0f76ce61cabdeaec22e97b8a4e231b149dd49fc0
        .try_into()
        .unwrap();

    println!("\nDeclaring and deploying AggregatorConsumer");
    let mut calldata = ArrayTrait::new();
    calldata.append(aggregator_address.into());
    let consumer = declare_and_deploy("AggregatorConsumer", calldata);
    println!("AggregatorConsumer deployed at address: {}\n", consumer.contract_address);
}

