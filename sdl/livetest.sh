#!/bin/sh

 while true; do
  clear
  go test ./...
  fswatch  -o ../ | echo "Files changed, re-testing..."
  sleep 1
 done
