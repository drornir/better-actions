#!/usr/bin/env bash

function main() {
    local spec_dir=$1
    if [[ -z "$spec_dir" ]]; then
        echo "Usage: $0 <spec_dir>"
        exit 1
    fi

    if [[ ! -d "$spec_dir" ]]; then
        echo "Error: $spec_dir is not a directory"
        exit 1
    fi

    while true; do
        echo -e "\e[32mStarting Ralph Loop\e[0m"
        echo
        echo
        gemini --yolo --output-format text < "${spec_dir}/prompt.md"
    done
}

main "$@"
