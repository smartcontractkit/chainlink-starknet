#!/usr/bin/env bash
# TODO: this script needs to be replaced with a predefined K8s enviroment

echo "Cleaning up postgres container.."

echo "Checking for existing 'chainlink.postgres' docker container..."
dpid=$(docker ps -a | grep chainlink.postgres | awk '{print $1}')
if [ -z "$dpid" ]; then
	echo "No docker postgres container running."
else
	docker kill $dpid
	docker rm $dpid
fi

echo "Cleanup finished."
