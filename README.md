# etcd-rest

Create REST API in Go using etcd as a backend and using JSON schema for validation.

## Example

```bash
./init-etcd.sh start
eval $(./init-etcd.sh env)
etcdctl mkdir /schemas
etcdctl set /schemas/ntp "$(cat examples/ntp/ntp_schema.json)"
curl -v -H "Content-Type: application/json" -X PUT -d "$(cat examples/ntp/ntp-site1.json)" 127.0.0.1:8080/ntp/site1
curl -v -H "Content-Type: application/json" 127.0.0.1:8080/ntp
curl -v -H "Content-Type: application/json" 127.0.0.1:8080/ntp/site1
curl -v -H "Content-Type: application/json" -X DELETE 127.0.0.1:8080/ntp/site1
```
