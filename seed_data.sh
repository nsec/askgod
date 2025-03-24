#!/bin/bash

# Add test data in the askgod instance. Make sure to run `make linux` before running this file, and to run it from the project root directory.

TEAM_NAME="seed-testteam"
echo "Adding team $TEAM_NAME"
./bin/linux/askgod --server http://localhost:9080 admin add-team name=$TEAM_NAME subnets=127.0.0.1/8 country=CA

FLAG_PREFIX="FLAG-SEED-"
echo "Adding 20 flags $FLAG_PREFIX{NUMBER}"
for i in {1..20}
do
./bin/linux/askgod --server http://localhost:9080 admin add-flag flag=$FLAG_PREFIX$i value=$i description="Seed flag #$i" return_string="Seed flag #$i"
done