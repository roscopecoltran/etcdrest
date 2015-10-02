# etcd-rest

Create REST API in Go using etcd as a backend and using JSON schema for validation.

## Usage

```bash
Usage of bin/etcd-rest:
  -bind string
    	Bind to address and port, can be set with env. variable ETCD_REST_BIND (default "127.0.0.1:8080")
  -peers string
    	Comma separated list of etcd nodes, can be set with env. variable ETCD_PEERS (default "http://192.168.99.100:5001")
  -version
    	Version
```

## Example

Start daemon.

```bash
./init-etcd.sh start
eval $(./init-etcd.sh env)
etcdctl mkdir /schemas
etcdctl set /schemas/ntp "$(cat examples/ntp/ntp_schema.json)"
etcd-import -dir /routes -input examples/ntp/routes.json -no-validate
bin/etcd-rest
```

Create / Get / Delete entry.

```bash
curl -v -H "Content-Type: application/json" -X PUT -d "$(cat examples/ntp/ntp-site1.json)" 127.0.0.1:8080/ntp/site1
curl -v -H "Content-Type: application/json" -X PUT -d "$(cat examples/ntp/ntp-site2.json)" 127.0.0.1:8080/ntp/site2
curl -v -H "Content-Type: application/json" 127.0.0.1:8080/ntp
curl -v -H "Content-Type: application/json" 127.0.0.1:8080/ntp/site1
curl -v -H "Content-Type: application/json" -X DELETE 127.0.0.1:8080/ntp/site1
```
