#!/bin/bash
source .env

# если хочешь брать host из переменных окружения, можно так:
export MIGRATION_DSN="host=postgres port=5432 dbname=$PG_DATABASE_NAME user=$PG_USER password=$PG_PASSWORD sslmode=disable"

sleep 5  # ждем, чтобы postgres поднялся

goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v