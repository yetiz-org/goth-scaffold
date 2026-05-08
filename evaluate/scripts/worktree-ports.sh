#!/bin/bash
# worktree-ports.sh — allocate and persist an isolated host-port block.

set -euo pipefail

RUN_DIR="$1"
WORKTREE_ID="$2"
REQUESTED_BASE="$3"
PORT_FILE="$RUN_DIR/ports.env"
WORKTREE_RUN_ROOT="$(dirname "$RUN_DIR")"

if [ -f "$PORT_FILE" ]; then
    tr '\n' ' ' < "$PORT_FILE"
    exit 0
fi

mkdir -p "$RUN_DIR"

_is_listening() {
    local port="$1"
    nc -z 127.0.0.1 "$port" >/dev/null 2>&1
}

_is_reserved_by_worktree() {
    local port="$1"
    local file

    [ -d "$WORKTREE_RUN_ROOT" ] || return 1

    while IFS= read -r file; do
        [ "$file" = "$PORT_FILE" ] && continue
        grep -Eq "^[A-Z_]+_PORT=${port}$" "$file" && return 0
    done < <(find "$WORKTREE_RUN_ROOT" -mindepth 2 -maxdepth 2 -name ports.env -type f 2>/dev/null)

    return 1
}

_block_is_available() {
    local base="$1"
    local offset port

    for offset in 0 1 2 3 4 5; do
        port=$((base + offset))
        if _is_listening "$port" || _is_reserved_by_worktree "$port"; then
            return 1
        fi
    done

    return 0
}

base="$REQUESTED_BASE"
min_port=20000
max_base=60994

if ! [[ "$base" =~ ^[0-9]+$ ]]; then
    echo "Invalid WORKTREE_PORT_BASE=$base" >&2
    exit 1
fi

if [ "$base" -lt "$min_port" ] || [ "$base" -gt "$max_base" ]; then
    base=$((min_port + ($(printf '%s' "$WORKTREE_ID" | cksum | awk '{print $1}') % (max_base - min_port + 1))))
fi

for attempt in $(seq 0 $((max_base - min_port))); do
    candidate=$((base + attempt))
    if [ "$candidate" -gt "$max_base" ]; then
        candidate=$((min_port + candidate - max_base - 1))
    fi

    if _block_is_available "$candidate"; then
        {
            echo "MYSQL_PORT=$candidate"
            echo "POSTGRES_PORT=$((candidate + 1))"
            echo "CASSANDRA_PORT=$((candidate + 2))"
            echo "REDIS_PORT=$((candidate + 3))"
            echo "ASYNQMON_PORT=$((candidate + 4))"
            echo "APP_PORT=$((candidate + 5))"
        } > "$PORT_FILE"
        tr '\n' ' ' < "$PORT_FILE"
        exit 0
    fi
done

echo "No available six-port block found for WORKTREE_ID=$WORKTREE_ID" >&2
exit 1
