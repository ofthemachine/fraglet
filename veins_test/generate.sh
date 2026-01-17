#!/bin/bash
# generate.sh - Generate veins_test from 100hellos sources
#
# Usage:
#   ./generate.sh elixir          # Generate test for elixir
#   ./generate.sh --all            # Generate all available
#   ./generate.sh --sync           # Update existing from 100hellos

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VEINS_YML="$REPO_ROOT/pkg/embed/veins.yml"
HELLOS_ROOT="${HELLOS_ROOT:-$HOME/repos/100hellos}"

if [[ ! -f "$VEINS_YML" ]]; then
    echo "Error: veins.yml not found at $VEINS_YML" >&2
    exit 1
fi

if [[ ! -d "$HELLOS_ROOT" ]]; then
    echo "Error: 100hellos directory not found at $HELLOS_ROOT" >&2
    echo "Set 100HELLOS_ROOT environment variable to override" >&2
    exit 1
fi

# Get extension for a vein name from veins.yml
# Prefers script extensions (e.g., .exs over .ex, .py over .pyw)
get_extension() {
    local vein_name="$1"
    # Find the vein entry and extract extensions, prefer script extensions
    awk -v name="$vein_name" '
        /^  - name: / { in_vein = ($3 == name) }
        in_vein && /extensions:/ {
            # Extract extensions from [.ext1, .ext2] format
            match($0, /\[([^\]]+)\]/, arr)
            if (arr[1]) {
                # Split by comma
                n = split(arr[1], exts, ",")
                # Prefer script extensions (exs, py, js, etc.)
                for (i = 1; i <= n; i++) {
                    gsub(/^[[:space:]]*\.|^\./, "", exts[i])
                    ext = "." exts[i]
                    # Prefer script extensions
                    if (ext ~ /\.(exs|py|js|ts|rb|lua|sh|bash)$/) {
                        print ext
                        exit
                    }
                }
                # Fall back to first extension
                gsub(/^[[:space:]]*\.|^\./, "", exts[1])
                print "." exts[1]
                exit
            }
        }
        /^  - name: / && !in_vein { in_vein = 0 }
    ' "$VEINS_YML" | head -1
}

# Extract code from verify.sh (first example after verify_fraglet)
extract_from_verify() {
    local verify_file="$1"
    if [[ ! -f "$verify_file" ]]; then
        return 1
    fi

    # Find first verify_fraglet call and extract the heredoc
    awk '
        /verify_fraglet/ {
            # Skip until we find <<
            while (getline > 0) {
                if (/<</) {
                    # Read until EOF marker
                    while (getline > 0) {
                        if (/^EOF$/) break
                        print
                    }
                    exit
                }
            }
        }
    ' "$verify_file"
}

# Extract first code block from guide.md Examples section
extract_from_guide() {
    local guide_file="$1"
    if [[ ! -f "$guide_file" ]]; then
        return 1
    fi

    # Find Examples section, then first code block (just the first simple example)
    awk '
        /^## Examples/ { in_examples = 1 }
        in_examples && /^```/ {
            # Read code block
            lang = $2
            code_started = 0
            while (getline > 0) {
                if (/^```/) break
                # Stop after first meaningful example (usually 3-5 lines)
                if (code_started && NF == 0 && prev_was_code) {
                    # Empty line after code, likely end of first example
                    break
                }
                if (NF > 0) {
                    code_started = 1
                    prev_was_code = 1
                } else {
                    prev_was_code = 0
                }
                print
            }
            exit
        }
    ' "$guide_file" | head -10
}

# Extract from hello-world file, stripping shebang
extract_from_hello() {
    local hello_file="$1"
    if [[ ! -f "$hello_file" ]]; then
        return 1
    fi

    # Strip shebang and BEGIN/END_FRAGLET markers
    sed -E '
        /^#!/d
        /BEGIN_FRAGLET/d
        /END_FRAGLET/d
    ' "$hello_file"
}

# Find hello-world file for a language
find_hello_world() {
    local lang="$1"
    local lang_dir="$HELLOS_ROOT/$lang"

    if [[ ! -d "$lang_dir" ]]; then
        return 1
    fi

    # Look in files/ directory
    find "$lang_dir/files" -name "hello-world.*" 2>/dev/null | head -1
}

