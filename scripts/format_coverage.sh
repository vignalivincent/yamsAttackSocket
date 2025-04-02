#!/bin/bash
prev_file=""
while read line; do
  file=
  base_file=$file
  func=
  coverage=
  if [ "$base_file" != "$prev_file" ]; then
    if [ -n "$prev_file" ]; then echo ""; fi
    echo "---- $base_file ----"
    prev_file=$base_file
  fi
  echo "- $file :: $func: $coverage"
done
