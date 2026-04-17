#!/bin/bash
# generate-configs.sh — generate evaluate/env/ and config.yaml.local from templates.
# Usage: ./generate-configs.sh [--force]
#   --force    Overwrite files that already exist.
# All variables default to evaluate/templates/defaults.env; override by exporting env vars.
#
# Database adapter selection:
#   Set DB_ADAPTER=mysql (default) or DB_ADAPTER=postgres before running.
#   The script picks the matching database-secret.<adapter>.json.template.

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EVALUATE_DIR="$(dirname "$SCRIPT_DIR")"
TEMPLATES_DIR="$EVALUATE_DIR/templates"
DEFAULTS_FILE="$TEMPLATES_DIR/defaults.env"
OUTPUT_ENV_DIR="${OUTPUT_ENV_DIR:-$EVALUATE_DIR/env}"
OUTPUT_CONFIG_FILE="${OUTPUT_CONFIG_FILE:-$EVALUATE_DIR/config.yaml.local}"

FORCE=false
while [[ $# -gt 0 ]]; do
    case $1 in
        --force) FORCE=true; shift ;;
        *) echo -e "${RED}[ERR]${NC} Unknown option: $1"; exit 1 ;;
    esac
done

if ! command -v envsubst &>/dev/null; then
    echo -e "${RED}[ERR]${NC} envsubst not found. Install via: brew install gettext"
    exit 1
fi

if [ ! -f "$DEFAULTS_FILE" ]; then
    echo -e "${RED}[ERR]${NC} Defaults file not found: $DEFAULTS_FILE"
    exit 1
fi

# Load defaults — external env values take precedence. We parse the file
# line-by-line instead of `source`-ing it because plain `source` would
# overwrite variables that the caller already exported (e.g. DB_ADAPTER).
while IFS= read -r line || [ -n "$line" ]; do
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
    key="${line%%=*}"
    val="${line#*=}"
    key="${key## }"; key="${key%% }"
    val="${val%$'\r'}"
    [[ -z "$key" ]] && continue
    if [ -z "${!key+x}" ]; then
        export "$key=$val"
    fi
done < "$DEFAULTS_FILE"

DB_ADAPTER="${DB_ADAPTER:-mysql}"
DB_ADAPTER=$(echo "$DB_ADAPTER" | tr '[:upper:]' '[:lower:]')

case "$DB_ADAPTER" in
    mysql|postgres) ;;
    *)
        echo -e "${RED}[ERR]${NC} Unsupported DB_ADAPTER=$DB_ADAPTER (expected 'mysql' or 'postgres')"
        exit 1
        ;;
esac

DB_SECRET_TEMPLATE="$TEMPLATES_DIR/database-secret.${DB_ADAPTER}.json.template"
if [ ! -f "$DB_SECRET_TEMPLATE" ]; then
    echo -e "${RED}[ERR]${NC} Database secret template not found: $DB_SECRET_TEMPLATE"
    exit 1
fi

echo -e "${BLUE}[INFO]${NC} Using database adapter: ${GREEN}${DB_ADAPTER}${NC} (template: $(basename "$DB_SECRET_TEMPLATE"))"

mkdir -p "$OUTPUT_ENV_DIR/database-${DB_NAME_YAML}"
mkdir -p "$OUTPUT_ENV_DIR/redis-${REDIS_NAME_YAML}"
mkdir -p "$OUTPUT_ENV_DIR/cassandra-${CASSANDRA_NAME_YAML}"

_generate() {
    local tmpl="$1"
    local output="$2"
    if [ -f "$output" ] && [ "$FORCE" = false ]; then
        echo -e "${YELLOW}[SKIP]${NC} $output already exists"
        return
    fi
    envsubst < "$tmpl" > "$output"
    echo -e "${GREEN}[OK]${NC} Generated $output"
}

_generate "$DB_SECRET_TEMPLATE" \
          "$OUTPUT_ENV_DIR/database-${DB_NAME_YAML}/secret.json"
_generate "$TEMPLATES_DIR/redis-secret.json.template" \
          "$OUTPUT_ENV_DIR/redis-${REDIS_NAME_YAML}/secret.json"
_generate "$TEMPLATES_DIR/cassandra-secret.json.template" \
          "$OUTPUT_ENV_DIR/cassandra-${CASSANDRA_NAME_YAML}/secret.json"
_generate "$TEMPLATES_DIR/config.yaml.template" \
          "$OUTPUT_CONFIG_FILE"
