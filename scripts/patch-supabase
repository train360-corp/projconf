#!/bin/bash

PSQL_URL="postgresql://supabase_admin:postgres@127.0.0.1:54322/postgres"

psql "$PSQL_URL" -c "ALTER SYSTEM SET projconf.x_admin_api_key = 'LOCAL-DEVELOPMENT-TEST';"
psql "$PSQL_URL" -c "SELECT pg_reload_conf();"