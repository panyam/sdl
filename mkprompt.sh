#!/bin/sh

for f in  `find . | grep -v apiclient | grep -v pnpm | grep -v llmprompts | grep -v node.mod | grep -v .git | grep -v sdl | grep -v bot | grep -v gen | grep -v web.static | grep -v output | grep -v content | grep -v dist | grep -v web.compo | grep -v web.templ | grep -e "\.go" -e "\.ts" -e "\.tsx" -e "\.html" -e INSTRUCTIONS -e SUMMARY`
do
  echo "FILE $f:"
  echo '```'
  cat $f
  echo '```'
done
