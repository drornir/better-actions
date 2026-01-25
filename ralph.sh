#!/usr/bin/env bash

set -euo pipefail

OUTFILE="$(mktemp)"
MAIN_PID=$$

function clean_up() {
    # Kill the tail process if it's still running
    if [[ -n "${TAIL_PID:-}" ]] && kill -0 "$TAIL_PID" 2>/dev/null; then
        kill "$TAIL_PID" 2>/dev/null || true
    fi
    rm -f "$OUTFILE"
}

trap clean_up EXIT

function tail_for_completion() {
    tail -f "$OUTFILE" 2>/dev/null | while IFS= read -r line; do
        if [[ "$line" == *"ALL TASKS COMPLETE!"* ]]; then
            echo
            echo -e "\e[32m✓ $line\e[0m"
            echo
            # Kill the main process to trigger clean exit
            kill -TERM "$MAIN_PID"
            exit 0
        fi
    done
}

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

    # Create the output file before tailing
    touch "$OUTFILE"

    # Start monitoring the output file in the background
    tail_for_completion &
    TAIL_PID=$!

    local script_dir
    script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local codex_flags=(
        --color always
        -c model="gpt-5.2-codex"
        -c model_reasoning_effort="medium"
        -c features.web_search_request=true
        -c sandbox_workspace_write.network_access=true
        --dangerously-bypass-approvals-and-sandbox
    )

    while true; do
        echo
        echo -e "\e[32mStarting Ralph Loop\e[0m"
        echo
        # gemini --yolo --model gemini-3-pro-preview \
        codex exec  "${codex_flags[@]}" \
          < "${spec_dir}/prompt.md" \
          | tee "$OUTFILE"
    done
}

main "$@"
