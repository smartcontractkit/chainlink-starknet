use sncast_std::{
    declare, deploy, DeclareResult, DeployResult, get_nonce, DisplayContractAddress,
    DisplayClassHash
};

use starknet::{ContractAddress, ClassHash};

fn declare_and_deploy(
    contract_name: ByteArray, constructor_calldata: Array<felt252>
) -> DeployResult {
    let mut class_hash: ClassHash =
        0x6d1dd0e5fa4e0284dcf341997f1d781bc2fb7d76ada684da7a2a33c38031df5
        .try_into()
        .unwrap();

    println!("Declaring contract...");
    let declare_result = declare(contract_name, Option::None, Option::None);
    if declare_result.is_err() {
        println!("{:?}", declare_result.unwrap_err());
    } else {
        class_hash = declare_result.unwrap().class_hash;
    }
    println!("Class hash = {:?}", class_hash);

    println!("Deploying contract...");
    let nonce = get_nonce('latest');
    let salt = get_nonce('pending');
    let deploy_result = deploy(
        class_hash,
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
        0x3c6f82da5dbfa89ec9dbe414f33d23d1720d15568e4a880afcc9b0c3d98d127
        .try_into()
        .unwrap();

    println!("\nDeclaring and deploying AggregatorConsumer");
    let mut calldata = ArrayTrait::new();
    calldata.append(aggregator_address.into());
    let consumer = declare_and_deploy("AggregatorConsumer", calldata);
    println!("AggregatorConsumer deployed at address: {}\n", consumer.contract_address);
}

