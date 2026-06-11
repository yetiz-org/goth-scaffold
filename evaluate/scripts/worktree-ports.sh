#!/bin/bash
# worktree-ports.sh — allocate and persist an isolated host-port block,
# plus host-daemon compose helpers (compose|host-path|host-remote subcommands).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVALUATE_DIR="$(dirname "$SCRIPT_DIR")"

# pick docker-compose or the docker compose plugin; abort without docker
_resolve_compose() {
    if [ -n "${COMPOSE:-}" ]; then
        printf '%s' "$COMPOSE"
        return
    fi

    if ! command -v docker >/dev/null 2>&1; then
        echo "[ERR] docker not found; install Docker first" >&2
        return 1
    fi

    if command -v docker-compose >/dev/null 2>&1 && docker-compose version >/dev/null 2>&1; then
        printf 'docker-compose'
    elif docker compose version >/dev/null 2>&1; then
        printf 'docker compose'
    else
        echo "[ERR] neither docker-compose nor 'docker compose' is usable" >&2
        return 1
    fi
}

_host_env_self_id() {
    hostname 2>/dev/null || cat /etc/hostname 2>/dev/null
}

_host_env_mounts() {
    if [ -z "${_HOST_ENV_MOUNTS_CACHED:-}" ]; then
        _HOST_ENV_MOUNTS="$(docker inspect "$(_host_env_self_id)" \
            --format '{{range .Mounts}}{{.Destination}}|{{.Source}}{{"\n"}}{{end}}' 2>/dev/null || true)"
        _HOST_ENV_MOUNTS_CACHED=1
    fi
    printf "%s" "$_HOST_ENV_MOUNTS"
}

# map a container path to the daemon-visible host path (identity if not mapped)
host_path() {
    local p="$1" best_dst="" best_src="" dst src
    while IFS='|' read -r dst src; do
        [ -n "$dst" ] || continue
        case "$p" in
            "$dst"|"$dst"/*)
                if [ "${#dst}" -gt "${#best_dst}" ]; then
                    best_dst="$dst"
                    best_src="$src"
                fi
                ;;
        esac
    done <<EOF
$(_host_env_mounts)
EOF
    if [ -n "$best_dst" ]; then
        printf "%s\n" "$best_src${p#"$best_dst"}"
    else
        printf "%s\n" "$p"
    fi
}

# true when running in a container that talks to the host's docker daemon
host_env_remote() {
    local probe="${1:-$PWD}"
    case "$probe" in
        /*) ;;
        *) probe="$(cd "$probe" 2>/dev/null && pwd || printf '%s' "$probe")" ;;
    esac
    [ "$(host_path "$probe")" != "$probe" ]
}

# join a compose network so services resolve by container name
ensure_self_on_network() {
    local net="$1"
    host_env_remote || return 0
    docker network connect "$net" "$(_host_env_self_id)" 2>/dev/null || true
}

# detach before `compose down` — network removal fails on active endpoints
disconnect_self_from_network() {
    local net="$1"
    host_env_remote || return 0
    docker network disconnect "$net" "$(_host_env_self_id)" 2>/dev/null || true
}

# compose wrapper for the scoped stack (env: RUN_DIR, CONTAINER_PREFIX), host-daemon aware
compose_cmd() {
    local arg has_up=0 has_down=0
    for arg in "$@"; do
        case "$arg" in
            up) has_up=1 ;;
            down) has_down=1 ;;
        esac
    done
    local run_dir_abs="$EVALUATE_DIR/${RUN_DIR#./}"
    local net="${CONTAINER_PREFIX}-network"
    local compose
    compose="$(_resolve_compose)"
    if [ "$has_up" = 1 ]; then
        mkdir -p "$run_dir_abs/mysql" "$run_dir_abs/postgres" "$run_dir_abs/cassandra" "$run_dir_abs/redis"
    fi
    if [ "$has_down" = 1 ]; then
        disconnect_self_from_network "$net"
    fi
    if host_env_remote "$EVALUATE_DIR"; then
        # "127.0.0.1:" + ":<port>" template = random loopback publish, avoids host port clashes
        export MYSQL_PORT="127.0.0.1:" POSTGRES_PORT="127.0.0.1:" CASSANDRA_PORT="127.0.0.1:" REDIS_PORT="127.0.0.1:" ASYNQMON_PORT="127.0.0.1:"
    fi
    RUN_DIR="$(host_path "$run_dir_abs")" $compose "$@"
    if [ "$has_up" = 1 ]; then
        ensure_self_on_network "$net"
    fi
}

case "${1:-}" in
    compose) shift; compose_cmd "$@"; exit $? ;;
    host-path) host_path "${2:?usage: $0 host-path PATH}"; exit $? ;;
    host-remote) host_env_remote "${2:-$PWD}"; exit $? ;;
esac

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
