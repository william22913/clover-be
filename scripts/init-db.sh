#!/bin/sh
# Runs once on a fresh database, as the superuser the postgres image boots with.
# Passwords come from the environment (see .env.example); no secret lives in git.
set -eu

: "${CLOVER_MIGRATOR_PASSWORD:?CLOVER_MIGRATOR_PASSWORD is required}"
: "${CLOVER_APP_PASSWORD:?CLOVER_APP_PASSWORD is required}"

psql -v ON_ERROR_STOP=1 \
	--username "$POSTGRES_USER" \
	--dbname "$POSTGRES_DB" \
	-v migrator_pw="$CLOVER_MIGRATOR_PASSWORD" \
	-v app_pw="$CLOVER_APP_PASSWORD" \
	-f /opt/clover/init-db.sql
