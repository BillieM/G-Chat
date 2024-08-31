#!/bin/bash

# Set environment variables
export GOOSE_DRIVER="sqlite3"
export GOOSE_DBSTRING="./db/app.db"
export GOOSE_MIGRATION_DIR="./db/migrations"

# Run Goose command
goose "$@"
