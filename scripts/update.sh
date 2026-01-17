#!/bin/bash

# Cron job that checks for updates and rebuilds if it finds any.
# Takes the compose profile as argument (prod or dev).

set -euo pipefail

cd "$(dirname "$0")/.."

if [[ "$1" == "dev" ]]; then
	echo "Environment: dev (tracking main)"

	git fetch origin main

	LOCAL=$(git rev-parse @)
	REMOTE=$(git rev-parse origin/main)

	if [[ "$LOCAL" == "$REMOTE" ]]; then
		echo "No changes found."
		exit 0
	fi

	echo "Pulling updates..."
	git checkout main
	git pull origin main
else
	echo "Environment: prod (tracking latest tag)"

	git fetch --tags origin

	LATEST_TAG=$(git tag --sort=-v:refname | head -n 1)

	if [[ -z "$LATEST_TAG" ]]; then
		echo "No tags found."
		exit 1
	fi

	LOCAL=$(git rev-parse @)
	TARGET=$(git rev-parse "$LATEST_TAG")

	echo "Latest tag: $LATEST_TAG"

	if [[ "$LOCAL" == "$TARGET" ]]; then
		echo "No changes found".
		exit 0
	fi

	echo "Changes detected, checking out $LATEST_TAG"
	git checkout "$LATEST_TAG"
fi


echo "Rebuilding containers..."
make PROFILE=$1 DETACHED=-d PULL=--pull

echo "Cleaning up old images..."
docker image prune -f

echo "Update completed: $(date)"
