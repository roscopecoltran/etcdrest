# etcdrest

Create REST API in Go using etcd as a backend with JSON schema for validation.

# Build

```bash
git clone https://github.com/mickep76/etcdrest.git
cd etcdrest
make
```

# Build Docker image

# Example

Make sure you have Docker configured.

**Start etcd:**

```
./init-etcd.sh start
eval "$(./init-etcd.sh env)"
```

# CAVEATS

- POST is not supported since we're not using unique ID's but rather each operation is idempotent

# ROADMAP

- JQ style filtering
- In-line JS pre/post hooks for business logic
- Template endpoint either Go Text Template or Pongo2 / alt. handlebars?
- Static endpoint
