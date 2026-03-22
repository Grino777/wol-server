#!/usr/bin/env bash

set -eEuo pipefail

# Create a new migration file for the database

ROOT="$(cd "$(dirname "$0")/.." &>/dev/null; pwd -P)"

if [[ -z "$1" ]]; then
    echo "Please provide a name for this migration."
    exit 1
fi

command -v migrate >/dev/null 2>&1 || {
    echo >&2 "Migrate command not found. Have you installed golang-migrate?";
    echo >&2 "https://github.com/golang-migrate/migrate/blob/master/cmd/migrate/README.md#installation";
    exit 1; 
}
migrate create -ext sql -dir $ROOT/migrations -seq $1

# Template for the newly created migration file
TEMPLATE='BEGIN;

-- Author the migration here. 

END;'

for m in $(find $ROOT/migrations -maxdepth 1 -not -type d -name "*.sql" -print | sort -n | tail -n 2); do
    echo "$TEMPLATE" >> "$m";
done
