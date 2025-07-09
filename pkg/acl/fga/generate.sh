#!/bin/sh
# Dynamically generate fga.mod from available .fga files

TARGET_JSON_FILE="fga.json"
TARGET_FGA_FILE="fga.fga"
TARGET_MD5_FILE="fga.md5"

echo "schema: '1.2'" > fga.mod
echo "contents:" >> fga.mod
for file in *.fga; do
  if [ "$file" = "$TARGET_FGA_FILE" ]; then
    continue
  fi
  if [ -f "$file" ]; then
    echo "  - $file" >> fga.mod
  fi
done


fga model transform --file fga.mod --input-format modular --output-format json | jq --sort-keys '.' > $TARGET_JSON_FILE
fga model transform --file fga.mod --input-format modular --output-format fga > $TARGET_FGA_FILE
md5sum $TARGET_JSON_FILE | cut -d' ' -f1 > $TARGET_MD5_FILE

rm fga.mod
