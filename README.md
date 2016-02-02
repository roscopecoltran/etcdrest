# etcdrest

Create REST API in Go using etcd as a backend with JSON schema for validation.

# Build

```bash
git clone https://github.com/mickep76/etcdrest.git
cd etcdrest
make
```

# CAVEATS

- POST is not supported since we're not using unique ID's but rather each operation is idempotent

# ROADMAP

- JQ style filtering
- In-line JS pre/post hooks for business logic
- Indexes
