#!/bin/sh
cpu_struct=`arch`;
echo $cpu_struct;
#rm -rf starknet-hardhat-example
#git clone git@github.com:Shard-Labs/starknet-hardhat-example.git;
#echo "nodejs 16.13.2" > starknet-hardhat-example/.tool-versions;
#cd starknet-hardhat-example;
cd contracts;
node --version;
dpid=`docker ps | grep devnet | awk '{print $1}'`;
echo "Checking for existing docker containers for devnet..."
if [ -z "$dpid" ]
then
    echo "No docker devnet container running...";
else
    docker kill $dpid;
fi

devnet_image=`docker ps -a | grep devnet_local | awk '{print $1}'`
if [ -z "$devnet_image" ]
then
    echo "No docker devnet imagse found...";
else
    docker rm $devnet_image;
fi

echo "Checking CPU structure..."
if [[ $cpu_struct == *"arm"* ]]
then
    echo "Starting arm devnet container..."
    docker run -p 5050:5050 -p 8545:8545 -d --name devnet_local shardlabs/starknet-devnet:0.2.5-arm;
else
    echo "Starting i386 devnet container..."
    docker run -p 5050:5050 -p 8545:8545 -d --name devnet_local shardlabs/starknet-devnet:0.2.5;
fi

echo "Installing dependencies..."
# npm ci
yarn install
# npx hardhat starknet-compile contracts/l1l2.cairo
yarn compile:l1
yarn compile
echo "Checking for running hardhat process..."

hardhat_image=`docker image ls | grep hardhat | awk '{print $3}'`
if [ -z "$hardhat_image" ]
then
    echo "No docker hardhat image found...";
else
    docker rm $hardhat_image;
fi

dpid=`docker ps | grep hardhat | awk '{print $1}'`;
echo "Checking for existing docker containers for hardhat..."
if [ -z "$dpid" ]
then
    echo "No docker hardhat container running...";
else
    docker kill $dpid;
fi
echo "Starting hardhat..."
docker run --net container:devnet_local -d ethereumoptimism/hardhat
echo "Starting L1<>L2 tests"
# npx hardhat test test/postman.test.ts --starknet-network devnet --network localhost
yarn test

dpid=`docker ps | grep devnet | awk '{print $1}'`;
echo "Checking for existing docker containers for devnet..."
if [ -z "$dpid" ]
then
    echo "No docker devnet container running...";
else
    docker kill $dpid;
fi

devnet_image=`docker ps -a | grep devnet_local | awk '{print $1}'`
if [ -z "$devnet_image" ]
then
    echo "No docker devnet imagse found...";
else
    docker rm $devnet_image;
fi

