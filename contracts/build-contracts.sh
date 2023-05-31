#!/bin/sh

set -eu

DIR=$(dirname $(readlink -f $0))

SRC_PATH="${DIR}/src"
OUT_PATH="${DIR}/target/release"

STARKNET_COMPILE="starknet-compile"
STARKNET_SIERRA_COMPILE="starknet-sierra-compile"

if [ "$#" -eq 0 ]; then
  LIBFUNC_VALUE="experimental_v0.1.0"
else
  LIBFUNC_VALUE="$1"
fi

case $LIBFUNC_VALUE in
  *.json)
    LIBFUNC_ARG="--allowed-libfuncs-list-file"
    echo "Using libfunc list file: ${LIBFUNC_VALUE}"
    ;;
  *)
    LIBFUNC_ARG="--allowed-libfuncs-list-name"
    echo "Using libfunc list name: ${LIBFUNC_VALUE}"
    ;;
esac

# Run the command, capture the output
OUTPUT=$($STARKNET_COMPILE $SRC_PATH 2>&1 || true)

# Extract all contract paths from the output
CONTRACT_PATHS=$(echo "$OUTPUT" | grep -oP '^\s+\S+::\S+\s*$' || true)

if [ -z "${CONTRACT_PATHS}" ]; then
  echo "Failed to find contract paths, make sure starknet-compile is working:"
  echo ""
  echo "\t${STARKNET_COMPILE} ${SRC_PATH}"
  echo ""
  exit 1
fi

mkdir -p "${OUT_PATH}"

# For each contract path
echo "${CONTRACT_PATHS}" | while read -r CONTRACT_PATH; do
  echo ""
  echo "Compiling: ${CONTRACT_PATH}"

  CLASS_NAME=$(echo $CONTRACT_PATH | awk -F '::' '{print $NF}')

  SIERRA_FILENAME="chainlink_${CLASS_NAME}.sierra.json"
  SIERRA_PATH="${OUT_PATH}/${SIERRA_FILENAME}"
  CASM_FILENAME="chainlink_${CLASS_NAME}.casm.json"
  CASM_PATH="${OUT_PATH}/${CASM_FILENAME}"

  echo "* ${SIERRA_FILENAME}" && \
  "${STARKNET_COMPILE}" "${SRC_PATH}" --contract-path $CONTRACT_PATH "${LIBFUNC_ARG}" "${LIBFUNC_VALUE}" "${SIERRA_PATH}" && \
  echo "* ${CASM_FILENAME}" && \
  "${STARKNET_SIERRA_COMPILE}" --add-pythonic-hints "${LIBFUNC_ARG}" "${LIBFUNC_VALUE}" "${SIERRA_PATH}" "${CASM_PATH}"
done

echo ""
echo "Done."
