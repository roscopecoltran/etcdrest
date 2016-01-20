#/bin/bash

set -eu

unset CLICOLOR
unset GREP_OPTIONS

TMPFILE="/tmp/test.json"
APIVERS="v1"
URL="http://localhost:8080"

if [ -n "${DOCKER_HOST:-}" ]; then
  DOCKER_IP_PORT=${DOCKER_HOST#tcp://}
  DOCKER_IP=${DOCKER_IP_PORT%:*}
  URL="http://${DOCKER_IP}:8080"
fi

create() {
    local endp="$1" resp

    echo "POST: ${URL}/${APIVERS}/${endp}?envelope=true" >&2
    echo "PAYLOAD:" >&2
    echo "$(cat $TMPFILE)" >&2

    resp=$(curl -s -H "Content-Type: application/json" -X POST -d "$(cat $TMPFILE)" "${URL}/${APIVERS}/${endp}?envelope=true")

    echo "RESPONSE:" >&2
    echo "$resp" >&2

    code=$(echo $resp | jq -r .code)
    if [ $code -ne 201 ]; then
        echo $resp
        exit $code
    fi

    echo $resp | jq -r .data.id
}

get() {
    local endp="$1" id="$2"

    echo "GET: ${URL}/${APIVERS}/${endp}/${id}?envelope=true"
    echo "DATA:"
    curl -s -H "Content-Type: application/json" "${URL}/${APIVERS}/${endp}/${id}?envelope=true"
    echo
}

cpt() {
    printf "\n########## $1 ##########\n\n"
}

fatal() {
    local msg="$1"

    echo "$msg" >&2
    exit 1
}

# Check for pre. requisites
which jq &>/dev/null || fatal "Missing pre. requisite: jq"
which curl &>/dev/null || fatal "Missing pre. requisite: curl"

#
# Tenant
#
cat << EOF > $TMPFILE
{
  "tenant": "test1"
}
EOF

cpt "Create Tenant"
tenant_id=$(create "tenants")
cpt "Get Tenant"
get "tenants" $tenant_id

#
# Site
#
cat << EOF > $TMPFILE
{
  "site": "test1",
  "domain": "example.com",
  "dns": [ "192.168.0.252", "192.168.0.253" ],
  "dockerRegistry": "registry.example.com",
  "artifactRepository": "repository.example.com",
  "namingScheme": "serial-number",
  "pxeTheme": "night"
}
EOF

cpt "Create Site"
site_id=$(create "sites")
cpt "Get Site"
get "sites" $site_id

#
# Subnet
#
cat << EOF > $TMPFILE
{
  "subnet": "192.168.0.0/24",
  "mask": "255.255.255.0",
  "gw": "192.168.0.254",
  "siteId": "${site_id}"
}
EOF

cpt "Create Subnet"
subnet_id=$(create "subnets")
cpt "Get Subnet"
get "subnets" $subnet_id

#
# Boot Image
#
cat << EOF > $TMPFILE
{
  "image": "test1",
  "type": "boot",
  "kOpts": "none"
}
EOF

cpt "Create Boot Image"
boot_image_id=$(create "images")
cpt "Get Boot Image"
get "images" $boot_image_id

#
# Boot Tag
#
cat << EOF > $TMPFILE
{
  "tag": "latest",
  "created": "2006-01-02T15:04:05Z",
  "imageId": "${boot_image_id}",
  "sha256": "67f28e21e04a1570781a63a247fce789352beae2889f1d720b2efbec50ef8e0d"
}
EOF

cpt "Create Boot Tag"
boot_tag_id=$(create "tags")
cpt "Get Boot Tag"
get "tags" $boot_tag_id

#
# Image 
#
cat << EOF > $TMPFILE
{
  "image": "test2",
  "type": "docker",
  "bootTagId": "${boot_tag_id}"
}
EOF

cpt "Create Image"
image_id=$(create "images")
cpt "Get Image"
get "images" $image_id

#
# Tag
#
cat << EOF > $TMPFILE
{
  "tag": "latest",
  "created": "2016-01-02T15:04:05Z",
  "imageId": "${image_id}",
  "sha256": "67f28e21e04a1570781a63a247fce789352beae2889f1d720b2efbec50ef8e0d"
}
EOF

cpt "Create Tag"
tag1_id=$(create "tags")
cpt "Get Tag"
get "tags" $tag1_id

#
# Tag 2
#
cat << EOF > $TMPFILE
{
  "tag": "untested",
  "created": "2015-12-01T13:01:05Z",
  "imageId": "${image_id}",
  "sha256": "37ff8e2ae04a1570781a63a247fce789352beae2889f1d720b2efbec50ef8e0d"
}
EOF

cpt "Create Tag"
tag2_id=$(create "tags")
cpt "Get Tag"
get "tags" $tag2_id

#
# Host
#
cat << EOF > $TMPFILE
{
  "hostname": "test1.example.com",
  "build": true,
  "debug": false,
  "gpt": false,
  "tagId": "${tag1_id}",
  "tenantId": "${tenant_id}",
  "labels": [
    "web-server"
  ],
  "siteId": "${site_id}"
}
EOF

cpt "Create Host"
host_id=$(create "hosts" "$(cat $TMPFILE)")
cpt "Get Host"
get "hosts" $host_id

#
# Interface 1
#
cat << EOF > $TMPFILE
{
  "interface": "eth0",
  "dhcp": false,
  "ipv4": "192.168.0.1",
  "hwAddr": "a1:4c:6f:31:6c:d2",
  "subnetId": "${subnet_id}",
  "hostId": "${host_id}"
}
EOF

cpt "Create Interface 1"
interface1_id=$(create "interfaces" "$(cat $TMPFILE)")
cpt "Get Interface 1"
get "interfaces" $interface1_id

#
# Interface 2
#
cat << EOF > $TMPFILE
{
  "interface": "eth1",
  "dhcp": true,
  "hwAddr": "b1:4c:6f:31:6c:d2",
  "hostId": "${host_id}"
}
EOF

cpt "Create Interface 2"
interface2_id=$(create "interfaces" "$(cat $TMPFILE)")
cpt "Get Interface 2"
get "interfaces" $interface2_id

rm -f $TMPFILE
