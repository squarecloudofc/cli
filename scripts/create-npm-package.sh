#!/bin/bash

SOURCE_DIR="scripts/npm"
TARGET_DIR="dist/npm/squarecloud"
VERSION=$(git describe --tags --abbrev=0 | sed 's/^v//')

SEMVER_REGEX="^[0-9]+\.[0-9]+\.[0-9]+$"
if [[ ! $VERSION =~ $SEMVER_REGEX ]]; then
  echo "Error: Invalid $VERSION version."
  exit 1
fi

mkdir -p "$TARGET_DIR"

cp "${SOURCE_DIR}/constants.js" "${TARGET_DIR}/"
cp "${SOURCE_DIR}/installer.js" "${TARGET_DIR}/"
cp "${SOURCE_DIR}/lib.js" "${TARGET_DIR}/"
cp "${SOURCE_DIR}/runner.js" "${TARGET_DIR}/"
cp "${SOURCE_DIR}/package.json" "${TARGET_DIR}/"
cp "${SOURCE_DIR}/README.md" "${TARGET_DIR}/"
cp "LICENSE" "${TARGET_DIR}/"

sed -i "s/{version}/$VERSION/g" $TARGET_DIR/package.json
