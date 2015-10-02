# etcd-rest

Create REST API in Go using etcd as a backend with JSON schema for validation.

## Usage

```bash
Usage of bin/etcd-rest:
  -bind string
    	Bind to address and port, can be set with env. variable ETCD_REST_BIND (default "127.0.0.1:8080")
  -peers string
    	Comma separated list of etcd nodes, can be set with env. variable ETCD_PEERS (default "http://127.0.0.1:4001,http://127.0.0.1:2379")
  -version
    	Version
```

## Example

Start daemon, example uses: https://github.com/mickep76/etcd-export

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

# Run using Docker

```bash
docker run --rm -p 8080:8080 -e ETCD_PEERS=http://etcd.example.com:5001 mickep76/etcd-rest:latest
```

# Build

```bash
git clone https://github.com/mickep76/etcd-export.git
cd etcd-export
./build
bin/etcd-rest --version
```

# Build RPM

```bash
sudo yum install -y rpm-build
make rpm
sudo rpm -i etcd-rest-<version>-<release>.rpm
```

# Build Docker image

```bash
make docker-build
```

# Install using Homebrew on Mac OS X

```bash
brew tap mickep76/funk-gnarge
brew install etcd-rest
```
