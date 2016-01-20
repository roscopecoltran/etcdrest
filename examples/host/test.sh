#/bin/bash

set -eu

unset CLICOLOR
unset GREP_OPTIONS

#TMPFILE="/tmp/test.json"
APIVERS="v1"
URL="http://localhost:8080"

cpt() {
    printf "\n########## $1 ##########\n\n"
}

fatal() {
    echo "$1" >&2
    exit 1
}

which curl &>/dev/null || fatal "Missing binary: curl"

cpt "Create host 1"
curl -s -i -H "Content-Type: application/json" -X PUT -d "$(cat test1.example.com.json)" "${URL}/${APIVERS}/hosts/test1.example.com"

cpt "Create host 2"
curl -s -i -H "Content-Type: application/json" -X PUT -d "$(cat test2.example.com.json)" "${URL}/${APIVERS}/hosts/test2.example.com"

cpt "Get hosts"
curl -s -i -H "Content-Type: application/json" "${URL}/${APIVERS}/hosts"

cpt "Delete host 1"
curl -s -i -H "Content-Type: application/json" -X DELETE "${URL}/${APIVERS}/hosts/test1.example.com"

cpt "Delete host 2"
curl -s -i -H "Content-Type: application/json" -X DELETE "${URL}/${APIVERS}/hosts/test2.example.com"

cpt "Get hosts"
curl -s -i -H "Content-Type: application/json" "${URL}/${APIVERS}/hosts"
