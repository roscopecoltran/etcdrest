# etcd-rest

Create REST API in Go using etcd as a backend and using JSON schema for validation.

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
