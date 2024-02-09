#!/usr/bin/env bash

echo "Cleaning up core containers.."

echo "Checking for existing 'chainlink.core' docker containers..."

for i in {1..4}
do
	echo " Checking for chainlink.core.$i"
	dpid=$(docker ps -a | grep chainlink.core.$i | awk '{print $1}')
	if [ -z "$dpid" ]; then
		echo "No docker core container running."
	else
		docker kill $dpid
		docker rm $dpid
	fi
done

echo "Cleanup finished."
