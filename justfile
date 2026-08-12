docs_dir := "docs/specs"
skills_dir := ".claude/skills"
skill_prefix := "spec-"

# List available recipes
default:
    @just --list

# Copy the spec docs and skills into another repo or directory
install dir:
    #!/usr/bin/env bash
    set -euo pipefail

    target="{{dir}}"

    if [ ! -d "$target" ]; then
        echo "error: target directory does not exist: $target" >&2
        exit 1
    fi

    target="$(cd "$target" && pwd)"
    source="$(pwd)"

    if [ "$target" = "$source" ]; then
        echo "error: target is this repository" >&2
        exit 1
    fi

    # Docs: replace the whole directory so removed docs do not linger.
    rm -rf "$target/{{docs_dir}}"
    mkdir -p "$target/{{docs_dir}}"
    cp "{{docs_dir}}"/*.md "$target/{{docs_dir}}/"
    doc_count=$(ls -1 "{{docs_dir}}"/*.md | wc -l | tr -d ' ')

    # Skills: replace every prefixed skill and prune ones we no longer ship.
    mkdir -p "$target/{{skills_dir}}"
    skill_count=0
    for path in "{{skills_dir}}"/{{skill_prefix}}*/; do
        name="$(basename "$path")"
        rm -rf "$target/{{skills_dir}}/$name"
        cp -R "$path" "$target/{{skills_dir}}/$name"
        skill_count=$((skill_count + 1))
    done

    pruned=0
    for path in "$target/{{skills_dir}}"/{{skill_prefix}}*/; do
        name="$(basename "$path")"
        if [ ! -d "{{skills_dir}}/$name" ]; then
            rm -rf "$path"
            pruned=$((pruned + 1))
        fi
    done

    echo "Installed $doc_count docs to $target/{{docs_dir}}"
    echo "Installed $skill_count skills to $target/{{skills_dir}}"
    if [ "$pruned" -gt 0 ]; then
        echo "Pruned $pruned skill(s) no longer published"
    fi
    echo
    echo "Point the target's AGENTS.md at the standards:"
    echo
    echo "    Before writing any code, read [{{docs_dir}}/code.md]({{docs_dir}}/code.md)."
    echo "    Before writing any tests, read [{{docs_dir}}/testing.md]({{docs_dir}}/testing.md)."

# List what install would copy
list:
    @echo "docs:"
    @ls -1 {{docs_dir}}/*.md | sed 's/^/  /'
    @echo "skills:"
    @ls -1d {{skills_dir}}/{{skill_prefix}}*/ | sed 's/^/  /'

# Check that every relative markdown link in the docs and skills resolves
check:
    #!/usr/bin/env bash
    set -euo pipefail
    status=0
    for file in {{docs_dir}}/*.md {{skills_dir}}/{{skill_prefix}}*/SKILL.md; do
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
