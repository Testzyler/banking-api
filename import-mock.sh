#!/bin/bash

# Load environment variables from .env file
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "Error: .env file not found!"
    exit 1
fi

DB_CONTAINER=banking_db
DB_NAME=$MYSQL_DATABASE
DB_USER=$MYSQL_USER
DB_PASS=$MYSQL_PASSWORD

# Check if required environment variables are set
if [ -z "$DB_NAME" ] || [ -z "$DB_USER" ] || [ -z "$DB_PASS" ]; then
    echo "Error: Missing required environment variables in .env file"
    echo "Required: MYSQL_DATABASE, MYSQL_USER, MYSQL_PASSWORD"
    exit 1
fi

echo "Using database: $DB_NAME with user: $DB_USER"

for file in ./mock/*.sql; do
  echo "Importing $file into $DB_NAME..."
  docker exec -i "$DB_CONTAINER" mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" < "$file"
done
