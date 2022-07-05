#!/bin/sh

if ! command -v godepgraph &> /dev/null
then
    echo 'godepgraph not found! \nCheck https://github.com/kisielk/godepgraph for instructions!'
    exit
fi

if ! command -v dot &> /dev/null
then
    echo 'dot not found! \nCheck https://graphviz.org/ for instructions!'
    exit
fi

usage="usage: $(basename "$0") [-h] [-a] [-p GO_PACKAGE] [-o OUTPUT_PNG_PATH]
Create a dependency diagram for the provided go module:
    -h show this tip
    -a include all dependencies (otherwise include only github.com/smartcontractkit dependencies)
    -p package name (i.e. github.com/smartcontractkit/chainlink-starknet/pkg/chainlink)
    -o output PNG path"

options=':hap:o:'
while getopts $options option; do
  case "$option" in
    h) echo "$usage"; exit;;
    a) ALL_DEPENDENCIES=1;;
    p) GO_MODULE_PATH=$OPTARG;;
    o) OUTPUT_PNG_PATH=$OPTARG;;
  esac
done

if [ ! "$GO_MODULE_PATH" ]; 
then
  echo "-p argument must be provided"
  echo "$usage" >&2; exit 1
fi

if [ ! "$OUTPUT_PNG_PATH" ]; 
then
  OUTPUT_PNG_PATH="./godepgraph.png"
fi

if [ ! "$ALL_DEPENDENCIES" ]; 
then
  godepgraph -s -o .,github.com/smartcontractkit $GO_MODULE_PATH | dot -Tpng -o $OUTPUT_PNG_PATH
else 
  godepgraph -s $GO_MODULE_PATH | dot -Tpng -o $OUTPUT_PNG_PATH
fi

echo "Finished!"
  







