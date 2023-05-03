#!/bin/bash

# Function to convert filenames to capitalized camel case
to_camel_case() {
    echo "$1" | sed -r 's/_([a-z])/ \U\1/g' | sed -r 's/-([a-z])/ \U\1/g' | sed -r 's/([a-z0-9])([A-Z])/\1\2/g' | sed -r 's/(^| )([a-z])/\U\2/g' | tr -d ' '
}

src="src"
target="target/release"

find "$src" -type f -iname "*.cairo" | while read -r cairo_file; do
    # Get the relative path and filename
    #relative_path="${cairo_file#$src/}"
    relative_path="${cairo_file}"
    subdir=$(dirname "$relative_path")
    filename=$(basename "$relative_path" .cairo)

    # Convert filename to camel case
    camel_case_filename=$(to_camel_case "$filename")

    # Create compiled sierra and casm filenames
    sierra_file="$target/chainlink_$camel_case_filename.sierra.json"
    casm_file="$target/chainlink_$camel_case_filename.casm.json"

    # Check if sierra_file and casm_file exist
    if [ -e "$sierra_file" ] && [ -e "$casm_file" ]; then
        # Create the corresponding directory
        mkdir -vp "$target/$subdir/$filename.cairo"

        # Copy the sierra and casm files
        cp -vf "$sierra_file" "$target/$subdir/$filename.cairo/$filename.json"
        cp -vf "$casm_file" "$target/$subdir/$filename.cairo/$filename.casm"
        jq '.abi' "${sierra_file}" > "$target/$subdir/$filename.cairo/${filename}_abi.json"
    fi
done