# Generate test files for a language
generate_lang() {
    local lang="$1"
    local lang_dir="$SCRIPT_DIR/$lang"
    local hellos_lang_dir="$HELLOS_ROOT/$lang"

    echo "Generating test for: $lang"

    # Get extension
    local ext=$(get_extension "$lang")
    if [[ -z "$ext" ]]; then
        echo "  Warning: No extension found for vein '$lang', skipping" >&2
        return 1
    fi

    # Create directory
    mkdir -p "$lang_dir"

    # Try to extract code in priority order
    local code=""
    local code_source=""

    # 1. Try verify.sh
    local verify_file="$hellos_lang_dir/fraglet/verify.sh"
    if code=$(extract_from_verify "$verify_file" 2>/dev/null) && [[ -n "$code" ]]; then
        code_source="verify.sh"
    # 2. Try guide.md
    elif code=$(extract_from_guide "$hellos_lang_dir/fraglet/guide.md" 2>/dev/null) && [[ -n "$code" ]]; then
        code_source="guide.md"
    # 3. Try hello-world file
    elif hello_file=$(find_hello_world "$lang") && [[ -n "$hello_file" ]]; then
        code=$(extract_from_hello "$hello_file" 2>/dev/null)
        code_source="hello-world"
    fi

    if [[ -z "$code" ]]; then
        echo "  Warning: No code found for $lang, creating minimal test" >&2
        # Create minimal hello world based on extension
        case "$ext" in
            .py) code='print("Hello, World!")' ;;
            .exs|.ex) code='IO.puts("Hello, World!")' ;;
            .js) code='console.log("Hello, World!");' ;;
            .rb) code='puts "Hello, World!"' ;;
            .lua) code='print("Hello, World!")' ;;
            .c) code='#include <stdio.h>\nint main() { printf("Hello, World!\\n"); return 0; }' ;;
            *) code='echo "Hello, World!"' ;;
        esac
        code_source="minimal"
    fi

    # Determine filename (use first extension)
    local filename="test${ext}"

    # Create shebang script
    local script_file="$lang_dir/$filename"
    {
        echo "#!/usr/bin/env -S fragletc --vein=$lang"
        echo "$code"
    } > "$script_file"
    chmod +x "$script_file"

    echo "  Created: $script_file (from $code_source)"

    # Create act.sh
    local act_file="$lang_dir/act.sh"
    {
        echo "#!/bin/sh"
        echo "set -e"
        echo "chmod +x ./*${ext} 2>/dev/null || true"
        echo "./$filename"
    } > "$act_file"
    chmod +x "$act_file"

    # Generate assert.txt by running the script
    # Note: This requires fragletc to be available
    local assert_file="$lang_dir/assert.txt"
    if command -v fragletc >/dev/null 2>&1; then
        echo "  Running script to generate assert.txt..."
        if output=$("$script_file" 2>&1); then
            echo "$output" > "$assert_file"
            echo "  Created: $assert_file"
        else
            echo "  Warning: Script execution failed, creating empty assert.txt" >&2
            echo "" > "$assert_file"
        fi
    else
        echo "  Warning: fragletc not found, creating empty assert.txt" >&2
        echo "  Run the test manually and update assert.txt" >&2
        echo "" > "$assert_file"
    fi
}

# Main
if [[ $# -eq 0 ]]; then
    echo "Usage: $0 <language> | --all | --sync" >&2
    exit 1
fi

if [[ "$1" == "--all" ]]; then
    # Generate all languages from veins.yml
    grep "^  - name:" "$VEINS_YML" | sed 's/^  - name: //' | while read -r lang; do
        generate_lang "$lang" || true
    done
elif [[ "$1" == "--sync" ]]; then
    # Update existing tests
    for lang_dir in "$SCRIPT_DIR"/*/; do
        if [[ -d "$lang_dir" ]]; then
            lang=$(basename "$lang_dir")
            if [[ "$lang" != "veins_test.go" && "$lang" != "README.md" && "$lang" != "generate.sh" ]]; then
                generate_lang "$lang" || true
            fi
        fi
    done
else
    # Generate specific language
    generate_lang "$1"
fi
