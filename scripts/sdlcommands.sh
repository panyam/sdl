
function sdlurl() {
  path=$1
  echo "http://localhost:8080/$path"
}

function sdlload() {
  endoint=`sdlurl /api/console/load`
  curl $endpoint
}
