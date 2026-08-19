docs_dir := "docs/zpecs"
skills_dir := "skills"

# List available recipes
default:
    @just --list

# Build the zpecs CLI binary
build:
    go build -o zpecs ./cmd/zpecs

# Install the zpecs CLI with go install
install:
    go install ./cmd/zpecs

# Check that every relative markdown link in the docs and skills resolves
check:
    #!/usr/bin/env bash
    set -euo pipefail
    status=0
    for file in {{docs_dir}}/*.md {{skills_dir}}/*/SKILL.md; do
        while read -r link; do
            case "$link" in
                http*|"#"*|"") continue ;;
            esac
            path="${link%%#*}"
            [ -z "$path" ] && continue
            case "$file" in
                {{docs_dir}}/*) resolved="{{docs_dir}}/$path" ;;
                *) resolved="$path" ;;
            esac
            if [ ! -f "$resolved" ]; then
                echo "$file: broken link: $link" >&2
                status=1
            fi
        done < <(grep -o ']([^)]*)' "$file" | sed 's/^](//;s/)$//')
    done
    if [ "$status" -eq 0 ]; then
        echo "All links resolve."
    fi
    exit $status
