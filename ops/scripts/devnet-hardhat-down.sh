#!/usr/bin/env bash
# TODO: this script needs to be replaced with a predefined K8s enviroment

echo "Cleaning up Hardhat container..."

dpid=`docker ps | grep chainlink-starknet.hardhat | awk '{print $1}'`;
echo "Checking for existing 'chainlink-starknet.hardhat' docker container..."
if [ -z "$dpid" ]
then
    echo "No docker Hardhat container running.";
else
    docker kill $dpid;
fi
docker rm "chainlink-starknet.hardhat";

echo "Cleaning up Starknet Devnet container..."

dpid=`docker ps | grep chainlink-starknet.starknet-devnet | awk '{print $1}'`;
echo "Checking for existing 'chainlink-starknet.starknet-devnet' docker container..."
if [ -z "$dpid" ]
then
    echo "No docker Starknet Devnet container running.";
else
    docker kill $dpid;
fi
docker rm "chainlink-starknet.starknet-devnet";

echo "Cleanup finished."
