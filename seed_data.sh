#!/bin/bash

# Add test data in the askgod instance. Make sure to run `make linux` before running this file, and to run it from the project root directory.

TEAM_NAME_PREFIX="seed-testteam-"
for i in {1..80}
do
    TEAM_NAME="${TEAM_NAME_PREFIX}${i}"
    echo "Adding team $TEAM_NAME"
    ./bin/linux/askgod --server http://localhost:9080 admin add-team name=$TEAM_NAME subnets=0.0.0.0/0 country=CA
done

FLAG_PREFIX="FLAG-SEED-"
echo "Adding 20 flags $FLAG_PREFIX{NUMBER}"
for i in {1..20}
do
./bin/linux/askgod --server http://localhost:9080 admin add-flag flag=$FLAG_PREFIX$i value=$i description="Seed flag #$i" return_string="Seed flag #$i"
done

echo "Adding 100 scores"
SOURCES=("cli" "cli+agent" "mcp" "web" "web+agent")
for i in {1..100}
do
    TEAM_ID=$(( ( RANDOM % 80 )  + 1 ))
    FLAG_ID=$(( ( RANDOM % 20 )  + 1 ))
    SOURCE=${SOURCES[$(( RANDOM % ${#SOURCES[@]} ))]}
    echo "Adding score for team_id=$TEAM_ID flag_id=$FLAG_ID source=$SOURCE value=$FLAG_ID"
    ./bin/linux/askgod --server http://localhost:9080 admin add-score team_id=$TEAM_ID flag_id=$FLAG_ID source=$SOURCE value=$FLAG_ID
done
