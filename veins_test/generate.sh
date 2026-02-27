#!/bin/bash
#
# Usage:
#   ./generate.sh --all           # Generate all, tracking progress
#   ./generate.sh elixir          # Generate test for elixir (no progress tracking)
#   ./generate.sh --reset         # Clear progress and start fresh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VEINS_YML="$REPO_ROOT/pkg/embed/veins.yml"
HELLOS_ROOT="${HELLOS_ROOT:-$HOME/repos/100hellos}"
PROGRESS_FILE="$SCRIPT_DIR/.generate-passed"

export FRAGLET_VEIN_TAG_DISCOVERY_ORDER="local,latest"

if [[ ! -f "$VEINS_YML" ]]; then
    echo "Error: veins.yml not found at $VEINS_YML" >&2
    exit 1
fi

if [[ ! -d "$HELLOS_ROOT" ]]; then
    echo "Error: 100hellos directory not found at $HELLOS_ROOT" >&2
    echo "Set HELLOS_ROOT environment variable to override" >&2
    exit 1
fi

get_hellos_dir() {
    local vein_name="$1"
    awk -v name="$vein_name" '
        /^  - name: / { in_vein = ($3 == name) }
        in_vein && /container: 100hellos\// {
            sub(/.*100hellos\//, "")
            sub(/:.*/, "")
            print
            exit
        }
    ' "$VEINS_YML"
}

get_extension() {
    local vein_name="$1"
    awk -v name="$vein_name" '
        /^  - name: / { in_vein = ($3 == name) }
        in_vein && /extensions:/ {
            s = $0
            gsub(/.*\[/, "", s)
            gsub(/\].*/, "", s)
            split(s, exts, ",")
            gsub(/^[[:space:]]*/, "", exts[1])
            print exts[1]
            exit
        }
    ' "$VEINS_YML"
}

