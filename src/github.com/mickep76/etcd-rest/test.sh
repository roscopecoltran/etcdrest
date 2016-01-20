#/bin/bash

set -eu

unset CLICOLOR
unset GREP_OPTIONS

TMPFILE="/tmp/test.json"
URL="http://localhost:8080"

cpt() {
    printf "\n########## $1 ##########\n\n"
}

fatal() {
    local msg="$1"

    echo "$msg" >&2
    exit 1
}

# Check for binaries
which curl &>/dev/null || fatal "Missing binary: curl"

# Create host
cat << EOF > $TMPFILE
{
  "interfaces": {
    "eth0": {
      "gw": "192.168.0.1",
      "hwaddr": "00:01:02:03:04:05",
      "ip": "192.168.0.2",
      "netmask": "255.255.255.0"
    }
  },
  "site": "emea-nl-1",
  "tenant": "ops"
}
EOF

cpt "Create host"
curl -s -i -H "Content-Type: application/json" -X PUT -d "$(cat $TMPFILE)" "${URL}/hosts/test1.example.com"

cpt "Get host"
curl -s -i -H "Content-Type: application/json" "${URL}/hosts/test1.example.com"

# Update host
cat << EOF > $TMPFILE
{
  "site": "amer-nl-1"
}
EOF

cpt "Update host"
curl -s -i -H "Content-Type: application/json" -X PUT -d "$(cat $TMPFILE)" "${URL}/hosts/test1.example.com"

cpt "Get host"
curl -s -i -H "Content-Type: application/json" "${URL}/hosts/test1.example.com"

cpt "Delete host"
curl -s -i -H "Content-Type: application/json" -X DELETE "${URL}/hosts/test1.example.com"
