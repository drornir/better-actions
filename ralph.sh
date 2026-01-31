#!/usr/bin/env bash

set -euo pipefail

# trigger a commit to get 1password to prompt
git commit -m 'fake commit' --allow-empty
git reset HEAD~1

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
        -c web_search_request=true
        -c sandbox_workspace_write.network_access=true
        --dangerously-bypass-approvals-and-sandbox
    )

    while true; do
        local completed=0
        local saw_complete_phrase_from_prompt=0
        echo
        echo -e "\e[32mStarting Ralph Loop\e[0m"
        echo

        while IFS= read -r line; do
            if [[ "$line" == *"ALL TASKS COMPLETE"* ]]; then
                if [[ "$saw_complete_phrase_from_prompt" != "1" ]]; then
                    saw_complete_phrase_from_prompt=1
                else
                    completed=1
                fi
            fi
        done < <(
            codex exec "${codex_flags[@]}" \
                < "${spec_dir}/prompt.md" \
                2>&1 | tee /dev/tty
        )

        if [[ "$completed" == "1" ]]; then
            echo
            echo -e "\e[32mGOT ALL TASKS COMPLETE from LLM\e[0m"
            exit 0
        fi
    done
}

main "$@"
