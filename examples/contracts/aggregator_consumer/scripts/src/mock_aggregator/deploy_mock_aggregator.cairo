use sncast_std::{
    declare, deploy, DeployResult, get_nonce, DisplayContractAddress, DisplayClassHash
};

use starknet::{ContractAddress, ClassHash};

fn declare_and_deploy(
    contract_name: ByteArray, constructor_calldata: Array<felt252>
) -> DeployResult {
    let mut class_hash: ClassHash =
        0x728d8a221e2204c88df0642b7c6dcee60f7c3d3b3d5c190cac1ceba5baf15e8
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
    let decimals = 16;
    println!("\nDeclaring and deploying MockAggregator");
    let mut calldata = ArrayTrait::new();
    calldata.append(decimals.into());
    let mock_aggregator = declare_and_deploy("MockAggregator", calldata);
    println!("MockAggregator deployed at address: {}\n", mock_aggregator.contract_address);
}