get_test_extension() {
    local vein_name="$1"
    local override
    override=$(awk -v name="$vein_name" '
        /^  - name: / { in_vein = ($3 == name) }
        in_vein && /testExtension:/ {
            gsub(/.*testExtension:[[:space:]]*/, "")
            print
            exit
        }
    ' "$VEINS_YML")
    if [[ -n "$override" ]]; then
        echo "$override"
    else
        get_extension "$vein_name"
    fi
}

extract_fragment() {
    local file="$1"
    local code
    code=$(awk '
        /<<[\047"]?EOF[\047"]?/ {
            while (getline > 0) {
                if ($0 == "EOF") exit
                print
            }
        }
    ' "$file")
    if [[ -n "$code" ]]; then
        echo "$code"
        return
    fi
    awk -F"'" '/printf.*>.*\$tmp/ { for (i=4; i<=NF; i+=2) if (length($i) > 0) print $i }' "$file"
}

extract_fraglet_code() {
    local hello_file="$1"
    local code
    code=$(awk '/BEGIN_FRAGLET/{found=1; next} /END_FRAGLET/{exit} found{print}' "$hello_file")
    if [[ -n "$code" ]]; then
        echo "$code"
        return
    fi
    sed '/^#!/d' "$hello_file"
}

find_source_file() {
    local hellos_dir="$1"
    local ext="$2"
    local result
    result=$(find "$HELLOS_ROOT/$hellos_dir/files" -name "hello-world${ext}" 2>/dev/null | head -1)
    if [[ -n "$result" ]]; then
        echo "$result"
        return
    fi
    result=$(find "$HELLOS_ROOT/$hellos_dir/files" -name "*${ext}" 2>/dev/null | head -1)
    if [[ -n "$result" ]]; then
        echo "$result"
        return
    fi
    find "$HELLOS_ROOT/$hellos_dir/files" -name "hello-world.*" ! -name "hello-world.sh" 2>/dev/null | head -1
}

generate_lang() {
    local lang="$1"
    local lang_dir="$SCRIPT_DIR/$lang"
    local hellos_dir
    hellos_dir=$(get_hellos_dir "$lang")
    [[ -z "$hellos_dir" ]] && hellos_dir="$lang"
    local hellos_lang_dir="$HELLOS_ROOT/$hellos_dir"

    local ext
    ext=$(get_extension "$lang")
    if [[ -z "$ext" ]]; then
        echo "  SKIP $lang: no extension" >&2
        return 1
    fi

    local test_ext
    test_ext=$(get_test_extension "$lang")

    local hello_file
    hello_file=$(find_source_file "$hellos_dir" "$ext")
    if [[ -z "$hello_file" ]]; then
        echo "  SKIP $lang: no source file" >&2
        return 1
    fi

    local code
    code=$(extract_fraglet_code "$hello_file")
    if [[ -z "$code" ]]; then
        echo "  SKIP $lang: empty extraction" >&2
        return 1
    fi

    mkdir -p "$lang_dir"

    local filename="test${test_ext}"
    local script_file="$lang_dir/$filename"
    {
        echo "#!/usr/bin/env -S fragletc --vein=$lang"
        echo "$code"
    } > "$script_file"
    chmod +x "$script_file"

    local act_file="$lang_dir/act.sh"
    local assert_file="$lang_dir/assert.txt"
    { echo "#!/bin/sh"; echo "set -e"; echo "chmod +x ./*${test_ext} 2>/dev/null || true"; echo "./$filename"; } > "$act_file"
    chmod +x "$act_file"

    local stdin_script="$hellos_lang_dir/fraglet/verify_stdin.sh"
    if [[ -f "$stdin_script" ]]; then
        local stdin_code
        stdin_code=$(extract_fragment "$stdin_script") || true
        if [[ -n "$stdin_code" ]]; then
            local stdin_file="$lang_dir/stdin_upper${test_ext}"
            { echo "#!/usr/bin/env -S fragletc --vein=$lang"; echo "$stdin_code"; } > "$stdin_file"
            chmod +x "$stdin_file"
            echo '' >> "$act_file"
            echo 'echo ""' >> "$act_file"
            echo 'echo "=== Test: Stdin ==="' >> "$act_file"
            echo "echo \"hello\" | ./stdin_upper${test_ext}" >> "$act_file"
        fi
    fi

    local args_script="$hellos_lang_dir/fraglet/verify_args.sh"
    if [[ -f "$args_script" ]]; then
        local args_code
        args_code=$(extract_fragment "$args_script") || true
        if [[ -n "$args_code" ]]; then
            local args_file="$lang_dir/echo_args${test_ext}"
            { echo "#!/usr/bin/env -S fragletc --vein=$lang"; echo "$args_code"; } > "$args_file"
            chmod +x "$args_file"
            echo '' >> "$act_file"
            echo 'echo ""' >> "$act_file"
            echo 'echo "=== Test: Argument passing ==="' >> "$act_file"
            echo "./echo_args${test_ext} foo bar baz" >> "$act_file"
        fi
    fi

    if ! command -v fragletc >/dev/null 2>&1; then
        echo "  FAIL $lang: fragletc not found" >&2
        return 1
    fi

    local stdout_tmp stderr_tmp
    stdout_tmp=$(mktemp)
    stderr_tmp=$(mktemp)
    (cd "$lang_dir" && ./act.sh </dev/null) >"$stdout_tmp" 2>"$stderr_tmp" || true
    local output
    output=$(cat "$stdout_tmp" "$stderr_tmp")
    rm -f "$stdout_tmp" "$stderr_tmp"
    echo "$output" > "$assert_file"

    if ! echo "$output" | grep -q "Hello World"; then
        echo "  FAIL $lang: assert.txt missing 'Hello World'" >&2
        echo "$output" >&2
        return 1
    fi
}

if [[ $# -eq 0 ]]; then
    echo "Usage: $0 <language> | --all | --reset" >&2
    exit 1
fi

if [[ "$1" == "--reset" ]]; then
    rm -f "$PROGRESS_FILE"
    echo "Progress reset."
    exit 0
fi

if [[ "$1" == "--all" ]]; then
    touch "$PROGRESS_FILE"
    skipped=0
    passed=0

    while read -r lang; do
        if grep -qx "$lang" "$PROGRESS_FILE" 2>/dev/null; then
            skipped=$((skipped + 1))
            continue
        fi

        if [[ $skipped -gt 0 ]]; then
            echo "  (skipped $skipped already-passed)"
            skipped=0
        fi

        echo "--- $lang"
        if generate_lang "$lang"; then
            echo "$lang" >> "$PROGRESS_FILE"
            passed=$((passed + 1))
            echo "  ✓ $lang"
        else
            echo ""
            echo "To retry after fixing: $0 --all"
            exit 1
        fi
    done < <(grep "^  - name:" "$VEINS_YML" | sed 's/^  - name: //')

    if [[ $skipped -gt 0 ]]; then
        echo "  (skipped $skipped already-passed)"
    fi

    total=$(wc -l < "$PROGRESS_FILE" | tr -d ' ')
    echo ""
    echo "Done. $passed newly passed, $total total passed."
else
    generate_lang "$1"
fi
