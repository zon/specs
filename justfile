docs_dir := "docs/zpecs"
skills_dir := "skills"
version := trim(`cat VERSION`)

# List available recipes
default:
    @just --list

# Build the zpecs CLI binary
build:
    go build -ldflags "-X main.version={{version}}" -o zpecs ./cmd/zpecs

# Install the zpecs CLI with go install
install:
    go install -ldflags "-X main.version={{version}}" ./cmd/zpecs

# Check that every relative markdown link in the docs and skills resolves
check:
    #!/usr/bin/env bash
    set -euo pipefail
    status=0
    for file in {{docs_dir}}/*.md docs/*.md docs/cli/*.md {{skills_dir}}/*/SKILL.md; do
        dir=$(dirname "$file")
        while read -r link; do
            case "$link" in
                http*|"#"*|"") continue ;;
            esac
            path="${link%%#*}"
            [ -z "$path" ] && continue
            if [ ! -f "$dir/$path" ] && [ ! -f "$path" ]; then
                echo "$file: broken link: $link" >&2
                status=1
            fi
        done < <(grep -o ']([^)]*)' "$file" | sed 's/^](//;s/)$//')
    done
    if [ "$status" -eq 0 ]; then
        echo "All links resolve."
    fi
    exit $status
