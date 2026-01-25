#!/usr/bin/env bash

set -euo pipefail

MAIN_PID=$$

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
        READY_FOR_COMPLETION=0
        echo
        echo -e "\e[32mStarting Ralph Loop\e[0m"
        echo
        # gemini --yolo --model gemini-3-pro-preview \
        codex exec  "${codex_flags[@]}" \
          < "${spec_dir}/prompt.md" \
          2>&1 | tee /dev/tty | while IFS= read -r line; do
            # Strip ANSI colors to make matching robust.
            clean_line="$(printf '%s' "$line" | sed -E 's/\x1b\[[0-9;]*m//g')"
            # Avoid matching the prompt echo; wait for runtime output markers.
            if [[ "$READY_FOR_COMPLETION" != "1" ]]; then
                if [[ "$clean_line" == mcp\ startup:* ]] || \
                   [[ "$clean_line" == "assistant" ]] || \
                   [[ "$clean_line" == "codex" ]] || \
                   [[ "$clean_line" == "thinking" ]] || \
                   [[ "$clean_line" == "tokens used" ]]; then
                    READY_FOR_COMPLETION=1
                fi
            fi
            if [[ "$READY_FOR_COMPLETION" == "1" ]] && [[ "$clean_line" == *"ALL TASKS COMPLETE"* ]]; then
                echo
                echo -e "\e[32m✓ $clean_line\e[0m"
                echo
                # Kill the main process to trigger clean exit
                kill -TERM "$MAIN_PID"
                exit 0
            fi
        done
    done
}

main "$@"
